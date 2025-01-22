package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

	maxSize := 10 << 20 // 10 MB

	err = r.ParseMultipartForm(int64(maxSize))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing form", err)
		return
	}

	image, fHeader, err := r.FormFile("thumbnail")
	contentType := fHeader.Header.Get("Content-Type")

	ext := ".png"

	if contentType == "image/jpeg" {
		ext = ".jpg"
	} else if contentType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid image format", nil)
		return
	}

	randos := make([]byte, 32)

	rand.Read(randos)

	thumbName := base64.RawURLEncoding.EncodeToString([]byte(randos)) + ext

	thumbnailPath := "http://localhost:" + cfg.port + "/" + filepath.Join(cfg.assetsRoot, thumbName)

	imageData := make([]byte, fHeader.Size)

	_, err = io.ReadFull(image, imageData)

	video, err := cfg.db.GetVideo(videoID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not the owner of the video", nil)
		return
	}

	thumbFile, err := os.Create("./assets/" + thumbName)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating thumbnail file", err)
		return
	}

	thumbFile.Write(imageData)

	thumbnailUrl := thumbnailPath

	video.ThumbnailURL = &thumbnailUrl

	err = cfg.db.UpdateVideo(video)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
