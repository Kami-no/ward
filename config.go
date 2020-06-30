package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type config struct {
	MBad  string `yaml:"MBad"`
	MDown string `yaml:"MDown"`
	MFail string `yaml:"MFail"`
	MGood string `yaml:"MGood"`
	MUp   string `yaml:"MUp"`

	SMail string `yaml:"SMail"`

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
	Projects map[int]*Project `yaml:"Projects"`
}

type Project struct {
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
