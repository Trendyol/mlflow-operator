package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type HTTPClient interface {
	SendGetRequest(url string, target interface{}) error
	SendPatchRequest(url string, data interface{}, target interface{}) error
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

func (h *httpClient) SendGetRequest(url string, target interface{}) error {
	resp, err := h.client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	r := json.NewDecoder(resp.Body).Decode(target)
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return r
}

func (h *httpClient) SendPatchRequest(url string, data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	r := json.NewDecoder(resp.Body).Decode(target)
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return r
}
