package internal

import (
	"encoding/json"
	"io"
	"net/http"
)

type Ipify struct {}

func (i *Ipify) GetPublicIP() (*string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	//Decode the JSON.
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	ip := result["ip"]
	return &ip, nil
}