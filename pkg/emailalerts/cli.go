// Package emailalerts is a web application for handling newsletter sign ups
package emailalerts

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/carlmjohnson/gateway"
	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/flagx/v2"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/getsentry/sentry-go"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
	"github.com/spotlightpa/email-alerts/pkg/kickbox"
	"github.com/spotlightpa/email-alerts/pkg/maxmind"
	"github.com/spotlightpa/email-alerts/pkg/turnstile"
)

const AppName = "email-alerts"

var (
	DeployURL string = "http://localhost"
)

func CLI(args []string) error {
	var app appEnv
	err := app.ParseArgs(args)
	if err == nil {
		err = app.Exec()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *appEnv) ParseArgs(args []string) error {
	fs := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fs.IntVar(&app.port, "port", -1, "specify a port to use http rather than AWS Lambda")

	app.l = log.New(os.Stderr, AppName+" ", log.LstdFlags)
	fs.BoolFunc("silent", "don't log debug output", func(s string) error {
		v, err := strconv.ParseBool(s)
		if v {
			app.l.SetOutput(io.Discard)
		}
		return err
	})
	kb := fs.String("kickbox-api-key", "", "API `key` for Kickbox")
	sentryDSN := fs.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	acHost := fs.String("active-campaign-host", "", "`host` URL for Active Campaign")
	acKey := fs.String("active-campaign-api-key", "", "API `key` for Active Campaign")
	turnKey := fs.String("turnstile-secret", "", "API `secret` for CloudFlare Turnstile")
	fs.StringVar(&app.signingSecret, "signing-secret", "", "`secret` for signing tokens")
	accountid := fs.String("maxmind-account-id", "", "`account id` with MaxMind")
	licensekey := fs.String("maxmind-license-key", "", "`license key` with MaxMind")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fs, AppName); err != nil {
		return err
	}
	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}
	cl := &http.Client{
		Timeout: 5 * time.Second,
		Transport: requests.LogTransport(nil,
			func(req *http.Request, res *http.Response, err error, duration time.Duration) {
				if err == nil {
					app.l.Printf("req.host=%q res.code=%d res.duration=%v",
						req.URL.Hostname(), res.StatusCode, duration)
				} else {
					app.l.Printf("req.host=%q err=%v res.duration=%v",
						req.URL.Hostname(), err, duration)
				}
			}),
	}
	app.kb = kickbox.New(*kb, app.l, cl)
	app.ac = activecampaign.New(*acHost, *acKey, cl)
	app.tc = turnstile.New(*turnKey, cl)
	app.maxcl = maxmind.New(*accountid, *licensekey, cl)
	return nil
}

type appEnv struct {
	port          int
	signingSecret string
	l             *log.Logger
	kb            *kickbox.Client
	ac            activecampaign.Client
	tc            turnstile.Client
	maxcl         maxmind.Client
}

func (app *appEnv) Exec() (err error) {
	listener := gateway.ListenAndServe
	var portStr string
	if app.isLambda() {
		u, _ := url.Parse(DeployURL)
		if u != nil {
			portStr = u.Hostname()
		}
	} else {
		portStr = fmt.Sprintf(":%d", app.port)
		listener = http.ListenAndServe
	}

	app.Printf("starting on %s", portStr)
	return listener(portStr, app.routes())
}

func (app *appEnv) initSentry(dsn string) error {
	var transport sentry.Transport
	if app.isLambda() {
		app.Printf("setting sentry sync with timeout")
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	}
	if dsn == "" {
		app.Printf("no Sentry DSN")
		return nil
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:       dsn,
		Release:   versioninfo.Revision,
		Transport: transport,
	})
}

func (app *appEnv) isLambda() bool {
	return app.port == -1
}

func (app *appEnv) Printf(format string, v ...any) {
	app.l.Printf(format, v...)
}
