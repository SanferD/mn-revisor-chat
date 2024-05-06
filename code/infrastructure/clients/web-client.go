package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPClientHelper struct{}

func InitializeHTTPClientHelper() (*HTTPClientHelper, error) {
	return &HTTPClientHelper{}, nil
}

func (httpClientHelper *HTTPClientHelper) GetHTML(ctx context.Context, url string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error on Get url for url='%s': %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-ok status code received, got status-code=%d: %v", resp.StatusCode, err)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error on readall for url='%s': %v", url, err)
	}
	return data, nil
}
