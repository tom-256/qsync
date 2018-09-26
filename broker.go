package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"fmt"
	"bytes"
)

type broker struct {
	config *Config
}

func newBroker(config *Config) *broker {
	return &broker{
		config: config,
	}
}

type item struct {
	Body           string    `json:"body"`
	Coediting      bool      `json:"coediting"`
	CreatedAt      time.Time `json:"created_at"`
	Id             string    `json:"id"`
	Private        bool      `json:"private"`
	Tags           []tag     `json:"tags"`
	Title          string    `json:"title"`
	UpdatedAt      time.Time `json:"updated_at"`
	URL            string    `json:"url"`
}

type tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

func (b *broker) FetchRemoteEntries() ([]*entry, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://qiita.com/api/v2/authenticated_user/items", nil)

	req.Header.Add("Authorization", " Bearer "+b.config.APIKey)
	response, err := client.Do(req)

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)

	var items []*item
	err = decoder.Decode(&items)
	if err != nil {
		return nil, err
	}

	entries := []*entry{}
	for _, i := range items {
		e, err := convertItemToEntry(i)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e)
	}

	return entries, err
}

func (b *broker) FetchRemoteEntry(id string) (*entry, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://qiita.com/api/v2/items/"+id, nil)

	req.Header.Add("Authorization", " Bearer "+b.config.APIKey)
	response, err := client.Do(req)

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)

	var item item
	err = decoder.Decode(&item)
	if err != nil {
		return nil, err
	}

	entry, err := convertItemToEntry(&item)

	return entry, err
}

//APIの返り値データをentry型に変換
func convertItemToEntry(i *item) (*entry, error) {

	entry := &entry{
		header: &header{
			Title:   i.Title,
			Tags:    i.Tags,
			Date:    &i.CreatedAt,
			Url:     i.URL,
			Id:      i.Id,
			Private: i.Private,
		},
		LastModified: &i.UpdatedAt,
		Content:      i.Body,
	}

	return entry, nil
}

func convertEntryToItem(e *entry) (*item, error){
	item := &item{
		Body:e.Content,
		CreatedAt:*e.Date,
		Id:e.Id,
		Private:e.Private,
		Tags:e.Tags,
		Title:e.Title,
		UpdatedAt:*e.LastModified,
		URL:e.Url,
	}
	return item,nil
}

func (b *broker) LocalPath(e *entry) string {
	extension := ".md"
	paths := []string{b.config.LocalRoot}
	pathFormat := "2006/01/02"
	datePath := e.Date.Format(pathFormat)
	paths = append(paths, datePath)
	idPath := e.Id + extension
	paths = append(paths, idPath)

	return filepath.Join(paths...)
}

func (b *broker) StoreFresh(e *entry, path string) (bool, error) {
	var localLastModified time.Time
	if fi, err := os.Stat(path); err == nil {
		localLastModified = fi.ModTime()
	}

	if e.LastModified.After(localLastModified) {
		logf("fresh", "remote=%s > local=%s", e.LastModified, localLastModified)
		if err := b.Store(e, path); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (b *broker) Store(e *entry, path string) error {
	logf("store", "%s", path)

	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(e.fullContent())
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return os.Chtimes(path, *e.LastModified, *e.LastModified)
}

func (b *broker) PutEntry(e *entry) error {
	i,err := convertEntryToItem(e)
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return err
	}

	client := http.Client{}
	req, err := http.NewRequest("PATCH", "https://qiita.com/api/v2/items/"+e.Id, bytes.NewBuffer(jsonBytes))

	req.Header.Add("Authorization", " Bearer "+b.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)

	var responseItem item
	err = decoder.Decode(&responseItem)
	if err != nil {
		return err
	}

	ne, err := convertItemToEntry(&responseItem)
	if err != nil {
		return err
	}

	path := b.LocalPath(ne)
	_, err = b.StoreFresh(ne, path)
	if err != nil {
		return err
	}
	return nil
}


func (b *broker) UploadFresh(e *entry) (bool, error) {
	re, err := b.FetchRemoteEntry(e.Id)
	if err != nil {
		return false, err
	}

	if e.LastModified.After(*re.LastModified) == false {
		return false, nil
	}

	return true, b.PutEntry(e)
}
