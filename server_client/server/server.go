package server

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testapp/models"
	repsPgSQL "testapp/repositories/pgsql"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var jsonData, xmlData []byte
var srv *http.Server

const FOO_PATH, JSON_PATH, XML_PATH = "/foo", "/json", "/xml"
const DOWNLOAD_PATH, UPLOAD_PATH, SAVE_DB_PATH = "/download", "/upload", "/savedb"
const SHOW_PATH = "/show/{id}"
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


func NewMux(conn *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()
	
	mux.HandleFunc(getPath(FOO_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello, world!"))
	})

	mux.HandleFunc(getPath(JSON_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp.Write(jsonData)
	})

	mux.HandleFunc(getPath(XML_PATH), func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/xml; charset=utf-8")
		resp.Write(xmlData)
	})

	mux.HandleFunc(getPath(DOWNLOAD_PATH), func(resp http.ResponseWriter, req *http.Request) {
		fileBytes, err := ReadFile(FILENAME)
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Can't read the file", FILENAME)
			return
		}
		resp.Header().Set("Content-Type", "image/png")
		resp.Header().Set("Content-Disposition", "attachment; filename="+FILENAME)
		resp.Write(fileBytes)
	})
	
	mux.HandleFunc(postPath(UPLOAD_PATH), func(resp http.ResponseWriter, req *http.Request) {
		// if file is too large
		if req.ContentLength > maxUploadSize {
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

	mux.HandleFunc(postPath(SAVE_DB_PATH), func(resp http.ResponseWriter, req *http.Request) {
		// if file is too large
		if req.ContentLength > maxUploadSize {
			writeResponse(resp, http.StatusRequestEntityTooLarge, "File is too large")
			return
		}

		if err := req.ParseMultipartForm(maxUploadSize); err != nil {
			writeResponse(resp, http.StatusBadRequest)
			return
		}
		defer req.MultipartForm.RemoveAll()

		SaveFilesToDB(resp, req, conn)
	})

	mux.HandleFunc(getPath(SHOW_PATH), func(resp http.ResponseWriter, req *http.Request) {
		id, err := uuid.Parse(req.PathValue("id"))
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Error parsing the id")
			return
		}

		imageRep := repsPgSQL.NewImageRepository(conn)
		image, err := imageRep.Get(req.Context(), id)
		if err != nil {
			writeResponse(resp, http.StatusInternalServerError, "Error getting the image from db")
			return
		}
		resp.Header().Set("Content-Type", image.ContentType)
		resp.Write(image.Content)
	})

	return mux
}

func SaveFilesToDB(resp http.ResponseWriter, req *http.Request, conn *gorm.DB) {
	var content []byte
	var image models.Image

	imageRep := repsPgSQL.NewImageRepository(conn)

	ctx := req.Context()

	for _, header := range req.MultipartForm.File["myfiles"] {
		file, err := header.Open()
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Error retrieving a file")
			return
		}

		content, err = io.ReadAll(file)
		if err != nil {
			writeResponse(resp, http.StatusBadRequest, "Error reading content of a file")
			return
		}
		file.Close()

		image = models.NewImage(uuid.New(), header.Header.Get("Content-Type"), content)

		if err = imageRep.Create(ctx, image); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeResponse(resp, http.StatusNotFound, "Image not found")
			} else {
				writeResponse(resp, http.StatusInternalServerError, "Database error while retrieving image")
			}
			return
		}

		fmt.Fprintf(resp, "File uploaded successfully: %s\n", header.Filename)
	}
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

func NewServer(port uint16, conn *gorm.DB) *http.Server {
	srv = &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		Handler: NewMux(conn),
	}

	return srv 
}