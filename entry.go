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

type entryHeader struct {
	Title   string     `yaml:"Title"`
	Tags    []string   `yaml:"Tags"`
	Date    *time.Time `yaml:"Date"`
	Url     string     `yaml:"Url"`
	Id      string     `yaml:"Id"`
	Private bool       `yaml:"private"`
}

type entry struct {
	*entryHeader
	LastModified *time.Time
	Content      string
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