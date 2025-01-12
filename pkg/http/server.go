package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

var srv *http.Server							   

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}

func FileExists(filename string) bool {
    _, err := os.Stat(filename)
    return !os.IsNotExist(err)
}

func ReadFile(FILENAME string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filepath.Join("assets", "tmp", FILENAME))
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
} 

func NewServer(conf Config, hh ...Handler) *http.Server {
	mux := NewMux()
	for _, h := range hh {
		h.Register(mux)
	}

	srv = &http.Server{
		Addr: fmt.Sprintf(":%v", conf.Port),
		Handler: mux,
	}

	return srv 
}