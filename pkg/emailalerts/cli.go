package emailalerts

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/gateway"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
	"github.com/spotlightpa/email-alerts/pkg/sendgrid"
)

const AppName = "email-alerts"

var (
	BuildVersion string = "Development"
	DeployURL    string = "http://localhost"
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

	app.l = log.New(nil, AppName+" ", log.LstdFlags)
	flagext.LoggerVar(fs, app.l, "silent", flagext.LogSilent, "don't log debug output")
	sentryDSN := fs.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	app.sg = sendgrid.NewMockClient(app.l)
	flagext.Callback(fs, "sendgrid-token", "", "`token` for SendGrid API",
		func(token string) error {
			app.sg = sendgrid.NewClient(token)
			return nil
		})
	getMC := mailchimp.FlagVar(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fs, AppName); err != nil {
		return err
	}
	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}
	app.mc = getMC(&http.Client{Timeout: 5 * time.Second})
	return nil
}

type appEnv struct {
	port int
	l    *log.Logger
	sg   *http.Client
	mc   mailchimp.V3
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
	routes := sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         !app.isLambda(),
		}).
		Handle(app.routes())

	app.Printf("starting on %s", portStr)
	return listener(portStr, routes)
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
		Release:   BuildVersion,
		Transport: transport,
	})
}

func (app *appEnv) isLambda() bool {
	return app.port == -1
}

func (app *appEnv) Printf(format string, v ...interface{}) {
	app.l.Printf(format, v...)
}
