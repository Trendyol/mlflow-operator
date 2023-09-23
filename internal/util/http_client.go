package util

import (
	"encoding/json"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HTTPClient) GetJSON(url string, target interface{}) error {
	r, err := h.client.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
