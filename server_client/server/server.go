package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var jsonData, xmlData []byte
var srv *http.Server

const FOO_PATH, JSON_PATH, XML_PATH = "/foo", "/json", "/xml"
const DOWNLOAD_PATH, UPLOAD_PATH = "/download", "/upload"
const FILENAME = "diagram.png"

const maxUploadSize = 10 << 20

type Author struct {
	Name string
	Year int `xml:"born"`
}

type Book struct {
	Name string
	Authors []Author
	Year int `json:"published"`
	Read bool
	Comments map[string]string `json:"Comments,omitempty" xml:"-"`
}

type BookList struct {
	XMLName xml.Name `xml:"myshelf"`
	Items   []Book  `xml:"item"`
}


var books = []Book{
	//{Name: "Cool Shorts", Authors: []Author{{Name: "Your Mom", Year: 45}}, Year: 2200, Read: false},
	{Name: "Mama Pijama", Authors: []Author{{Name: "Frank Galagher", Year: 500}}, Year: 1900, Read: true, 
	Comments: map[string]string{
		"Vova": "Cool book", 
		"Sasha": "Awful",
	}},
	{Name: "FlowerGirl", Authors: []Author{{Name: "Frank", Year: 400},
										   {Name: "Lisy", Year: 2000}, 
										   {Name: "Aimee", Year: 1000}}, Year: 2000, Read: true},
	{Name: "FlowerBoy", Authors: []Author{{Name: "Frank", Year: 400},
										   {Name: "Lisy", Year: 2000}, 
										   {Name: "Aimee", Year: 1000}}, Year: 2000, Read: true}}


func init() {
	var err error

	jsonData, err = json.MarshalIndent(books, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	bookList := BookList{Items: books}

	xmlData, err = xml.MarshalIndent(bookList, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
}										   


func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(FOO_PATH, func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello, world!"))
	})

	mux.HandleFunc(JSON_PATH, func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Write(jsonData)
	})

	mux.HandleFunc(XML_PATH, func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/xml; charset=utf-8")
		resp.Write(xmlData)
	})

	mux.HandleFunc(DOWNLOAD_PATH, func(resp http.ResponseWriter, req *http.Request) {
		fileBytes, err := ReadFile(FILENAME)
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Can't read the file", FILENAME)
			return
		}
		resp.Header().Set("Content-Type", "image/png")
		resp.Header().Set("Content-Disposition", "attachment; filename="+FILENAME)
		resp.Write(fileBytes)
	})
	
	mux.HandleFunc(UPLOAD_PATH, func(resp http.ResponseWriter, req *http.Request) {
		// if file is too large
		if req.ContentLength > (maxUploadSize) {
			writeResponse(resp, http.StatusRequestEntityTooLarge, "File is too large")
			return
		}

		if err := req.ParseMultipartForm(maxUploadSize); err != nil {
			writeResponse(resp, http.StatusBadRequest)
			return
		}
		defer req.MultipartForm.RemoveAll()

		SaveFiles(resp, req)
	})

	return mux
}


func SaveFiles(resp http.ResponseWriter,req *http.Request) {
	// for every file
	for _, header := range req.MultipartForm.File["myfiles"] {
		// get file
		file, err := header.Open()
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Error retrieving file")
			return
		}
		defer file.Close()

		if !DirectoryExists("./uploads") {
			writeResponse(resp, http.StatusInternalServerError, "Directory doesn't exist")
			return
		}

		filename := filepath.Join("./uploads", header.Filename)
		newFile, err := os.Create(filename)
		if err != nil {
			writeResponse(resp, http.StatusInternalServerError, "Can't create a new file")
			return
		}
		defer newFile.Close()

		if _, err := io.Copy(newFile, file); err != nil {
			writeResponse(resp, http.StatusInternalServerError, "Can't save the file")
			return
		}

		fmt.Fprintf(resp, "File uploaded successfully: %s\n", filename)
	}
}

func FileExists(filename string) bool {
    _, err := os.Stat(filename)
    return !os.IsNotExist(err)
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func ReadFile(FILENAME string) ([]byte, error) {
	fileBytes, err := os.ReadFile("./tmp/"+FILENAME)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
} 

func NewServer(port uint16) *http.Server {
	srv = &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		Handler: NewMux(),
	}

	return srv 
}