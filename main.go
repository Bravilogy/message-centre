package main

import (
	"bytes"
	"encoding/json"
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
	Share(http.ResponseWriter)
}

type Service struct {
	Type    string          `json:"service"`
	Payload json.RawMessage `json:"payload"`
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

func (service Twitter) Share(w http.ResponseWriter) {
}

func (service Facebook) Share(w http.ResponseWriter) {
	fmt.Fprintln(w, service.Title+" is "+service.Description)
}

func (service Slack) Share(w http.ResponseWriter) {
	if service.Text == "" {
		http.Error(w, "Missing argument: Text", http.StatusNotFound)
		return
	}

	hook := os.Getenv("SLACK_HOOK")
	jsonValue, _ := json.Marshal(service)
	resp, err := http.Post(hook, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Could not connect to Slack", http.StatusNotFound)
		return
	}

	jsonMessage(w, "All done. I have sent -", service.Text)
}

func jsonMessage(w http.ResponseWriter, msg ...string) {
	response := map[string]string{
		"message": strings.Join(msg, " "),
	}

	jsonValue, _ := json.Marshal(response)
	w.Header().Set("content-type", "application/json")
	w.Write(jsonValue)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "What's your service about?", http.StatusNotFound)
		return
	}

	switch strings.ToLower(service.Type) {
	case "twitter":
		var s Twitter
		err := json.Unmarshal(service.Payload, &s)
		if err != nil {
			http.Error(w, "No marshals for Twitter today.", http.StatusNotFound)
			return
		}

		s.Share(w)

	case "facebook":
		var s Facebook
		err := json.Unmarshal(service.Payload, &s)
		if err != nil {
			http.Error(w, "No marshals for Facebook today.", http.StatusNotFound)
			return
		}

		s.Share(w)

	case "slack":
		var s Slack
		err := json.Unmarshal(service.Payload, &s)
		if err != nil {
			http.Error(w, "No marshals for Slack today.", http.StatusNotFound)
			return
		}

		s.Share(w)

	default:
		http.Error(w, "No dwarves here", http.StatusNotFound)
	}
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
