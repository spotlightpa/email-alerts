package maxmind

import (
	"context"
	"net/http"
	"slices"

	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/resperr/v2"
)

type Client struct {
	accountID, licenseKey string
	cl                    *http.Client
}

func New(accountID, licenseKey string, cl *http.Client) Client {
	return Client{accountID, licenseKey, cl}
}

func (mc Client) IPInCountry(ctx context.Context, ip string, countrycodes ...string) (bool, error) {
	var resp struct {
		Country struct {
			IsoCode string `json:"iso_code"`
		} `json:"country"`
	}
	err := requests.
		URL("https://geoip.maxmind.com/geoip/v2.1/country/").
		Path(ip).
		Client(mc.cl).
		BasicAuth(mc.accountID, mc.licenseKey).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		return false, resperr.New(http.StatusBadGateway, "connecting to maxmind: %w", err)
	}

	return slices.Contains(countrycodes, resp.Country.IsoCode), nil
}
