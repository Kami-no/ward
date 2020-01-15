package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type config struct {
	GProject  int      `yaml:"GProject"`
	GToken    string   `yaml:"GToken"`
	GURL      string   `yaml:"GURL"`
	GUser     string   `yaml:"GUser"`
	MBad      string   `yaml:"MBad"`
	MGood     string   `yaml:"MGood"`
	VBackend  []string `yaml:"VBackend"`
	VFrontend []string `yaml:"VFrontend"`
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
