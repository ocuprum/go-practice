package handlers

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

	"testapp/internal/handlers"
	"testapp/internal/models"
	repPgSQL "testapp/internal/repositories/pgsql"
	"testapp/internal/services"
	"testapp/pkg/config"
	pkgHTTP "testapp/pkg/http"
	pkgPgSQL "testapp/pkg/pgsql"
)

var client *http.Client 

func init() {
	if err := os.Chdir(filepath.Join("..")); err != nil {
		log.Fatal(err)
	}

	client = &http.Client{
		Transport: &http.Transport{MaxIdleConnsPerHost: 100},
	}

	config, err := config.LoadConfig("config", "yaml", filepath.Join(".", "configs"))
	if err != nil {
		log.Fatalf("Error loading a config: %v", err)
	}

	// Connecting to database
	db, err := pkgPgSQL.NewPgSQLConnection(config.PgSQL)
	if err != nil {
		log.Fatalf("Error connecting to pgsql db: %v", err)
	}

	imageRep := repPgSQL.NewImageRepository(db)
	imageServ := services.NewImageService(imageRep)
	imageHandler := handlers.NewImageHandler(imageServ)
	formatHandler := handlers.NewFormatHandler()

	srv := pkgHTTP.NewServer(config.HTTP, imageHandler, formatHandler)

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
		{handlers.FOO_PATH, http.StatusOK},
		{handlers.JSON_PATH, http.StatusOK},
		{handlers.XML_PATH, http.StatusOK},
		{"/fooo", http.StatusNotFound},
		{handlers.DOWNLOAD_PATH, http.StatusOK},
	}

	for _, test := range tests {
		u := fmt.Sprintf("http://0.0.0.0:8080%s", test.endpoint)
		resp, err := client.Get(u)
		if err != nil {
			t.Fatalf("Error making GET request: %v", err)
		}
		defer resp.Body.Close()
		assert.Equal(t, test.want, resp.StatusCode)
	}
}

func TestServerJSON(t *testing.T) {
	resp, err := client.Get(fmt.Sprintf("http://0.0.0.0:8080%s", handlers.JSON_PATH))
	require.NoError(t, err, "Client failed to GET the http://0.0.0.0:8080%s", handlers.JSON_PATH)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Error reading the resp.Body")

	var books []models.Book
	err = json.Unmarshal(body, &books)
	require.NoError(t, err, "Error unmarshalling the books json")

	want := 3
	assert.Equal(t, want, len(books), "Books len = %d, want %d", len(books), want)
}

func TestDownloadFile(t *testing.T) {
	resp, err := client.Get(fmt.Sprintf("http://0.0.0.0:8080%s", handlers.DOWNLOAD_PATH))
	require.NoError(t, err, "Client failed to GET the http://0.0.0.0:8080%s", handlers.DOWNLOAD_PATH)

	//check content type
	ct := resp.Header.Get("Content-Type")
	want := "image/png"
	require.Equal(t, want, ct, "Content-Type is %s, want = %s", ct, want)

	//check content disposition
	cd := resp.Header.Get("Content-Disposition")
	cdSlice := strings.Split(cd, ";")
	cd = cdSlice[0]
	want = "attachment"
	require.Equal(t, want, cd, "Content-Disposition is %s, want = %s", cd, want)

	//check filename
	require.Greater(t, len(cdSlice), 1, "There is no filename parameter in Content-Disposition")
	for _, param := range cdSlice[1:] {
		if strings.HasPrefix(strings.TrimSpace(param), "filename=") {
			filename := strings.TrimSpace(param)[len("filename="):]
			require.Equal(t, handlers.FILENAME, filename, "Filename is %s, want = %s", filename, handlers.FILENAME)
		}
	}
	
	//check file
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Error reading the resp.Body")


	fileBytes, err := os.ReadFile(filepath.Join(".", "assets", "tmp", handlers.FILENAME))
	require.NoError(t, err, "Error reading the resp.Body")

	assert.True(t, bytes.Equal(body, fileBytes), "Body content is not identical to original file")
}

func sendFile(t *testing.T, filename string, content []byte) (resp *http.Response) {
	var body = &bytes.Buffer{}
	
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("myfiles", filename)
	require.NoError(t, err, "Error creating a form")

	_, err = io.Copy(fileWriter, bytes.NewReader(content))
	require.NoError(t, err, "Error copying content to a writer")

	require.NoError(t, writer.Close(), "Error closing multipart writer")

	u := fmt.Sprintf("http://0.0.0.0:8080%s", handlers.UPLOAD_PATH)
	req, err := http.NewRequest(http.MethodPost, u, body)
	require.NoError(t, err, "Error creating a new POST request to %s", u)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = client.Do(req)
	require.NoError(t, err, "Error sending the POST request to %s", u)

	defer resp.Body.Close()

	return resp
}

func testUpload(t *testing.T, filename string, content []byte, errMsgTemplate string, want int) {
	resp := sendFile(t, filename, content)
	got := resp.StatusCode
	assert.Equal(t, want, got, errMsgTemplate, handlers.UPLOAD_PATH, got, want)
}

func TestUploadFile(t *testing.T) {
	fp := filepath.Join(".", "assets", "uploads", handlers.FILENAME)
	content, err := os.ReadFile(fp)
	require.NoError(t, err, "Error opening the file %s", fp)

	testUpload(t, handlers.FILENAME, content, "Wrong resp.statusCode on %s endpoint, statusCode is %d, want %d", http.StatusOK)
}

func TestMaxUploadSize(t *testing.T) {
	content := make([]byte, handlers.MaxUploadSize + 1)
	filename := "maxSizeTest.txt"
	testUpload(t, filename, content, "Sending oversized file on %s endpoint, statusCode is %d, want %d", http.StatusRequestEntityTooLarge)
}

func TestSaveFile(t *testing.T) {
	filename := "saveFileTest.txt"
	content := []byte("this is test file")
	testUpload(t, filename, content, "Error saving the file on %s endpoint, statusCode is %d, want %d", http.StatusOK)
	
	savedFilePath := filepath.Join(".", "assets", "uploads", filename)
	require.FileExists(t, savedFilePath, "File %s is not saved on server in %s dir", filename, handlers.UPLOAD_PATH)

	savedContent, err := os.ReadFile(savedFilePath)
	require.NoError(t, err, "Error reading saved file %s", savedFilePath)
	assert.Equal(t, savedContent, content, "Content of a saved file %s is not identical to content from client's file", filename)
}

func endpointBenchmark(b *testing.B, endpoint string) {
	u := fmt.Sprintf("http://localhost:8080%s", endpoint)
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(u)
		if err != nil {
			b.Fatalf("Error making GET request: %v", err)
		}
		defer resp.Body.Close()
		require.Equal(b, 200, resp.StatusCode, "Client failed to GET the http://0.0.0.0:8080%s", endpoint)
	}
}

func BenchmarkServerFooEndpoint(b *testing.B) {
	endpointBenchmark(b, handlers.FOO_PATH)
}

func BenchmarkServerJSONEndpoint(b *testing.B) {
	endpointBenchmark(b, handlers.JSON_PATH)
}

func BenchmarkServerXMLEndpoint(b *testing.B) {
	endpointBenchmark(b, handlers.XML_PATH)
}