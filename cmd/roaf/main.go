/*
Get the next dates for trash and paper pickup from ROAF

This program need a .env file for the following variables:

		ROAF_LOGFILE	- name of the logfile to use
	    NORKART_PROXY	- the URL to the norkart proxy ("https://norkartrenovasjon.azurewebsites.net/proxyserver.ashx")
	    ROAF_BASEURI	- URI for the query ("https://komteksky.norkart.no/MinRenovasjon.Api/api/tommekalender/%3F")
	    ROAF_KOMMNR		- kommunenummer (int)
	    ROAF_GATENAVN	- street name (string)
	    ROAF_HUSNR		- house number (string)
	    ROAF_GATEKODE	- street code
	    ROAF_APPKEY
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"

	"internal/iroaf"
)

var datolst []iroaf.Fraksjon
var jsonoutput bool

func main() {
	configureFlags()
	flag.Parse()
	
	loadEnv()

	logFile, err := os.OpenFile(os.Getenv("ROAF_LOGFILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	configureLogging(logFile)

	// build the URL
	url := buildUrl()

	// create request
	req := createGetRequest(url)

	res := doRequest(req)
	parseResponse(res)
}

func configureFlags() {
	flag.BoolVar(&jsonoutput, "j", false, "generate JSON-output")
}

func loadEnv() {
	// load environment from an env
	uhomedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Unable to get user homedir")
	}

	envLocations := []string{uhomedir + "/.roaf", ".roaf", ".env"}

	for _, loc := range envLocations {
		if _, err := os.Stat(loc); err != nil {
			continue
		}
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}
	}
}

// configureLogging does what it says on the tin
func configureLogging(f *os.File) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewJSONHandler(f, opts))
	slog.SetDefault(logger)
}

// buildUrl builds up a correct URL with params and query
func buildUrl() *url.URL {
	slog.Debug("Building URL")
	base, err := url.Parse(os.Getenv("NORKART_PROXY"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	q := base.Query()
	q.Set("server", os.Getenv("ROAF_BASEURI"))
	base.RawQuery = q.Encode()

	params := url.Values{}
	params.Add("kommunenr", os.Getenv("ROAF_KOMMNR"))
	params.Add("gatenavn", os.Getenv("ROAF_GATENAVN"))
	params.Add("husnr", os.Getenv("ROAF_HUSNR"))
	params.Add("gatekode", os.Getenv("ROAF_GATEKODE"))

	base.RawQuery += params.Encode()
	slog.Debug(base.String())
	return base
}

// createGetRequest builds and retuns a GET request,
// with correct headers set
func createGetRequest(uri *url.URL) *http.Request {
	slog.Debug("Creating request")
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	req.Header.Add("Kommunenr", os.Getenv("ROAF_KOMMNR"))
	req.Header.Add("RenovasjonAppKey", os.Getenv("ROAF_APPKEY"))

	return req
}

// doRequest executes the request and returns the response in a byte array
func doRequest(req *http.Request) []byte {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return resBody
}

// parseReponse unmarshals the JSON-response and prints out to stdout
func parseResponse(res []byte) {
	err := json.Unmarshal(res, &datolst)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if jsonoutput {
		for i, _ := range datolst {
			datolst[i].Enrich()
		}
		j, err := json.Marshal(datolst)
		if err != nil {
			fmt.Errorf("Unable to create JSON: %v", err)
		} else {
			fmt.Printf("%s\n", string(j))
		}
	} else {
		for i, _ := range datolst {
			fmt.Printf("%s\n", datolst[i])
		}
	}
}
