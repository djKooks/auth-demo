package main

import (
	"log"
	"net/http"

	"github.com/go-session/session"
)

// 6. Go to login page
func LoginHandler(w http.ResponseWriter, req *http.Request) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Method == "POST" {
		// 8. after press login button, store userID
		name := req.PostFormValue("username")
		pw := req.PostFormValue("password")

		// TODO: need to check name and password by comparing user data from Database
		if name == "2222" && pw == "222222" {
			store.Set("LoggedInUserID", name)
			store.Save()

			// 9. move to 'LoggedHandler'
			w.Header().Set("Location", "/authorize")
			w.WriteHeader(http.StatusFound)
			return
		} else {
			log.Println("invalid name or password")
		}

	}

	// 7. show login page
	outputHtml(w, req, "static/login.html")
}

//10. logged-in handler
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

// 3, 14. handle auth request
func UserAuthorizeHandler(w http.ResponseWriter, req *http.Request) (userID string, err error) {
	store, err := session.Start(nil, w, req)
	if err != nil {
		return
	}

	uid, ok := store.Get("LoggedInUserID")

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
	store.Delete("LoggedInUserID")
	store.Save()

	return
}
