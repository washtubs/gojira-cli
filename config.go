package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const defaultConfigContents string = `labels:
  - sample label 1
  - sample label 2
queries:
  "project foo": project = "FOO"
  "project bar": project = "BAR"
  "all":
actions:
  "helloworld": echo {{ .Issue.ID }}
client:
  url: ""
  keyfile: ""
  certfile: ""
  username: ""
  passfile: ""
`

type FavoritesConfig struct {
	Users []string `yaml:"users"`
}

type Config struct {
	// WOuld prefer to use the saved JQL queries for the user, but I
	JQLs    map[string]string `yaml:"queries"`
	Actions map[string]string `yaml:"actions"`
	// Jira doesn't seem to keep a list of existing labels so I gotta add them via config
	LabelsAllowed []Label          `yaml:"labels"`
	Client        JiraClientConfig `yaml:"client"`
	Favorites     *FavoritesConfig `yaml:"favorites"`
}

type JiraClientConfig struct {
	Url       string `yaml:"url"`
	TokenFile string `yaml:"tokenFile"`
	Username  string `yaml:"username"`
	Passfile  string `yaml:"passfile"`
}

type ConfigLoader interface {
	LoadConfig() (*Config, error)
}

type defaultConfigLoader struct{}

func (cl defaultConfigLoader) LoadConfig() (*Config, error) {
	xdgConfig := "gojira-cli/config.yml"
	filePath, err := xdg.SearchConfigFile(xdgConfig)
	if err != nil {
		fmt.Println("No config file. Adding one.")
		filePath, err = xdg.ConfigFile(xdgConfig)
		if err != nil {
			return nil, err
		}

		f, err := os.Create(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating %s", filePath)
		}
		defer f.Close()
		f.WriteString(defaultConfigContents)
		f.Close()
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	decoder := yaml.NewDecoder(bytes.NewReader(bs))
	err = decoder.Decode(config)
	if err != nil {
		return config, err // return incomplete object as well
	}

	//log.Printf("config=%+v", config)

	return config, nil
}

func NewConfigLoader() ConfigLoader {
	return defaultConfigLoader{}
}

//type JQLConfig struct {
//Name string `yaml:"name"`
//JQL  string `yaml:"jql"`
//}

//type JiraConfig struct {
//Queries []JQLConfig `yaml:"queries"`
//Labels  []string    `yaml:"labels"`
//}
