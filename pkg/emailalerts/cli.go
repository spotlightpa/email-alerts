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
	"time"

	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/gateway"
	"github.com/carlmjohnson/versioninfo"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
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
	silent := fs.Bool("silent", false, "don't log debug output")
	sentryDSN := fs.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	getMC := mailchimp.FlagVar(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fs, AppName); err != nil {
		return err
	}
	if *silent {
		app.l.SetOutput(io.Discard)
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
		Release:   versioninfo.Revision,
		Transport: transport,
	})
}

func (app *appEnv) isLambda() bool {
	return app.port == -1
}

func (app *appEnv) Printf(format string, v ...interface{}) {
	app.l.Printf(format, v...)
}
