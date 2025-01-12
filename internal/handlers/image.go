package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"testapp/internal/services"
	pkgHTTP "testapp/pkg/http"
)

const (
	DOWNLOAD_PATH = "/download" 
	UPLOAD_PATH = "/upload" 
	SAVE_DB_PATH = "/savedb"
	SHOW_PATH = "/show/{id}"

	FILENAME = "diagram.png"

	MaxUploadSize = 10 << 20
)

type ImageHandler struct {
	serv *services.ImageService
}

func NewImageHandler(serv *services.ImageService) *ImageHandler {
	return &ImageHandler{serv: serv}
}

func (h *ImageHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(pkgHTTP.GetPath(DOWNLOAD_PATH), h.download)
	mux.HandleFunc(pkgHTTP.PostPath(UPLOAD_PATH), h.upload)
	mux.HandleFunc(pkgHTTP.PostPath(SAVE_DB_PATH), h.saveDB)
	mux.HandleFunc(pkgHTTP.GetPath(SHOW_PATH), h.show)
}

func (h *ImageHandler) download(resp http.ResponseWriter, req *http.Request) {
	fileBytes, err := h.serv.ReadFile(req.Context(), FILENAME)
	if err != nil {
		pkgHTTP.WriteResponse(resp, http.StatusBadRequest, "Can't read the file", FILENAME)
		return
	}
	resp.Header().Set("Content-Type", "image/png")
	resp.Header().Set("Content-Disposition", "attachment; filename="+FILENAME)
	resp.Write(fileBytes)
}

func (h *ImageHandler) upload(resp http.ResponseWriter, req *http.Request) {
	// if file is too large
	if req.ContentLength > MaxUploadSize {
		pkgHTTP.WriteResponse(resp, http.StatusRequestEntityTooLarge, "File is too large")
		return
	}

	if err := req.ParseMultipartForm(MaxUploadSize); err != nil {
		pkgHTTP.WriteResponse(resp, http.StatusBadRequest)
		return
	}
	defer req.MultipartForm.RemoveAll()

	h.saveFiles(resp, req)
}

func (h *ImageHandler) saveDB(resp http.ResponseWriter, req *http.Request) {
	// if file is too large
	if req.ContentLength > MaxUploadSize {
		pkgHTTP.WriteResponse(resp, http.StatusRequestEntityTooLarge, "File is too large")
		return
	}

	if err := req.ParseMultipartForm(MaxUploadSize); err != nil {
		pkgHTTP.WriteResponse(resp, http.StatusBadRequest)
		return
	}
	defer req.MultipartForm.RemoveAll()

	h.saveFilesToDB(resp, req)
}

func (h *ImageHandler) show(resp http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		pkgHTTP.WriteResponse(resp, http.StatusBadRequest, "Error parsing the id")
		return
	}

	image, err := h.serv.Get(req.Context(), id)
	if err != nil {
		pkgHTTP.WriteResponse(resp, http.StatusInternalServerError, "Error getting the image from db")
		return
	}
	resp.Header().Set("Content-Type", image.ContentType)
	resp.Write(image.Content)
}

func (h *ImageHandler) saveFilesToDB(resp http.ResponseWriter, req *http.Request) {
	for _, header := range req.MultipartForm.File["myfiles"] {
		file, err := header.Open()
		if err != nil {
			pkgHTTP.WriteResponse(resp, http.StatusBadRequest, "Error retrieving a file")
			return
		}
		defer file.Close()

		err = h.serv.SaveFileToDB(req.Context(), header.Header.Get("Content-Type"), file)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				pkgHTTP.WriteResponse(resp, http.StatusNotFound, "Image not found")
			} else {
				pkgHTTP.WriteResponse(resp, http.StatusInternalServerError, "Database error while retrieving image")
			}
		}

		fmt.Fprintf(resp, "File uploaded successfully: %s\n", header.Filename)
	}
}

func (h *ImageHandler) saveFiles(resp http.ResponseWriter, req *http.Request) {
	// for every file
	for _, header := range req.MultipartForm.File["myfiles"] {
		// get file
		file, err := header.Open()
		if err != nil {
			pkgHTTP.WriteResponse(resp, http.StatusBadRequest, "Error retrieving file")
			return
		}
		defer file.Close()

		err = h.serv.SaveFile(req.Context(), header.Filename, file)
		if err != nil {
			pkgHTTP.WriteResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}

		fmt.Fprintf(resp, "File uploaded successfully: %s\n", header.Filename)
	}
}