package main

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AccessToken string `yaml:"access_token"`
	LocalRoot   string `yaml:"local_root"`
}

func loadConfig(r io.Reader) (*Config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	c := Config{}
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
