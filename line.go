package main

import (
	"net/http"
)

func sendLineMessage(apiURL, accessToken, message string) error {
	req, _ := http.NewRequest("POST", apiURL, nil)
	req.Header.Set("Content-Type",
		"application/json")
	req.Header.Set("Authorization", "Bearer "+
		accessToken)

	resp, _ := httpClient.Do(req)
	defer resp.Body.Close()

	return nil
}
