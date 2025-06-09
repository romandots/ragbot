package tansultant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client provides methods to get information from the Tansultant API.
type Client struct {
	HTTPClient      *http.Client
	token           string
	addressEndpoint string
	pricesEndpoint  string
}

// NewClient creates a client using environment variables.
func NewClient() *Client {
	loadConfig()
	return &Client{
		HTTPClient:      &http.Client{Timeout: 10 * time.Second},
		token:           tansConfig.token,
		addressEndpoint: tansConfig.addressEndpoint,
		pricesEndpoint:  tansConfig.pricesEndpoint,
	}
}

func (c *Client) request(url string, v interface{}) error {
	if url == "" {
		return fmt.Errorf("endpoint not set")
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Branches returns available branches from the API.
func (c *Client) Branches() ([]Branch, error) {
	var branches []Branch
	if err := c.request(c.addressEndpoint, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

// Prices returns available passes and prices from the API.
func (c *Client) Prices() ([]Price, error) {
	var prices []Price
	if err := c.request(c.pricesEndpoint, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}
