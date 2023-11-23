package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type HTTPClient interface {
	GetJSON(url string, target interface{}) error
}

type httpClient struct {
	client *retryablehttp.Client
}

func NewHTTPClient() HTTPClient {
	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = 5

	return &httpClient{
		client: retryableClient,
	}
}

func (h *httpClient) GetJSON(url string, target interface{}) error {
	resp, err := h.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
