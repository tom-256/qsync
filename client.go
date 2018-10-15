package main

import (
	"net/url"
	"log"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"io"
	"path"
	"encoding/json"
	"fmt"
)

type Client struct {
	URL         *url.URL
	HTTPClient  *http.Client
	AccessToken string
	Logger      *log.Logger
}

func NewClient(accessToken string, logger *log.Logger) (*Client, error) {
	if len(accessToken) == 0 {
		return nil, errors.New("missing access token")
	}
	parsedURL, err := url.ParseRequestURI("https://qiita.com/api/v2")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s")
	}
	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}
	client := &http.Client{}
	c := &Client{
		URL:         parsedURL,
		HTTPClient:  client,
		AccessToken: accessToken,
		Logger:      logger,
	}
	return c, nil
}

func (c *Client) newRequest(method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", " Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func (c *Client) GetItems() ([]*item, error) {
	spath := "/authenticated_user/items"
	req, err := c.newRequest("GET", spath, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	var items []*item
	if err := decodeBody(res, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) GetItem(id string) (*item, error) {
	spath := fmt.Sprintf("/items/%s", id)
	req, err := c.newRequest("GET", spath, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	var item *item
	if err := decodeBody(res, &item); err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Client) PatchItem(id string, body io.Reader) (*item, error) {
	spath := fmt.Sprintf("/items/%s", id)
	req, err := c.newRequest("PATCH", spath, body)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	var item *item
	if err := decodeBody(res, &item); err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Client) PostItem(body io.Reader) (*item, error) {
	spath := "/items"
	req, err := c.newRequest("POST", spath, body)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 201 {
		return nil, errors.New(res.Status)
	}

	var item *item
	if err := decodeBody(res, &item); err != nil {
		return nil, err
	}

	return item, nil
}
