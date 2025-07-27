package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20 // 10 MB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to get thumbnail file", err)
		return
	}
	defer file.Close()

	fileType := header.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(fileType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", err)
		return
	}

	if mediatype != "image/jpeg" || mediatype != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Unsupported file type", nil)
		return
	}

	dbVid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get video from database", err)
		return
	}

	if dbVid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You are not allowed to upload a thumbnail for this video", nil)
		return
	}

	fileExt := strings.Split(fileType, "/")[1]
	filepath := filepath.Join(cfg.assetsRoot, videoID.String()+"."+fileExt)
	osfile, err := os.Create(filepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create thumbnail file", err)
		return
	}
	defer osfile.Close()

	_, err = io.Copy(osfile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save thumbnail file", err)
		return
	}

	stringURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID.String(), fileExt)
	dbVid.ThumbnailURL = &stringURL

	err = cfg.db.UpdateVideo(dbVid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video in database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVid)
}
