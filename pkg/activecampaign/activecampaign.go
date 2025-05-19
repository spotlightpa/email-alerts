// Package activecampaign wraps the Active Campaign API.
package activecampaign

import (
	"context"
	"net/http"

	"github.com/carlmjohnson/requests"
)

func New(host, key string, cl *http.Client) Client {
	return Client{
		requests.
			New().
			Scheme("https").
			Host(host).
			Client(cl).
			Header("Api-Token", key),
	}
}

type Client struct {
	rb *requests.Builder
}

type FieldValue struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type Contact struct {
	Email       string       `json:"email"`
	FirstName   string       `json:"firstName"`
	LastName    string       `json:"lastName"`
	Phone       string       `json:"phone"`
	FieldValues []FieldValue `json:"fieldValues"`
}

func (cl Client) CreateContact(ctx context.Context, contact Contact) error {
	type CreateContact struct {
		Contact Contact `json:"contact"`
	}

	return cl.rb.Clone().
		Path("/api/3/contacts").
		BodyJSON(CreateContact{contact}).
		CheckStatus(201, 422).
		Fetch(ctx)
}

func (cl Client) FindContactByEmail(ctx context.Context, email string) (FindContactResponse, error) {
	var data FindContactResponse
	err := cl.rb.Clone().
		Path("/api/3/contacts").
		Param("email", email).
		ToJSON(&data).
		Fetch(ctx)
	return data, err
}

type FindContactResponse struct {
	Contacts []ContactInfo   `json:"contacts"`
	Meta     FindContactMeta `json:"meta"`
}

type ContactInfo struct {
	ID                  ContactID `json:"id,string"`
	Email               string    `json:"email"`
	Cdate               string    `json:"cdate"`
	Phone               string    `json:"phone"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	Orgid               string    `json:"orgid"`
	SegmentioID         string    `json:"segmentio_id"`
	BouncedHard         string    `json:"bounced_hard"`
	BouncedSoft         string    `json:"bounced_soft"`
	BouncedDate         string    `json:"bounced_date"`
	IP                  string    `json:"ip"`
	Ua                  string    `json:"ua"`
	Hash                string    `json:"hash"`
	SocialdataLastcheck string    `json:"socialdata_lastcheck"`
	EmailLocal          string    `json:"email_local"`
	EmailDomain         string    `json:"email_domain"`
	Sentcnt             string    `json:"sentcnt"`
	RatingTstamp        string    `json:"rating_tstamp"`
	Gravatar            string    `json:"gravatar"`
	Deleted             string    `json:"deleted"`
	Adate               string    `json:"adate"`
	Udate               string    `json:"udate"`
	Edate               string    `json:"edate"`
	ScoreValues         []any     `json:"scoreValues"`
	Organization        any       `json:"organization"`
	AccountContacts     []string  `json:"accountContacts,omitempty"`
}

type FindContactMeta struct {
	Total string `json:"total"`
}

type (
	ListID    int
	ContactID int
	Status    int
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -trimprefix List -type ListID
const (
	ListMaster          ListID = 1
	ListPALocal         ListID = 3
	ListPAPost          ListID = 4
	ListInvestigator    ListID = 5
	ListHowWeCare       ListID = 6
	ListTalkOfTheTown   ListID = 7
	ListPennStateAlerts ListID = 8
	ListBerksCounty     ListID = 9
	ListBreakingNews    ListID = 10
	ListWeekInReview    ListID = 11
	ListEvents          ListID = 13
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -trimprefix Status -type Status
const (
	StatusUnconfirmed  Status = 0
	StatusActive       Status = 1
	StatusUnsubscribed Status = 2
	StatusBounced      Status = 3
)

func (cl Client) AddToList(ctx context.Context, listID ListID, contactID ContactID, status Status) error {
	type ContactList struct {
		List    ListID    `json:"list"`
		Contact ContactID `json:"contact"`
		Status  Status    `json:"status"`
	}
	type AddToList struct {
		ContactList ContactList `json:"contactList"`
	}
	return cl.rb.Clone().
		Path("/api/3/contactLists").
		BodyJSON(AddToList{ContactList: ContactList{
			List:    listID,
			Contact: contactID,
			Status:  status,
		}}).
		Fetch(ctx)
}
const SignUpSourceFieldID = "162"
