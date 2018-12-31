package main

import (
	"net/http"

	"github.com/go-session/session"
)

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
