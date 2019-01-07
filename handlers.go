package main

import (
	"log"
	"net/http"

	"github.com/go-session/session"
)

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Method == "POST" {
		// 8. after press login button, store userID
		name := req.PostFormValue("username")
		store.Set("LoggedInUserID", name)
		store.Save()

		// 9. move to 'authHandler'
		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}

	// 7. show login page
	log.Println("Show login page")

	outputHtml(w, req, "static/login.html")
}

func LoggedHandler(w http.ResponseWriter, req *http.Request) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 11. go to login handler again if there is no 'LoggedInUserID' in store...but it will ignore in normal case
	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	// 12. Show auth page
	outputHtml(w, req, "static/auth.html")
}

func UserAuthorizeHandler(w http.ResponseWriter, req *http.Request) (userID string, err error) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		return
	}

	uid, ok := store.Get("LoggedInUserID")
	log.Println("uid: ", uid)

	if !ok {
		// 4. There will be no 'LoggedInUserID' in store first time
		if req.Form == nil {
			req.ParseForm()
		}

		store.Set("ReturnUri", req.Form)
		store.Save()

		// 5. move to login page
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid.(string)
	// store.Delete("LoggedInUserID")
	// store.Save()

	log.Println("direct login : ", userID)
	return
}
