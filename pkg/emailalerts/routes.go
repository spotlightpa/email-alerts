package emailalerts

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	// r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.l}))
	r.Use(app.versionMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Post(`/api/add-contact`, app.postAddContact)
	if app.isLambda() {
		r.NotFound(app.notFound)
	} else {
		r.NotFound(http.FileServer(http.Dir("./public")).ServeHTTP)
	}
	return r
}

// common errors
var (
	errNotFound   = withStatus(http.StatusNotFound, fmt.Errorf("not found"))
	errBadRequest = withStatus(http.StatusBadRequest, fmt.Errorf("bad request"))
)

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(r.Context(), w, fmt.Errorf("%q: %w", r.URL, errNotFound))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Printf("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	w.Write(b)
}

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.Printf("start pingErr %q", code)

	app.errorResponse(r.Context(), w,
		withStatus(statusCode, fmt.Errorf("test ping")))
}

func (app *appEnv) postAddContact(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postAddContact")
	email := r.FormValue("email")
	first := r.FormValue("first_name")
	last := r.FormValue("last_name")
	fips := r.FormValue("fips")

	if err := app.addContact(r.Context(), first, last, email, fips); err != nil {
		app.logErr(r.Context(), err)
		http.Redirect(w, r, "/sorry.html", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/thanks.html", http.StatusSeeOther)
}
