# prerequisites
* golang 1.22
* git

# how to install and run
1. Make sure you have Go installed and GOPATH and GOROOT configured properly
2. Clone this repository to your machine (ex. `git clone [url of this repo]`)
3. Run `go mod tidy` to install all external libraries
4. Run `go run . --file=test/sample1.yaml`

## if you want to build and run
4. Run `go build .` to build the binary
5. Run `.\fetch-sre.exe --file=test/sample1.yaml` to run the sample yaml file provided, or a filepath to a YAML file with HTML endpoints.

# run the test suite
* Run `go test -v` to run the test suite

# third party libraries
* `"gopkg.in/yaml.v2"`
* `"github.com/robfig/cron"`