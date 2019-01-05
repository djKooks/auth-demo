package main

import (
	"net/http"

	"github.com/go-session/session"
)

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		return
	}

	if req.Method == "POST" {
		store.Set("LoggedInUserID", "000000")
		store.Save()

		w.Header().Set("Location", "/auth")
	}

	outputHtml(w, req, "static/login.html")
}

func LoggedHandler(w http.ResponseWriter, req *http.Request) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, ok := store.Get("LoggedInUserID")
	if !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	outputHtml(w, req, "static/auth.html")
}

func UserAuthorizeHandler(w http.ResponseWriter, req *http.Request) (userID string, err error) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		return
	}

	uid, ok := store.Get("LoggedInUserID")

	// if there is no loggedin user id...
	if !ok {
		if req.Form == nil {
			req.ParseForm()
		}

		store.Set("ReturnUri", req.Form)
		store.Save()

		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)

		return
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()

	return
}
