package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Fraksjon struct {
	FraksjonId  int
	TommeDatoer []string
}

var datolst []Fraksjon

func main() {
	// load environment from .env
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

	// create request
	req := createGetRequest(url)

	res := doRequest(req)
	parseResponse(res)
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

func parseResponse(res []byte) {
	err := json.Unmarshal(res, &datolst)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	nextrest, err := time.Parse("2006-01-02T15:04:05", datolst[0].TommeDatoer[0])
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	nextpapp, err := time.Parse("2006-01-02T15:04:05", datolst[1].TommeDatoer[0])
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Restavfall: %s\n", nextrest.Format("2006-01-02"))
	fmt.Printf("Papp/Papir: %s\n", nextpapp.Format("2006-01-02"))
}
