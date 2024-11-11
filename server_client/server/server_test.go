package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var client *http.Client 

func init() {
	if err := os.Chdir(".."); err != nil {
		log.Fatal(err)
	}

	client = &http.Client{
		Transport: &http.Transport{MaxIdleConnsPerHost: 100},
	}

	var port uint16 = 8080
	srv = NewServer(port)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}


func TestServerGetEndpoint(t *testing.T) {
	var tests = []struct{
		endpoint string
		want int
	}{
		{FOO_PATH, http.StatusOK},
		{JSON_PATH, http.StatusOK},
		{XML_PATH, http.StatusOK},
		{"/fooo", http.StatusNotFound},
		{DOWNLOAD_PATH, http.StatusOK},
	}

	for _, test := range tests {
		resp, err := client.Get(fmt.Sprintf("http://0.0.0.0:8080%s", test.endpoint))
		if err != nil {
			t.Error(err)
		}
		if got := resp.StatusCode; got != test.want {
			t.Errorf("Wrong resp.statusCode on %s endpoint, statusCode is %d, want %d", test.endpoint, got, test.want)
		}
	}
}

func TestServerJSON(t *testing.T) {
	resp, err := client.Get("http://0.0.0.0:8080/json")
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var books []Book
	if err = json.Unmarshal(body, &books); err != nil {
		t.Fatal(err)
	}

	want := 3
	if got := len(books); got != want {
		t.Errorf("Books len = %d, want %d", got, want)
	}
}

func TestDownloadFile(t *testing.T) {
	resp, err := client.Get("http://0.0.0.0:8080/download")
	if err != nil {
		t.Fatal(err)
	}

	//check content type
	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type is %s, want = image/png", ct)
	}

	//check content disposition
	cd := resp.Header.Get("Content-Disposition")
	cdSlice := strings.Split(cd, ";")
	cd = cdSlice[0]
	if cd != "attachment" {
		t.Errorf("Content-Type is %s, want = attachment", cd)
	}

	//check filename
	if len(cdSlice) > 1 {
		for _, param := range cdSlice[1:] {
			if strings.HasPrefix(strings.TrimSpace(param), "filename=") {
				filename := strings.TrimSpace(param)[len("filename="):]
				if filename != FILENAME {
					t.Errorf("Filename is %s, want = %s", filename, FILENAME)
				}
			}
		}
	} else {
		t.Fatalf("There is no filename parameter in Content-Disposition")
	}
	
	//check file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fileBytes, err := os.ReadFile("./tmp/"+FILENAME)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(body, fileBytes) {
		t.Errorf("Body content is not identical to original file")
	}
}

func TestUploadFile(t *testing.T) {
	resp, err := SendFile()
	if err != nil {
		t.Fatal(err)
	}

	if got := resp.StatusCode; got != http.StatusOK {
		t.Errorf("Wrong resp.statusCode on %s endpoint, statusCode is %d, want %d", UPLOAD_PATH, got, http.StatusOK)
	}
}

func TestMaxUploadSize(t *testing.T) {
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("myfiles", FILENAME)
	if err != nil {
		t.Fatal(err)
	}

	content := make([]byte, maxUploadSize+1) 

	_, err = io.Copy(fileWriter, bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://0.0.0.0:8080%s", UPLOAD_PATH), body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != http.StatusRequestEntityTooLarge {
		t.Errorf("Sending oversized file on %s endpoint, statusCode is %d, want %d", UPLOAD_PATH, got, http.StatusRequestEntityTooLarge)
	}
}

func TestSaveFile(t *testing.T) {
	filename := "testfile.txt"

	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("myfiles", filename)
	if err != nil {
		t.Fatal(err)
	}

	content := []byte("this is test file")

	_, err = io.Copy(fileWriter, bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://0.0.0.0:8080%s", UPLOAD_PATH), body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	savedFilePath := filepath.Join("./uploads", filename)
	if !FileExists(savedFilePath) {
		t.Errorf("File %s is not saved on server in %s dir", filename, UPLOAD_PATH)
	} else {
		savedContent, err := os.ReadFile(savedFilePath)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(savedContent, content) {
			t.Errorf("Content of a saved file %s is not identical to content from client's file", filename)
		}
	}
}


func BenchmarkServerFooEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := client.Get("http://0.0.0.0:8080/foo")
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkServerJSONEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := client.Get("http://0.0.0.0:8080/json")
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkServerXMLEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := client.Get("http://0.0.0.0:8080/xml")
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkServerDownloadEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := client.Get("http://0.0.0.0:8080/download")
		if err != nil {
			b.Error(err)
		}
	}
}


func SendFile() (*http.Response, error) {
	var body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("myfiles", FILENAME)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath.Join("./uploads", FILENAME))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fileWriter, file) 
	if err != nil {
		return nil, err
	}

	file.Close()
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://0.0.0.0:8080%s", UPLOAD_PATH), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}