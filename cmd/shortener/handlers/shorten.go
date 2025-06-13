package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
	"uno/cmd/shortener/utils"
)

func ShortenURLHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		if originalURL == "" {
			http.Error(w, "empty URL", http.StatusBadRequest)
			return
		}

		if existingID, ok := store.FindByOriginal(originalURL); ok {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, cfg.BaseURL+"/"+existingID)
			return
		}

		shortID := utils.GenerateShortID()
		store.Save(shortID, originalURL)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, cfg.BaseURL+"/"+shortID)
	}
}

func ApiShortenHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var req models.APIRequest
		if err := req.UnmarshalJSON(data); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		originalURL := strings.TrimSpace(req.URL)
		if originalURL == "" {
			http.Error(w, "empty URL", http.StatusBadRequest)
			return
		}

		if existingID, ok := store.FindByOriginal(originalURL); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			resp := models.APIResponse{Result: cfg.BaseURL + "/" + existingID}
			data, err := resp.MarshalJSON()
			if err != nil {
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
			w.Write(data)
			return
		}

		shortID := utils.GenerateShortID()
		store.Save(shortID, originalURL)

		resp := models.APIResponse{
			Result: cfg.BaseURL + "/" + shortID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if data, err := resp.MarshalJSON(); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		} else {
			w.Write(data)
		}
	}
}

func BatchShortenHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var requests []models.BatchRequest
		if err := models.UnmarshalBatchRequest(data, &requests); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if len(requests) == 0 {
			http.Error(w, "empty batch", http.StatusBadRequest)
			return
		}

		pairs := make(map[string]string, len(requests))
		responses := make([]models.BatchResponse, 0, len(requests))

		for _, req := range requests {
			shortID := utils.GenerateShortID()
			pairs[shortID] = req.OriginalURL

			responses = append(responses, models.BatchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      cfg.BaseURL + "/" + shortID,
			})
		}

		if err := store.SaveBatch(pairs); err != nil {
			http.Error(w, "failed to save batch", http.StatusInternalServerError)
			return
		}

		respData, err := models.MarshalBatchResponse(responses)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(respData)
	}
}
