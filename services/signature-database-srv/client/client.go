package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	client *http.Client
	host   string
}

func New() *Client {
	return &Client{
		client: &http.Client{},
		host:   `https://api.openchain.xyz/signature-database`,
	}
}

func (c *Client) do(method string, path string, in any, out any) error {
	var bodyReader io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.host, path), bodyReader)
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	var responseWrapper struct {
		Ok     bool            `json:"ok"`
		Error  string          `json:"error"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseWrapper); err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if !responseWrapper.Ok {
		return errors.New(responseWrapper.Error)
	}

	if err := json.Unmarshal(responseWrapper.Result, out); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	return nil
}

func (c *Client) Import(data AllTypes[[]string]) (ImportResponse, error) {
	var resp ImportResponse

	err := c.do("POST", "/v1/import", ImportRequest(data), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
