package kickbox

import (
	"context"
	"log"
	"slices"

	"github.com/carlmjohnson/requests"
)

type Client struct {
	apiKey string
	l      *log.Logger
}

func New(apiKey string, l *log.Logger) *Client {
	return &Client{apiKey, l}
}

func (c *Client) Verify(ctx context.Context, email string) bool {
	if c.apiKey == "" {
		c.l.Print("kickbox: warning, no API key set")
		return true
	}
	var obj response
	err := requests.
		URL("https://api.kickbox.com/v2/verify").
		Param("apikey", c.apiKey).
		Param("email", email).
		ToJSON(&obj).
		Fetch(ctx)
	if err != nil {
		c.l.Printf("bad response from kickbox: err=%v", err)
		return true
	}
	c.l.Printf("kickbox: email=%q result=%q", email, obj.Result)
	return slices.Contains([]string{"deliverable", "unknown"}, obj.Result)
}

// https://docs.kickbox.com/docs/single-verification-api
type response struct {
	// result string - The verification result: deliverable, undeliverable, risky, unknown
	Result     string  `json:"result"`
	Reason     string  `json:"reason"`
	Role       bool    `json:"role"`
	Free       bool    `json:"free"`
	Disposable bool    `json:"disposable"`
	AcceptAll  bool    `json:"accept_all"`
	DidYouMean string  `json:"did_you_mean"`
	Sendex     float64 `json:"sendex"`
	Email      string  `json:"email"`
	User       string  `json:"user"`
	Domain     string  `json:"domain"`
	Success    bool    `json:"success"`
	Message    any     `json:"message"`
}
