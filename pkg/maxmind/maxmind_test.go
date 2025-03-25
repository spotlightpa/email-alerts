package maxmind_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/email-alerts/pkg/maxmind"
)

func TestMaxmind(t *testing.T) {
	{
		cl := http.Client{}
		cl.Transport = reqtest.ReplayString(`HTTP/2.0 200 OK
Content-Length: 761
Alt-Svc: h3=":443"; ma=86400
Cf-Cache-Status: DYNAMIC
Cf-Ray: 92598e92a8413aea-IAD
Content-Type: application/vnd.maxmind.com-country+json; charset=UTF-8; version=2.1
Date: Mon, 24 Mar 2025 22:20:38 GMT
Server: cloudflare
Set-Cookie: _cfuvid=WZy8vYF3W.sNdVVApFRNobUF4PIdn.mvg3G5ItHzX7g-1742854838206-0.0.1.1-604800000; path=/; domain=.maxmind.com; HttpOnly; Secure; SameSite=None
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload

{"continent":{"code":"NA","geoname_id":6255149,"names":{"en":"North America","es":"Norteamérica","fr":"Amérique du Nord","ja":"北アメリカ","pt-BR":"América do Norte","ru":"Северная Америка","zh-CN":"北美洲","de":"Nordamerika"}},"country":{"iso_code":"US","geoname_id":6252001,"names":{"pt-BR":"EUA","ru":"США","zh-CN":"美国","de":"USA","en":"United States","es":"Estados Unidos","fr":"États Unis","ja":"アメリカ"}},"maxmind":{"queries_remaining":249994},"registered_country":{"iso_code":"US","geoname_id":6252001,"names":{"fr":"États Unis","ja":"アメリカ","pt-BR":"EUA","ru":"США","zh-CN":"美国","de":"USA","en":"United States","es":"Estados Unidos"}},"traits":{"ip_address":"45.63.4.247","network":"45.63.0.0/20"}}`)
		maxcl := maxmind.New("", "", &cl)
		ok, err := maxcl.IPInCountry(t.Context(), "1.2.3.4", "US")
		be.NilErr(t, err)
		be.True(t, ok)
	}
	{
		cl := http.Client{}
		cl.Transport = reqtest.ReplayString(`HTTP/2.0 200 OK
Alt-Svc: h3=":443"; ma=86400
Cf-Cache-Status: DYNAMIC
Cf-Ray: 92598e92a8413aea-IAD
Content-Type: application/vnd.maxmind.com-country+json; charset=UTF-8; version=2.1
Date: Mon, 24 Mar 2025 22:20:38 GMT
Server: cloudflare
Set-Cookie: _cfuvid=WZy8vYF3W.sNdVVApFRNobUF4PIdn.mvg3G5ItHzX7g-1742854838206-0.0.1.1-604800000; path=/; domain=.maxmind.com; HttpOnly; Secure; SameSite=None
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload

{"continent":{"code":"AS","geoname_id":6255147,"names":{"ja":"アジア","pt-BR":"Ásia","ru":"Азия","zh-CN":"亚洲","de":"Asien","en":"Asia","es":"Asia","fr":"Asie"}},"country":{"iso_code":"KR","geoname_id":1835841,"names":{"ru":"Республика Корея","zh-CN":"韩国","de":"Südkorea","en":"South Korea","es":"Corea del Sur","fr":"Corée du Sud","ja":"大韓民国","pt-BR":"Coreia do Sul"}},"maxmind":{"queries_remaining":249993},"registered_country":{"iso_code":"KR","geoname_id":1835841,"names":{"zh-CN":"韩国","de":"Südkorea","en":"South Korea","es":"Corea del Sur","fr":"Corée du Sud","ja":"大韓民国","pt-BR":"Coreia do Sul","ru":"Республика Корея"}},"traits":{"ip_address":"123.45.67.89","network":"123.32.0.0/12"}}`)
		maxcl := maxmind.New("", "", &cl)
		ok, err := maxcl.IPInCountry(t.Context(), "123.45.67.89", "US")
		be.NilErr(t, err)
		be.False(t, ok)
	}
}
