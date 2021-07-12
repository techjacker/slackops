package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// SlackConnector is a slack API client interface
type SlackConnector interface {
	TestConnection() error
	PostMessage(string) error
}

// NewSlackClient is a slack API client factory
func NewSlackClient(token, channel string) SlackConnector {
	return &SlackClient{
		Channel: channel,
		Token:   token,
		client:  &http.Client{},
	}
}

// SlackClient is a slack API client
type SlackClient struct {
	Channel string
	Token   string
	client  *http.Client
}

// PostMessage sends a create message call to the slack API
func (c *SlackClient) PostMessage(msg string) error {
	payload := slackMessage{
		Channel: c.Channel,
		Text:    msg,
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&payload)
	if err != nil {
		return err
	}
	return c.makeReq("https://slack.com/api/chat.postMessage", &buf)
	// body, err := json.Marshal(&payload)
	// if err != nil {
	// 	return err
	// }
	// return c.makeReq("https://slack.com/api/chat.postMessage", bytes.NewReader(body))
}

type slackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// TestConnection tests that you can connect to the slack APi
func (c *SlackClient) TestConnection() error {
	return c.makeReq("https://slack.com/api/api.test", strings.NewReader(`{}`))
}

func (c *SlackClient) makeReq(url string, payload io.Reader) error {
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf(`Bearer %s`, c.Token))
	req.Header.Add("Content-type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("body = %s", body)
	_, err = ioutil.ReadAll(resp.Body)
	return err
}
