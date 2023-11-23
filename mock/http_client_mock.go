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

func (m *MockHTTPClient) GetJSON(url string, target interface{}) error {
	if responseJSON, ok := m.Responses[url]; ok {
		err := json.NewDecoder(io.NopCloser(strings.NewReader(responseJSON))).Decode(target)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unexpected URL: %s", url)
}
