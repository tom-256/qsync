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
	app.Commands = []cli.Command{
		commandPull,
		commandPush,
		commandPost,
	}

	err := app.Run(os.Args)
	if err != nil {
		if err != errCommandHelp {
			logf("error", "%s", err)
		}
		os.Exit(1)
	}
}

func loadSingleConfigFile(fname string) (*Config, error) {
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

func loadConfiguration() (*Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	conf, err := loadSingleConfigFile(filepath.Join(home, ".config", "qsync", "config.yaml"))
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, fmt.Errorf("no config files found")
	}

	return conf, nil
}

var commandPull = cli.Command{
	Name:  "pull",
	Usage: "pull entries from remote",
	Action: func(c *cli.Context) error {

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		b := newBroker(conf)
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

var commandPush = cli.Command{
	Name:  "push",
	Usage: "Push local entries to remote",
	Action: func(c *cli.Context) error {
		path := c.Args().First()
		if path == "" {
			cli.ShowCommandHelp(c, "push")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		entry, err := entryFromReader(f)
		if err != nil {
			return err
		}

		_, err = newBroker(conf).UploadFresh(entry)
		if err != nil {
			return err
		}
		return nil
	},
}

var commandPost = cli.Command{
	Name:  "post",
	Usage: "Post a new entry to remote",
	Action: func(c *cli.Context) error {
		if c.Args().Present() {
			cli.ShowCommandHelp(c, "post")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		err = newBroker(conf).PostEntry()
		if err != nil {
			return err
		}
		return nil
	},
}
