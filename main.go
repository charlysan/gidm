package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var headersSlice *cli.StringSlice
var port string
var baseURL string
var debug bool = false

var headers map[string]string
var redirectURI *url.URL

func main() {
	headersSlice = cli.NewStringSlice()

	app := &cli.App{
		Name:  "gidm",
		Usage: "Simple midm tool",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "H",
				Usage:       "inject header to request",
				Destination: headersSlice,
			},
			&cli.StringFlag{
				Name:        "P",
				Usage:       "listen to port",
				Value:       "8080",
				Destination: &port,
			},
			&cli.StringFlag{
				Name:        "U",
				Usage:       "redirect to url",
				Value:       "http://localhost:9000",
				Destination: &baseURL,
			},
			&cli.BoolFlag{
				Name:        "d",
				Usage:       "enable debugging",
				Value:       false,
				Destination: &debug,
			},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(redirectURI)

	if debug {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("Error reading request body: %s", err.Error())
			return
		}

		reader := ioutil.NopCloser(bytes.NewBuffer(buf))
		fmt.Println()
		log.Printf("%s %s %s%s \n\n%s\n\n", req.Method, req.Proto, req.Host, req.URL.Path, string(buf))
		for key, value := range req.Header {
			fmt.Printf("%s: %s\n", key, value)
		}

		req.Body = reader
	}

	// Inject headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// httputil.NewSingleHostReverseProxy does not set the host of the request to the host of the destination server.
	req.Host = redirectURI.Host

	proxy.ServeHTTP(res, req)
}

func run(c *cli.Context) error {
	uri, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Wrong URL format", err.Error())
	}

	redirectURI = uri

	fmt.Println("Listening on port:", port)
	fmt.Println("Redirecting to:", baseURL)

	if len(headersSlice.Value()) > 0 {
		headers = make(map[string]string, len(headersSlice.Value()))
		for _, h := range headersSlice.Value() {
			header := strings.Split(strings.ReplaceAll(h, " ", ""), ":")
			if len(header) == 2 {
				headers[header[0]] = header[1]
			}
		}
		fmt.Println("Headers to be injected:")
		for key, value := range headers {
			fmt.Println(" ", key+": "+value)
		}
	}

	http.HandleFunc("/", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	return nil
}
