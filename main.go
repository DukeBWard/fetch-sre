package main

import (
	"flag"
	"fmt"
	"log"
	"math"
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

	fmt.Println("Starting up...")

	// init a map of domain statuses
	domainStatusMap := make(map[string]*DomainStatus)

	// load endpoints from YAML if valid
	endpoints, err := loadYAML(*fp)
	if err != nil {
		log.Fatal("Failure to load YAML: ", err)
	}

	startTimer(endpoints, domainStatusMap)

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

		// this will give us all domains, including subdomains
		// publicsuffix.EffectiveTLDPlusOne will strip subdomain if wanted
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
			log.Printf("UP: %s, %d, %dms", domain, res.StatusCode, totalTime)
		} else {
			log.Printf("DOWN: %s, %d, %dms", domain, res.StatusCode, totalTime)
		}
	}
}

func startTimer(endpoints []Endpoint, domainStatusMap map[string]*DomainStatus) {
	c := cron.New()

	// use cron for scheduling
	err := c.AddFunc("@every 15s", func() {
		//do work
		runChecks(endpoints, domainStatusMap)
		getAvailPercent(domainStatusMap)
		fmt.Println("================================================================")
	})
	if err != nil {
		log.Fatal("Error adding cron job: ", err)
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

func getAvailPercent(domainStatusMap map[string]*DomainStatus) {
	for domain, status := range domainStatusMap {
		availPercent := 0.0
		if status.Requests > 0 {
			// 100 * (number of HTTP requests that had an outcome of UP / number of HTTP requests)
			fmt.Println(domain, "upcount=", status.UpCount, "reqcount=", status.Requests)
			availPercent = 100 * (float64(status.UpCount) / float64(status.Requests))
		}
		availPercentRound := int(math.Round(availPercent))
		fmt.Printf("%s has %d%% availability percentage\n", domain, availPercentRound)
	}
}
