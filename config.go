package main

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Default *blogConfig
}

type blogConfig struct {
	RemoteRoot string `yaml:"-"`
	LocalRoot  string `yaml:"local_root"`
}

func loadConfig(r io.Reader) (*config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var blogs map[string]*blogConfig
	err = yaml.Unmarshal(bytes, &blogs)
	if err != nil {
		return nil, err
	}

	c := &config{
		Default: blogs["default"],
	}

	delete(blogs, "default")
	for key, b := range blogs {
		if b == nil {
			b = &blogConfig{}
			blogs[key] = b
		}
		b.RemoteRoot = key
	}
	return c, nil
}

func (c *config) Get(remoteRoot string) *blogConfig {
	bc, ok := c.Blogs[remoteRoot]
	if !ok {
		return nil
	}
	return mergeBlogConfig(bc, c.Default)
}

func mergeBlogConfig(b1, b2 *blogConfig) *blogConfig {
	if b1 == nil {
		if b2 != nil {
			return b2
		}
		b1 = &blogConfig{}
	}
	if b2 == nil {
		return b1
	}
	if b1.LocalRoot == "" {
		b1.LocalRoot = b2.LocalRoot
	}
	return b1
}

func mergeConfig(c1, c2 *config) *config {
	if c1 == nil {
		c1 = &config{
			Blogs: make(map[string]*blogConfig),
		}
	}
	if c2 == nil {
		return c1
	}

	c1.Default = mergeBlogConfig(c1.Default, c2.Default)
	for k, bc := range c2.Blogs {
		c1.Blogs[k] = mergeBlogConfig(c1.Blogs[k], bc)
	}
	return c1
}
