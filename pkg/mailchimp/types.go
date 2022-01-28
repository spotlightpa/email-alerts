// Package mailchimp has MailChimp API stuff
package mailchimp

import (
	"encoding/json"
	"time"
)

type PutUserRequest struct {
	EmailAddress    string            `json:"email_address"`
	StatusIfNew     string            `json:"status_if_new"`
	Interests       map[string]bool   `json:"interests,omitempty"`
	MergeFields     map[string]string `json:"merge_fields,omitempty"`
	EmailType       string            `json:"email_type,omitempty"`
	IPOpt           string            `json:"ip_opt,omitempty"`
	IPSignup        string            `json:"ip_signup,omitempty"`
	Language        string            `json:"language,omitempty"`
	Status          string            `json:"status,omitempty"`
	TimestampOpt    string            `json:"timestamp_opt,omitempty"`
	TimestampSignup string            `json:"timestamp_signup,omitempty"`
	VIP             bool              `json:"vip,omitempty"`
}

type PutUserResponse struct {
	ID                string                 `json:"id"`
	EmailAddress      string                 `json:"email_address"`
	EmailClient       string                 `json:"email_client"`
	EmailType         string                 `json:"email_type"`
	Interests         map[string]bool        `json:"interests"`
	Language          string                 `json:"language"`
	LastChanged       time.Time              `json:"last_changed"`
	ListID            string                 `json:"list_id"`
	MemberRating      int                    `json:"member_rating"`
	MergeFields       map[string]interface{} `json:"merge_fields"`
	Source            string                 `json:"source"`
	Status            string                 `json:"status"`
	Tags              []Tag                  `json:"tags"`
	TagsCount         int                    `json:"tags_count"`
	OptInAt           time.Time              `json:"timestamp_opt"`
	SignUpAt          json.RawMessage        `json:"timestamp_signup"`
	UniqueEmailID     string                 `json:"unique_email_id"`
	UnsubscribeReason string                 `json:"unsubscribe_reason"`
	VIP               bool                   `json:"vip"`
	WebID             int                    `json:"web_id"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PostTagRequest struct {
	Tags []UserTag `json:"tags"`
}

type UserTag struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
