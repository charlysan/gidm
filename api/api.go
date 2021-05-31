package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type CustomHandler struct {
	h          http.Handler
	Router     *mux.Router
	ReqHeaders *map[string]string
	ResHeaders *map[string]string
	ReqBodyStr *map[string]string
	ResBodyStr *map[string]string
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *CustomHandler) HandleRequestHeaders(res http.ResponseWriter, req *http.Request) {
	handleRequest(h.ReqHeaders, res, req)
	fmt.Println("Request Headers Updated")
	for key, value := range *h.ReqHeaders {
		fmt.Println(" ", key+": "+value)
	}
}

func (h *CustomHandler) HandleResponseHeaders(res http.ResponseWriter, req *http.Request) {
	handleRequest(h.ResHeaders, res, req)
	fmt.Println("Response Headers updated")
	for key, value := range *h.ResHeaders {
		fmt.Println(" ", key+": "+value)
	}
}

func (h *CustomHandler) HandleRequestBodyStr(res http.ResponseWriter, req *http.Request) {
	handleRequest(h.ReqBodyStr, res, req)
	fmt.Println("Request Body string replacers updated:", h.ReqBodyStr)
	for key, value := range *h.ReqBodyStr {
		fmt.Println(" ", key+" -> "+value)
	}
}

func (h *CustomHandler) HandleResponseBodyStr(res http.ResponseWriter, req *http.Request) {
	handleRequest(h.ResBodyStr, res, req)
	fmt.Println("Response Body string replacers updated")
	for key, value := range *h.ResBodyStr {
		fmt.Println(" ", key+" -> "+value)
	}
}

func handleRequest(str *map[string]string, res http.ResponseWriter, req *http.Request) {
	var payload map[string]string
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&payload); err != nil {
		res.WriteHeader(400)
		return
	}
	defer req.Body.Close()

	*str = payload

	res.WriteHeader(200)
}
