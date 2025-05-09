package maxmind

import (
	"context"
	"log"
	"net/http"
	"slices"

	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/resperr/v2"
)

type Client struct {
	accountID, licenseKey string
	cl                    *http.Client
	l                     *log.Logger
}

func New(accountID, licenseKey string, cl *http.Client, l *log.Logger) Client {
	return Client{accountID, licenseKey, cl, l}
}

type Result int

//go:generate go run golang.org/x/tools/cmd/stringer@latest -trimprefix Result -type Result
const (
	ResultFailed Result = iota
	ResultProvisional
	ResultPassed
)

func (mc Client) IPInsights(ctx context.Context, ip string, countrycodes ...string) (Result, error) {
	var resp struct {
		Country struct {
			IsoCode string `json:"iso_code"`
		} `json:"country"`
		Traits struct {
			ConnectionType string `json:"connection_type"`
		} `json:"traits"`
	}
	err := requests.
		URL("https://geoip.maxmind.com/geoip/v2.1/insights/").
		Path(ip).
		Client(mc.cl).
		BasicAuth(mc.accountID, mc.licenseKey).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		mc.l.Printf("maxmind.Client.IPInCountry(%q): got err: %v", ip, err)
		return ResultProvisional, resperr.New(http.StatusBadGateway, "connecting to maxmind: %w", err)
	}
	mc.l.Printf("maxmind.Client.IPInCountry(%q): got: %q-%q",
		ip, resp.Country.IsoCode, resp.Traits.ConnectionType)
	if !slices.Contains(countrycodes, resp.Country.IsoCode) {
		return ResultFailed, nil
	}
	if resp.Traits.ConnectionType == "Corporate" {
		return ResultProvisional, nil
	}
	return ResultPassed, nil
}
