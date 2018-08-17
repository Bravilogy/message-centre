package services

import "net/http"

type Twitter struct {
	Tweet string
}

func (service Twitter) Share(w http.ResponseWriter) error {
	return nil
}
