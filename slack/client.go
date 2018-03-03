package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL = "https://slack.com/api/"
)

// Client is a client that authenticates requests to the Slack api.
type Client struct {
	Token   string
	BaseURL string
}

// NewClient creates a new slack client.
func NewClient(oauthToken string) *Client {
	return &Client{
		Token:   oauthToken,
		BaseURL: baseURL,
	}
}

func (c *Client) buildUrl(endpoint string) string {
	return fmt.Sprintf("%v%v", c.BaseURL, endpoint)
}

// Post sends a message to a given channel.
func (c *Client) Post(channel, message string) error {
	endpoint := "chat.postMessage"
	m := &Message{
		Channel: channel,
		Text:    message,
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}
	req, err := c.signedRequest("POST", c.buildUrl(endpoint), bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("failed to get signed request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("Bad response code: %v", resp.StatusCode)
	}
	return nil
}

// signedRequest adds the right headers and returns the request for further modification.
// TODO maybe make this public if needed.
func (c *Client) signedRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("could not make a request: %v", err)
	}
	req.Header.Add("Authorization", c.bearerTokenHeader())
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	return req, nil
}

func (c *Client) bearerTokenHeader() string {
	return fmt.Sprintf("Bearer %v", c.Token)
}

// Message is a struct for marshaling data into a format slack likes.
type Message struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}
