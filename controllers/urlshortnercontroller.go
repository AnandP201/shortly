package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"shortly/generator"
)

type ShortnerController struct {
	SlugGen *generator.SlugGenerator
}

func (s *ShortnerController) UrlShortnerCreate(w http.ResponseWriter, r *http.Request) {

	sequenceId := os.Getenv("POD_NAME")

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Server healthy",
		"id":     sequenceId,
	})

}
