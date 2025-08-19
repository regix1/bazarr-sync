package client

import (
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	client http.Client
	Token  string
}

var client *HttpClient

// Singleton
func GetClient(token string) *HttpClient {
	if client == nil {
		client = &HttpClient{
			Token: token,
			client: http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}
	return client
}

// Rewrite of the Do method adding the api auth as a header
func (c *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Add("X-API-KEY", c.Token)
	if req.Method == "PATCH" {
		// Sync operations can take longer
		tempClient := http.Client{Timeout: 0}
		req.Header.Add("X-API-KEY", c.Token)
		return tempClient.Do(req)
	}
	return c.client.Do(req)
}

// The get request
func (c *HttpClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post request
func (c *HttpClient) Post(url string, params url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Patch request
func (c *HttpClient) Patch(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
