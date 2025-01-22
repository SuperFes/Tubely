package main

import (
	"github.com/google/uuid"
	"net/http"
	"path/filepath"
)

func (cfg *apiConfig) handlerThumbnailGet(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	data := []byte("/" + filepath.Join(cfg.assetsRoot, videoID.String()+".png"))

	_, err = w.Write(data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error writing response", err)
		return
	}
}
