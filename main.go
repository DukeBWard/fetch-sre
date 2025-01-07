package main

import (
	"flag"
	"log"
	"os"

	"github.com/robfig/cron"
	"gopkg.in/yaml.v2"
)

type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Method  string            `yaml:"method,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

type DomainStatus struct {
	Requests int
	UpCount  int
}

func main() {
	// get file path from --flag
	fp := flag.String("file", "", "File path to YAML file with HTTP endpoints")
	flag.Parse()

	if *fp == "" {
		log.Fatal("Error: proper usage= --file=/path/to/endpoints.yaml")
	}

	// init a map of domain statuses
	domainStatusMap := make(map[string]*DomainStatus)

	// load endpoints from YAML if valid
	endpoints, err := loadYAML(*fp)
	if err != nil {
		log.Fatalf("Failure to load YAML: ", err)
	}

}

func startTimer() {
	c := cron.New()

	// use cron for scheduling
	err := c.AddFunc("@every 15s", func() {
		//do stuff
	})
	if err != nil {
		log.Fatalf("Error adding cron job: ", err)
	}

	c.Start()
}

func loadYAML(fp string) ([]Endpoint, error) {
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	// unmarshals endpoints from a yaml file or errors out
	err = yaml.Unmarshal(data, &endpoints)
	return endpoints, err
}
