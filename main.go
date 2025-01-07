package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron"
	"gopkg.in/yaml.v2"
)

// http endpoints
// used map here, no dictionary in go
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

	// block the main thread
	select {}
}

func runChecks(endpoints []Endpoint, domainStatusMap map[string]*DomainStatus) {
	client := http.Client{}

	for _, endpoint := range endpoints {
		// get hostname from url
		parsedURL, err := url.Parse(endpoint.URL)
		if err != nil {
			log.Printf("invalid url: %s", endpoint.URL)
			continue
		}

		domain := parsedURL.Hostname()

		// if domain is not in map, add it
		if _, exits := domainStatusMap[domain]; !exits {
			domainStatusMap[domain] = &DomainStatus{}
		}
		domainStatusMap[domain].Requests++

		// create a new request
		method := endpoint.Method
		if method == "" {
			method = "GET"
		}

		var bodyReader *strings.Reader
		if endpoint.Body != "" {
			bodyReader = strings.NewReader(endpoint.Body)
		} else {
			bodyReader = strings.NewReader("")
		}

		req, err := http.NewRequest(method, endpoint.URL, bodyReader)
		if err != nil {
			log.Printf("Error in request: %s", err)
			continue
		}

		// append headers to request
		for key, val := range endpoint.Headers {
			req.Header.Set(key, val)
		}

		// make the request but get milliseconds res time
		startTime := time.Now()
		res, err := client.Do(req)
		totalTime := time.Since(startTime).Milliseconds()

		if err != nil {
			// if error, its down
			log.Printf("DOWN: %s, %s", domain, err)
			continue
		}

		// don't forget to close body
		res.Body.Close()

		// check for 200 status code and proper latency < 500
		if (res.StatusCode >= 200 && res.StatusCode < 300) && totalTime < 500 {
			domainStatusMap[domain].UpCount++
			log.Printf("UP: %s, %dms", domain, totalTime)
		} else {
			log.Printf("DOWN: %s, %dms", domain, totalTime)
		}
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
