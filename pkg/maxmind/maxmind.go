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
		mc.l.Printf("maxmind.Client.IPInCountry(%q): got err: %v", ip, err)
		return false, resperr.New(http.StatusBadGateway, "connecting to maxmind: %w", err)
	}
	mc.l.Printf("maxmind.Client.IPInCountry(%q): got: %v", ip, resp.Country.IsoCode)
	return slices.Contains(countrycodes, resp.Country.IsoCode), nil
}
