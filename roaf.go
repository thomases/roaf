package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// load environment
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(os.Getenv("ROAF_LOGFILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	configureLogging(logFile)

	// build the URL
	url := buildUrl()
	fmt.Printf("URL: %q\n", url.String())

	// create request
	req := createGetRequest(url)

	doRequest(req)
}

func configureLogging(f *os.File) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewJSONHandler(f, opts))
	slog.SetDefault(logger)
}

func buildUrl() *url.URL {
	slog.Debug("Building URL")
	base, err := url.Parse(os.Getenv("NORKART_PROXY"))
	if err != nil {
		log.Fatal(err)
	}

	//base.Path += os.Getenv("ROAF_BASEURI")
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

func createGetRequest(uri *url.URL) *http.Request {
	slog.Debug("Creating request")
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Kommunenr", os.Getenv("ROAF_KOMMNR"))
	req.Header.Add("RenovasjonAppKey", os.Getenv("ROAF_APPKEY"))

	return req
}

func doRequest(req *http.Request) {
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", resBody)
}
