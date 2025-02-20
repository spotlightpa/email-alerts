// Package turnstile is a client for CloudFlare validation
package turnstile

import (
	"context"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
)

func New(secret string, client *http.Client) Client {
	return Client{
		secret: secret,
		client: client,
	}
}

type Client struct {
	secret string
	client *http.Client
}

func (cl Client) Validate(ctx context.Context, response, remoteIP string) (bool, error) {
	data := struct {
		Secret   string `json:"secret"`
		Response string `json:"response"`
		RemoteIP string `json:"remoteip,omitempty"`
	}{cl.secret, response, remoteIP}
	var res Response
	err := requests.
		URL(`https://challenges.cloudflare.com/turnstile/v0/siteverify`).
		Client(cl.client).
		BodyJSON(data).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return false, err
	}
	return res.Success, nil
}

type Response struct {
	Success       bool      `json:"success"`
	ChallengeTime time.Time `json:"challenge_ts"`
	Hostname      string    `json:"hostname"`
	ErrorCodes    []string  `json:"error-codes"`
	Action        string    `json:"action"`
	CData         string    `json:"cdata"`
}
