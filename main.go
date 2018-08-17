package main

import (
	"encoding/json"
	"log"
	"message-centre/services"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/jpillora/ipfilter"
)

var AllowedIPs = []string{"::1"}

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
			var s services.Twitter
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
			var s services.Facebook
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
			var s services.Slack
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
		BlockByDefault: true,
		AllowedIPs:     AllowedIPs,
	})

	if err != nil {
		log.Fatalln("Could not initiate an IP filter")
	}

	http.ListenAndServe(":8080", f.Wrap(r))
}
