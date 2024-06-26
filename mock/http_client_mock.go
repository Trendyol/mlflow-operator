package mock

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type MockHTTPClient struct {
	Responses map[string]string
}

func (m *MockHTTPClient) SendGetRequest(url string, target interface{}) error {
	if responseJSON, ok := m.Responses[url]; ok {
		err := json.NewDecoder(io.NopCloser(strings.NewReader(responseJSON))).Decode(target)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unexpected URL: %s", url)
}

func (m *MockHTTPClient) SendPatchRequest(url string, _ interface{}, target interface{}) error {
	if responseJSON, ok := m.Responses[url]; ok {
		err := json.NewDecoder(io.NopCloser(strings.NewReader(responseJSON))).Decode(target)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unexpected URL: %s", url)
}
