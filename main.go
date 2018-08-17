package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/jpillora/ipfilter"
)

type Shareable interface {
	Share(http.ResponseWriter) error
}

type Service struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Payload struct {
	Services []Service
}

type Twitter struct {
	Tweet string
}

type Facebook struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Slack struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

func (service Twitter) Share(w http.ResponseWriter) error {
	return nil
}

func (service Facebook) Share(w http.ResponseWriter) error {
	fmt.Fprintln(w, service.Title+" is "+service.Description)
	return nil
}

func (service Slack) Share(w http.ResponseWriter) error {
	if service.Text == "" {
		return errors.New("Missing argument: Text")
	}

	hook := os.Getenv("SLACK_HOOK")
	jsonValue, _ := json.Marshal(service)
	resp, err := http.Post(hook, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil || resp.StatusCode != 200 {
		return errors.New("Could not connect to Slack")
	}

	return nil
}

func jsonMessage(w http.ResponseWriter, msg string) {
	response := map[string]string{
		"message": msg,
	}

	jsonValue, _ := json.Marshal(response)
	w.Header().Set("content-type", "application/json")
	w.Write(jsonValue)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Something's up with that payload", http.StatusNotFound)
		return
	}

	var result = make(map[string]int)

	for _, service := range payload.Services {
		switch strings.ToLower(service.Type) {
		case "twitter":
			var s Twitter
			err := json.Unmarshal(service.Payload, &s)
			if err != nil {
				http.Error(w, "No marshals for Twitter today.", http.StatusBadRequest)
				return
			}

			if err := s.Share(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result["twitter"] = result["twitter"] + 1

		case "facebook":
			var s Facebook
			err := json.Unmarshal(service.Payload, &s)
			if err != nil {
				http.Error(w, "No marshals for Facebook today.", http.StatusBadRequest)
				return
			}

			if err := s.Share(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result["facebook"] = result["facebook"] + 1

		case "slack":
			var s Slack
			err := json.Unmarshal(service.Payload, &s)
			if err != nil {
				http.Error(w, "No marshals for Slack today.", http.StatusNotFound)
				return
			}

			if err := s.Share(w); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			result["slack"] = result["slack"] + 1
		}
	}

	if len(result) < 1 {
		jsonMessage(w, "Looks like there is nothing for me to do here")
		return
	}

	jsonValue, _ := json.Marshal(result)
	w.Header().Set("content-type", "application/json")
	w.Write(jsonValue)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Could not parse .env", err.Error())
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("POST")

	f, err := ipfilter.New(ipfilter.Options{
		AllowedIPs:     []string{"::1"},
		BlockByDefault: true,
	})

	if err != nil {
		log.Fatalln("Could not initiate an IP filter")
	}

	http.ListenAndServe(":8080", f.Wrap(r))
}
