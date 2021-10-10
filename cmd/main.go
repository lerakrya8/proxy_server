package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"leraProxy/database"
	"leraProxy/request"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type HandlerHttp struct {
	DB     database.Database
	router mux.Router
}

func (h *HandlerHttp) ConfigureRouter() {
	h.router.HandleFunc("/repeat/{id:[0-9]+}", h.Repeat)
	h.router.HandleFunc("/requests", h.AllRequests)
	h.router.HandleFunc("/requests/{id:[0-9]+}", h.OneRequest)
}

func (h *HandlerHttp) Repeat(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	request, err := h.DB.GetRequest(requestID)
	if err != nil {
		fmt.Println(err)
		return
	}

	req := &http.Request{
		Method: request.Method,
		URL: &url.URL{
			Scheme: "http",
			Host:   request.Host,
			Path:   request.URL,
		},
		Body:   ioutil.NopCloser(strings.NewReader(request.Body)),
		Host:   request.Host,
		Header: request.Headers,
	}

	h.HandleHTTPRequest(w, req)
}

func (h *HandlerHttp) AllRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.DB.GetAllRequests()
	if err != nil {
		fmt.Println(err)
		return
	}

	var response string
	for _, request := range requests {
		response += "Id: " + strconv.Itoa(int(request.ID)) +
			"\nMethod: " + request.Method +
			"\nURL: " + request.URL +
			"\nHost: " + request.Host +
			"\nBody: " + request.Body + "\n"
	}

	w.Write([]byte(response))
}

func (h *HandlerHttp) OneRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		return
	}
	request, err := h.DB.GetRequest(requestID)
	if err != nil {
		fmt.Println(err)
		return
	}

	var response string
	response += "Id: " + strconv.Itoa(int(request.ID)) +
		"\nMethod: " + request.Method +
		"\nURL: " + request.URL +
		"\nHost: " + request.Host +
		"\nBody: " + request.Body + "\n"

	w.Write([]byte(response))
}

func (h *HandlerHttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *HandlerHttp) HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	requestToSave := &request.Request{
		Method:  r.Method,
		Host:    r.Host,
		URL:     r.URL.Path,
		Headers: r.Header,
		Body:    string(body),
	}

	headers, err := json.Marshal(requestToSave.Headers)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	err = h.DB.Save(requestToSave, string(headers))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	response, err := h.makeRequest(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.WriteHeader(response.StatusCode)

	io.Copy(w, response.Body)
	defer response.Body.Close()
}

func (h *HandlerHttp) makeRequest(r *http.Request) (*http.Response, error) {
	request, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	request.Header = r.Header

	proxyResponse, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	return proxyResponse, nil
}

func main() {
	con, err := sql.Open("postgres",
		"host=localhost port=5432 user=lera dbname=proxy_bd password=123456 sslmode=disable")

	if err != nil {
		fmt.Println(err)
		return
	}
	err = con.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}

	handler := &HandlerHttp{
		DB: database.Database{con},
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handler.HandleHTTPRequest),
	}

	go server.ListenAndServe()

	handlerApi := &HandlerHttp{
		DB: database.Database{con},
	}

	handlerApi.ConfigureRouter()
	serverApi := http.Server{
		Addr:    ":8000",
		Handler: handlerApi,
	}

	serverApi.ListenAndServe()
	fmt.Println("закончили")
}
