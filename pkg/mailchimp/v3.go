package mailchimp

import (
	"context"
	"crypto/md5"
	"errors"
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
		rb: new(requests.Builder).
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
	var (
		resp    PutUserResponse
		errResp ErrorResponse
	)
	err := v3.rb.Clone().
		Pathf("/3.0/lists/%s/members/%s",
			v3.listID, SubscriberHash(req.EmailAddress)).
		Put().
		BodyJSON(&req).
		ToJSON(&resp).
		ErrorJSON(&errResp).
		Fetch(ctx)
	if err != nil {
		if errors.Is(err, requests.ErrInvalidHandled) {
			err = resperr.WithUserMessagef(err, "%s: \n%s",
				errResp.Title, errResp.Detail)
		}
		if errResp.Status == http.StatusBadRequest {
			err = ErrorBadAddress{req.EmailAddress, err}
		} else {
			err = resperr.New(http.StatusBadGateway,
				"problem connecting to MailChimp: %w", err)
		}
	}
	return err
}

type TagAction bool

const (
	AddTag    TagAction = true
	RemoveTag TagAction = false
)

func (v3 V3) UserTags(ctx context.Context, email string, action TagAction, tags ...string) error {
	var req PostTagRequest
	status := "active"
	if action == RemoveTag {
		status = "inactive"
	}
	req.Tags = make([]UserTag, len(tags))
	for i, tag := range tags {
		req.Tags[i] = UserTag{Name: tag, Status: status}
	}
	return v3.rb.Clone().
		Pathf("/3.0/lists/%s/members/%s/tags",
			v3.listID, SubscriberHash(email)).
		BodyJSON(&req).
		CheckStatus(http.StatusNoContent).
		Fetch(ctx)
}

type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

type ErrorBadAddress struct {
	email string
	cause error
}

func (b ErrorBadAddress) Error() string {
	return fmt.Sprintf("bad address: %q", b.email)
}

func (b ErrorBadAddress) Unwrap() error { return b.cause }

func (ErrorBadAddress) StatusCode() int { return http.StatusBadRequest }
