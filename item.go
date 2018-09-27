package main

import "time"

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

func (i item) ConvertToEntry() *entry{
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

	return entry
}
