package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
)

type App struct {
	Router      *mux.Router
	Middlewares *Middleware
	config      *Env
}

type shortenReq struct {
	URL                 string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

type shortLinkResp struct {
	ShortLink string `json:"shortLink"`
}

func (a *App) Initialize(e *Env) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	m := alice.New(a.Middlewares.LoggingHandler, a.Middlewares.RecoverHandler)
	a.Router.Handle("/api/shorten", m.ThenFunc(a.createShortLink)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.getShortLinkInfo)).Methods("GET")
	a.Router.Handle("/{shortLink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(a.redirect)).Methods("GET")
}

func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}

	if err := validator.Validate(req); err != nil {
		responseWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", req)})
		return
	}
	defer r.Body.Close()
	fmt.Printf("%v\n", req)
	s, err := a.config.S.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		responseWithError(w, err)
	}
	responseWithJSON(w, http.StatusCreated, shortLinkResp{s})
}

func responseWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		responseWithJSON(w, e.Status(), e.Error())
	default:
		responseWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}

func (a *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query()
	s := val.Get("shortLink")
	fmt.Printf("%s\n", s)

	d, err := a.config.S.ShortLinkInfo(s)
	if err != nil {
		responseWithError(w, err)
	}
	responseWithJSON(w, http.StatusCreated, d)

}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("%s\n", vars)

	d, err := a.config.S.UnShorten(vars["shortLink"])
	if err != nil {
		responseWithError(w, err)
	}
	http.Redirect(w, r, d, http.StatusTemporaryRedirect)
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
