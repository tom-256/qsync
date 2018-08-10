package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const timeFormat = "2006-01-02T15:04:05-07:00"

type entryURL struct {
	*url.URL
}

type entryHeader struct {
	Title      string     `yaml:"Title"`
	Category   []string   `yaml:"Category,omitempty"`
	Date       *time.Time `yaml:"Date"`
	URL        *entryURL  `yaml:"URL"`
	EditURL    string     `yaml:"EditURL"`
	IsDraft    bool       `yaml:"Draft,omitempty"`
	CustomPath string     `yaml:"CustomPath,omitempty"`
}

func (eu *entryURL) MarshalYAML() (interface{}, error) {
	return eu.String(), nil
}

func (eu *entryURL) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	eu.URL = u
	return nil
}

func (eh *entryHeader) remoteRoot() (string, error) {
	// EditURL: https://blog.hatena.ne.jp/Songmu/songmu.hateblog.jp/atom/entry/...
	// "songmu.hateblog.jp" is remote root in above case.
	paths := strings.Split(eh.EditURL, "/")
	if len(paths) < 5 {
		return "", fmt.Errorf("failed to get remoteRoot form EditURL: %s", eh.EditURL)
	}
	return paths[4], nil
}

// Entry is an entry stored on remote blog providers
type entry struct {
	*entryHeader
	LastModified *time.Time
	Content      string
	ContentType  string
}

func (e *entry) HeaderString() string {
	d, err := yaml.Marshal(e.entryHeader)
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
	eh := entryHeader{}
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
		entryHeader: &eh,
		Content:     content,
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
