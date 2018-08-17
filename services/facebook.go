package services

import (
	"fmt"
	"net/http"
)

type Facebook struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (service Facebook) Share(w http.ResponseWriter) error {
	fmt.Fprintln(w, service.Title+" is "+service.Description)
	return nil
}
