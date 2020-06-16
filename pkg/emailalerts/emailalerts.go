package emailalerts

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/carlmjohnson/feed2json"
	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/gateway"
	"github.com/peterbourgon/ff/v3"
)

const AppName = "email-alerts"

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

	app.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	flagext.LoggerVar(fs, app.Logger,
		"verbose", flagext.LogVerbose, "log debug output")

	if err := ff.Parse(fs, args, ff.WithEnvVarPrefix("EMAIL_ALERTS")); err != nil {
		return err
	}
	return nil
}

type appEnv struct {
	port int
	*log.Logger
}

func (app *appEnv) Exec() (err error) {
	app.Println("starting")
	defer func() { app.Println("done") }()

	listener := gateway.ListenAndServe
	portStr := ""
	if app.port != -1 {
		portStr = fmt.Sprintf(":%d", app.port)
		listener = http.ListenAndServe
		http.Handle("/", http.FileServer(http.Dir("./public")))
	}

	http.Handle("/api/feed", feed2json.Handler(
		feed2json.StaticURLInjector("https://news.ycombinator.com/rss"),
		nil, nil, nil, cacheControlMiddleware))

	return listener(portStr, nil)
}

func cacheControlMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=300")
		h.ServeHTTP(w, r)
	})
}
