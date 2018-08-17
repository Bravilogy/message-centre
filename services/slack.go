package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type Slack struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
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
