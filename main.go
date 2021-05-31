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
	"strconv"
	"strings"

	"github.com/charlysan/gidm/api"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
)

var reqHeadersSlice *cli.StringSlice
var resHeadersSlice *cli.StringSlice
var reqBodyStrSlice *cli.StringSlice
var resBodyStrSlice *cli.StringSlice
var port string
var portInteractive string
var baseURL string
var debug bool = false

var reqHeaders map[string]string
var resHeaders map[string]string
var reqBodyStr map[string]string
var resBodyStr map[string]string
var redirectURI *url.URL

func main() {
	reqHeadersSlice = cli.NewStringSlice()
	resHeadersSlice = cli.NewStringSlice()
	reqBodyStrSlice = cli.NewStringSlice()
	resBodyStrSlice = cli.NewStringSlice()

	app := &cli.App{
		Name:  "gidm",
		Usage: "Simple midm tool",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "reqh",
				Usage:       "inject request header",
				Destination: reqHeadersSlice,
			},
			&cli.StringSliceFlag{
				Name:        "resh",
				Usage:       "inject response header",
				Destination: resHeadersSlice,
			},
			&cli.StringSliceFlag{
				Name:        "reqb",
				Usage:       "replace string in request body (slash notation: old/new)",
				Destination: reqBodyStrSlice,
			},
			&cli.StringSliceFlag{
				Name:        "resb",
				Usage:       "replace string in response body (slash notation: old/new)",
				Destination: resBodyStrSlice,
			},
			&cli.StringFlag{
				Name:        "p",
				Usage:       "listen to port",
				Value:       "8080",
				Destination: &port,
			},
			&cli.StringFlag{
				Name:        "u",
				Usage:       "redirect to url",
				Value:       "http://localhost:9000",
				Destination: &baseURL,
			},
			&cli.StringFlag{
				Name:        "i",
				Usage:       "enable interactive mode (API server will listen on specified port)",
				Value:       "9090",
				Destination: &portInteractive,
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

	// Inject headers
	for key, value := range reqHeaders {
		req.Header.Set(key, value)
	}

	// Modify Body
	if len(reqBodyStr) > 0 {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("Error reading request body: %s", err.Error())
			return
		}

		for old, new := range reqBodyStr {
			buf = bytes.Replace(buf, []byte(old), []byte(new), -1)
		}
		req.ContentLength = int64(len(buf))
		req.Header.Set("Content-Length", strconv.Itoa(len(buf)))

		reader := ioutil.NopCloser(bytes.NewBuffer(buf))
		req.Body = reader
	}

	// httputil.NewSingleHostReverseProxy does not set the host of the request to the host of the destination server.
	req.Host = redirectURI.Host
	proxy.Transport = &myTransport{}

	if debug {
		reqDump, _ := httputil.DumpRequest(req, true)
		log.Println(string(reqDump))
	}

	proxy.ServeHTTP(res, req)
}

type myTransport struct {
	// Uncomment this if you want to capture the transport
	CapturedTransport http.RoundTripper
}

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	for old, new := range resBodyStr {
		b = bytes.Replace(b, []byte(old), []byte(new), -1)
	}
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	// Inject headers
	for key, value := range resHeaders {
		resp.Header.Set(key, value)
	}

	if debug {
		resDump, _ := httputil.DumpResponse(resp, true)
		log.Println(string(resDump))
	}

	return resp, nil
}

func run(c *cli.Context) error {
	uri, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Wrong URL format", err.Error())
	}

	redirectURI = uri

	fmt.Println("Listening on port:", port)
	fmt.Println("Redirecting to:", baseURL)

	// Parse request headers
	if len(reqHeadersSlice.Value()) > 0 {
		reqHeaders = parseStringHeader(reqHeadersSlice)
		fmt.Println("\nRequest headers to be injected:")
		for key, value := range reqHeaders {
			fmt.Println(" ", key+": "+value)
		}
	}

	// Parse response headers
	if len(resHeadersSlice.Value()) > 0 {
		resHeaders = parseStringHeader(resHeadersSlice)
		fmt.Println("\nResponse headers to be injected:")
		for key, value := range resHeaders {
			fmt.Println(" ", key+": "+value)
		}
	}

	// Parse request Body Strings
	if len(reqBodyStrSlice.Value()) > 0 {
		reqBodyStr = parseStringReplacers(reqBodyStrSlice)
		fmt.Println("\nRequest body strings to be replaced:")
		for old, new := range reqBodyStr {
			fmt.Println(" ", old+" -> "+new)
		}
	}

	// Parse response Body Strings
	if len(resBodyStrSlice.Value()) > 0 {
		resBodyStr = parseStringReplacers(resBodyStrSlice)
		fmt.Println("\nResponse body strings to be replaced:")
		for old, new := range resBodyStr {
			fmt.Println(" ", old+" -> "+new)
		}
	}

	if len(portInteractive) > 0 {
		handler := api.CustomHandler{
			Router:     mux.NewRouter(),
			ReqHeaders: &reqHeaders,
			ResHeaders: &resHeaders,
			ReqBodyStr: &reqBodyStr,
			ResBodyStr: &resBodyStr,
		}

		handler.Router.HandleFunc("/requestHeaders", handler.HandleRequestHeaders).Methods("PUT")
		handler.Router.HandleFunc("/responseHeaders", handler.HandleResponseHeaders).Methods("PUT")
		handler.Router.HandleFunc("/requestBodyReplacers", handler.HandleRequestBodyStr).Methods("PUT")
		handler.Router.HandleFunc("/responseBodyReplacers", handler.HandleResponseBodyStr).Methods("PUT")

		go func() {
			log.Fatal(http.ListenAndServe(":"+portInteractive, handler.Router))
		}()
	}
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	return nil
}

func parseStringReplacers(srcStrSlice *cli.StringSlice) map[string]string {
	resBodyStr = make(map[string]string, len(srcStrSlice.Value()))
	for _, repStr := range srcStrSlice.Value() {
		// TODO: support "/" parsing
		srcTgt := strings.Split(repStr, "/")
		if len(srcTgt) == 2 {
			resBodyStr[srcTgt[0]] = srcTgt[1]
		}
	}

	return resBodyStr
}

func parseStringHeader(srcStrSlice *cli.StringSlice) map[string]string {
	strHeader := make(map[string]string, len(srcStrSlice.Value()))
	for _, h := range srcStrSlice.Value() {
		header := strings.Split(strings.ReplaceAll(h, " ", ""), ":")
		if len(header) == 2 {
			strHeader[header[0]] = header[1]
		}
	}

	return strHeader
}
