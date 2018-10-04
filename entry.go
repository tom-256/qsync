package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
)

type header struct {
	Title   string     `yaml:"Title"`
	Tags    []tag      `yaml:"Tags"`
	Date    *time.Time `yaml:"Date"`
	URL     string     `yaml:"URL"`
	ID      string     `yaml:"ID"`
	Private bool       `yaml:"Private"`
}

type entry struct {
	*header
	LastModified *time.Time
	Content      string
}

func (e *entry) HeaderString() string {
	d, err := yaml.Marshal(e.header)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	headers := []string{
		"---",
		string(d),
	}
	return strings.Join(headers, "\n") + "---\n\n"
}

func (e *entry) fullContent() string {
	c := e.HeaderString() + e.Content
	if !strings.HasSuffix(c, "\n") {
		// fill newline for suppressing diff "No newline at end of file"
		c += "\n"
	}
	return c
}

var delimReg = regexp.MustCompile(`---\n+`)

func entryFromReader(source io.Reader) (*entry, error) {
	b, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	content := string(b)
	isNew := !strings.HasPrefix(content, "---\n")
	eh := header{}
	if !isNew {
		c := delimReg.Split(content, 3)
		if len(c) != 3 || c[0] != "" {
			return nil, fmt.Errorf("entry format is invalid")
		}

		err = yaml.Unmarshal([]byte(c[1]), &eh)
		if err != nil {
			return nil, err
		}
		content = c[2]
	}
	entry := &entry{
		header:  &eh,
		Content: content,
	}

	if f, ok := source.(*os.File); ok {
		fi, err := os.Stat(f.Name())
		if err != nil {
			return nil, err
		}
		t := fi.ModTime()
		entry.LastModified = &t
	}

	return entry, nil
}

func (e *entry) ConvertToItem() *item {
	item := &item{
		Body:      e.Content,
		CreatedAt: e.Date,
		ID:        e.ID,
		Private:   e.Private,
		Tags:      e.Tags,
		Title:     e.Title,
		UpdatedAt: e.LastModified,
		URL:       e.URL,
	}
	return item
}

func (e *entry) ConvertToPostItem() *item {
	item := &item{
		Body:    e.Content,
		Private: e.Private,
		Tags:    e.Tags,
		Title:   e.Title,
	}
	return item
}
