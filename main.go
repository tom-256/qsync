package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/mitchellh/go-homedir"
)

var errCommandHelp = fmt.Errorf("command help shown")

func main() {
	app := cli.NewApp()
	app.Commands = []cli.command{
		commandPull,
	}

	token := os.Getenv("QIITA_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("access token is not preset, set QIITA_ACCESS_TOKEN")
		os.Exit(1)
	}

	err := app.Run(os.Args)
	if err != nil {
		if err != errCommandHelp {
			logf("error", "%s", err)
		}
		os.Exit(1)
	}
}

func loadSingleConfigFile(fname string) (*config, error) {
	if _, err := os.Stat(fname); err != nil {
		return nil, nil
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadConfig(f)
}

func loadConfiguration() (*config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return loadConfigFiles(pwd)
}

func loadConfigFiles(pwd string) (*config, error) {
	conf, err := loadSingleConfigFile(filepath.Join(pwd, "blogsync.yaml"))
	if err != nil {
		return nil, err
	}

	home, err := homedir.Dir()
	if err != nil && conf == nil {
		return nil, err
	}
	if err == nil {
		homeConf, err := loadSingleConfigFile(filepath.Join(home, ".config", "blogsync", "config.yaml"))
		if err != nil {
			return nil, err
		}
		conf = mergeConfig(conf, homeConf)
	}
	if conf == nil {
		return nil, fmt.Errorf("no config files found")
	}
	return conf, nil
}

var commandPull = cli.commnand{
	Name:  "pull",
	Usage: "pull entries from remote",
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "pull")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			return fmt.Errorf("blog not found: %s", blog)
		}

		b := newBroker(blogConfig)
		remoteEntries, err := b.FetchRemoteEntries()
		if err != nil {
			return err
		}

		for _, re := range remoteEntries {
			path := b.LocalPath(re)
			_, err := b.StoreFresh(re, path)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
