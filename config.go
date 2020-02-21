package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type config struct {
	GProject  int      `yaml:"GProject"`
	GPUrl     string   `yaml:"GPUrl"`
	GURL      string   `yaml:"GURL"`
	GUser     string   `yaml:"GUser"`
	LBase     string   `yaml:"LBase"`
	LHost     string   `yaml:"LHost"`
	LPass     string   `yaml:"LPass"`
	LUser     string   `yaml:"LUser"`
	MBad      string   `yaml:"MBad"`
	MDown     string   `yaml:"MDown"`
	MFail     string   `yaml:"MFail"`
	MGood     string   `yaml:"MGood"`
	MUp       string   `yaml:"MUp"`
	SHost     string   `yaml:"SHost"`
	SMail     string   `yaml:"SMail"`
	SPass     string   `yaml:"SPass"`
	SPort     int      `yaml:"SPort"`
	SUser     string   `yaml:"SUser"`
	VBackend  []string `yaml:"VBackend"`
	VFrontend []string `yaml:"VFrontend"`

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
			Host string `yaml:"Host"`
			Port int    `yaml:"Port"`
		} `yaml:"SMTP"`
		GitLab string `yaml:"GitLab"`
	} `yaml:"Endpoints"`
	Awards struct {
		Like         string `yaml:"Like"`
		Dislike      string `yaml:"Dislike"`
		Ready        string `yaml:"Ready"`
		NotReady     string `yaml:"NotReady"`
		NonCompliant string `yaml:"NonCompliant"`
	} `yaml:"Awards"`
	Projects []*Project `yaml:"Projects"`
}

type Project struct {
	ID    int                 `yaml:"ID"`
	Teams map[string][]string `yaml:"Teams"`
}

func (c *config) getConfig() *config {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
