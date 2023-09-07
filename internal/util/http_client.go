package util

import (
	"encoding/json"
	"net/http"
	"time"
)

type HttpClient struct {
	client *http.Client
}

func (h *HttpClient) GetJSON(url string, target interface{}) error {
	r, err := h.client.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}
