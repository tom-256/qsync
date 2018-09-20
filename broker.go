package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
	RenderedBody   string `json:"rendered_body"`
	Body           string `json:"body"`
	Coediting      bool   `json:"coediting"`
	CreatedAt      string `json:"created_at"`
	Id             string `json:"id"`
	CommentsCount  int    `json:"comments_count"`
	Group          string `json:"group"`
	LikesCount     int    `json:"likes_count"`
	Private        bool   `json:"private"`
	ReactionsCount int    `json:"reactions_count"`
	Tags           []tag  `json:"tags"`
	Title          string `json:"title"`
	UpdatedAt      string `json:"updated_at"`
	URL            string `json:"url"`
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
		e, err := convertItemsToEntry(i)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e)
	}

	return entries, err
}

//APIの返り値データをentry型に変換
func convertItemsToEntry(i *item) (*entry, error) {
	createdAt,err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil,err
	}

	//tags to string
	tagsString := []string{}
	for _,t := range i.Tags  {
		tagsString = append(tagsString, t.Name)
	}

	updatedAt,err := time.Parse(time.RFC3339, i.UpdatedAt)
	if err != nil {
		return nil,err
	}

	entry := &entry{
		entryHeader: &entryHeader{
			Title:    i.Title,
			Tags:     tagsString,
			Date:     &createdAt,
			Url:      i.URL,
			Id:       i.Id,
			Private:  i.Private,
		},
		LastModified: &updatedAt,
		Content:      i.Body,
	}

	return entry, nil
}

func (b *broker) LocalPath(e *entry) string {
	extension := ".md"
	paths := []string{b.config.LocalRoot}
	pathFormat := "2006/01/02"
	datePath := e.Date.Format(pathFormat)
	paths = append(paths, datePath)
	idPath := e.Id+extension
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
