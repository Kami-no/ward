package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

const fileName = "config.yaml"

type Config struct {
	SMail    string `yaml:"SMail"`
	NotifyBy string `yaml:"NotifyBy"`

	Credentials struct {
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
	} `yaml:"Credentials"`
	Endpoints struct {
		DC struct {
			Host   string `yaml:"Host"`
			Port   int    `yaml:"Port"`
			Domain string `yaml:"Domain"`
			Base   string `yaml:"Base"`
		} `yaml:"DC"`
		SMTP struct {
			Host     string `yaml:"Host"`
			Port     int    `yaml:"Port"`
			User     string `yaml:"User"`
			Password string `yaml:"Password"`
		} `yaml:"SMTP"`
		GitLab  string `yaml:"GitLab"`
		Webhook string `yaml:"Webhook"`
	} `yaml:"Endpoints"`
	Awards struct {
		Like         string `yaml:"Like"`
		Dislike      string `yaml:"Dislike"`
		Ready        string `yaml:"Ready"`
		NotReady     string `yaml:"NotReady"`
		NonCompliant string `yaml:"NonCompliant"`
	} `yaml:"Awards"`
	Projects map[int]*Project `yaml:"Projects"`
}

type Project struct {
	Teams map[string][]string `yaml:"Teams"`
	Votes int                 `yaml:"Votes"`
}

type TeamsWithMembers map[string]map[string]struct{}

func (p *Project) GetTeamsWithMembers() TeamsWithMembers {
	teamNamesWithMembers := make(TeamsWithMembers, len(p.Teams))
	for team, members := range p.Teams {
		membersMap := make(map[string]struct{})
		for _, member := range members {
			membersMap[member] = struct{}{}
		}
		teamNamesWithMembers[team] = membersMap
	}

	return teamNamesWithMembers
}

func NewConfig() *Config {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return &c
}
