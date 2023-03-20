package solidityclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	host   string
	client *http.Client
}

func New() *Client {
	return NewWithHost("https://api.openchain.xyz/solidity-compiler")
}

func NewWithHost(host string) *Client {
	return &Client{
		host:   host,
		client: &http.Client{},
	}
}

func (c *Client) Compile(input *CompileRequest) (*SolcStandardOutput, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.host+"/v1/compile", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCompilerUnavailable, err.Error())
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%w: expected http 200 but got %d", ErrCompilerUnavailable, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var compilerResp CompileResponse
	if err := json.Unmarshal(body, &compilerResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !compilerResp.Ok {
		return nil, fmt.Errorf("%w: %s", ErrCompilerUnavailable, compilerResp.Error)
	}

	return compilerResp.Result, nil
}
