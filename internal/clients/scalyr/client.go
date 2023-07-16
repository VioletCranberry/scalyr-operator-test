package clients

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ApiClient struct {
	apiUrl     string
	apiKey     string
	apiTimeout int
	httpClient *resty.Client
}

type ApiResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewApiClient(apiUrl string, apiKey string, apiTimeout int) *ApiClient {
	timeout := time.Duration(apiTimeout) * time.Second
	if !strings.HasPrefix(apiUrl, "http") {
		apiUrl = fmt.Sprintf("https://%s", apiUrl)
	}
	client := resty.New()
	client.
		SetTimeout(timeout).
		SetBaseURL(apiUrl).
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiKey)

	return &ApiClient{
		httpClient: client,
	}
}

func (client *ApiClient) PostRequest(
	apiEndpoint string,
	request interface{},
	response interface{}) (*resty.Response, error) {
	resp, err := client.
		httpClient.R().
		SetBody(request).
		SetResult(response).
		SetError(response).
		Post(apiEndpoint)
	return resp, err
}
