package handlers

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"

	"testapp/internal/models"
	pkgHTTP "testapp/pkg/http"
)

const (
	FOO_PATH = "/foo"
	JSON_PATH = "/json"
	XML_PATH = "/xml"
)

var jsonData, xmlData []byte

func init() {
	var err error

	jsonData, err = json.MarshalIndent(models.Books, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	xmlData, err = xml.MarshalIndent(models.BookListS, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
}	

type FormatHandler struct {
}

func NewFormatHandler() *FormatHandler {
	return &FormatHandler{}
}

func (*FormatHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(pkgHTTP.GetPath(FOO_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello, world!"))
	})

	mux.HandleFunc(pkgHTTP.GetPath(JSON_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Write(jsonData)
	})

	mux.HandleFunc(pkgHTTP.GetPath(XML_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/xml; charset=utf-8")
		resp.Write(xmlData)
	})
}