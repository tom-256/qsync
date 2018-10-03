package main

import (
	"time"
)

type item struct {
	Body      string     `json:"body"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Id        string     `json:"id,omitempty"`
	Private   bool       `json:"private"`
	Tags      []tag      `json:"tags"`
	Title     string     `json:"title"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	URL       string     `json:"url,omitempty"`
}

type tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

func (i *item) ConvertToEntry() *entry {
	entry := &entry{
		header: &header{
			Title:   i.Title,
			Tags:    i.Tags,
			Date:    i.CreatedAt,
			Url:     i.URL,
			Id:      i.Id,
			Private: i.Private,
		},
		LastModified: i.UpdatedAt,
		Content:      i.Body,
	}

	return entry
}
