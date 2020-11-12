package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func webhookSend(endpoint string, recipients []string, subj string, msg string, url string) error {

	rcpt := fmt.Sprint(strings.Join(recipients[:], ","))
	requestBody, err := json.Marshal(map[string]string{
		"rcpt":    rcpt,
		"title":   subj,
		"message": msg,
		"url":     url,
	})
	if err != nil {
		return fmt.Errorf("Failed to generate json: %v", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Failed to post message: %v", err)
	}

	defer resp.Body.Close()

	return nil
}
