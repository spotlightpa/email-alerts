package mailchimp

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
)

func FlagVar(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mc-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mc-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

type V3 struct {
	rb     *requests.Builder
	listID string
}

func NewV3(apiKey, listID string, c *http.Client) V3 {
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	var datacenter string
	if n := strings.LastIndex(apiKey, "-"); n != -1 {
		datacenter = apiKey[n+1:]
	}
	return V3{
		rb: requests.URL("").
			Client(c).
			BasicAuth("", apiKey).
			Hostf("%s.api.mailchimp.com", datacenter),
		listID: listID,
	}
}

func SubscriberHash(email string) string {
	email = emailx.Normalize(email)
	return fmt.Sprintf("%x", md5.Sum([]byte(email)))
}

func (v3 V3) PutUser(ctx context.Context, req *PutUserRequest) error {
	var resp PutUserResponse
	err := v3.rb.Clone().
		Pathf("/3.0/lists/%s/members/%s",
			v3.listID, SubscriberHash(req.EmailAddress)).
		Put().
		BodyJSON(&req).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		if requests.HasStatusErr(err, http.StatusBadRequest) {
			err = resperr.WithUserMessagef(
				resperr.New(
					http.StatusBadRequest, "bad address %q", req.EmailAddress),
				"Server rejected email address %q",
				req.EmailAddress,
			)
		} else {
			err = resperr.New(http.StatusBadGateway,
				"problem connecting to MailChimp: %w", err)
		}
	}
	return err
}
