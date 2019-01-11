package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// authentication server url
const AuthServerURL = "http://localhost:9096"

var (
	config = oauth2.Config{
		ClientID:     "delta-test",
		ClientSecret: "delta-test-secret",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:9094/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthServerURL + "/authorize",
			TokenURL: AuthServerURL + "/token",
		},
	}
	globalToken *oauth2.Token // this is token kept by client...it will be used to get user information from auth server
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := config.AuthCodeURL("delta-auth")
		log.Println("handle root: " + u)
		http.Redirect(w, r, u, http.StatusFound)
	})

	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.Form.Get("state")
		if state != "delta-auth" {
			http.Error(w, "State invalid", http.StatusBadRequest)
			return
		}
		code := r.Form.Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		globalToken = token

		// Authentication success. Go to main page
		w.Header().Set("Location", "/main")
		w.WriteHeader(http.StatusFound)
		return
	})

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		outputHTML(w, r, "static/client.html")
	})

	// get user information from auth server
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			// if there is no token in client side, request to auth server
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		resp, err := http.Get(fmt.Sprintf("%s/userinfo?access_token=%s", AuthServerURL, globalToken.AccessToken))
		if err != nil {
			log.Println("get user information error: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var responseMap map[string]interface{}

		// get user information with JSON format
		json.Unmarshal(bodyBytes, &responseMap)
		log.Println("get user information with access token: ", responseMap["clientID"])

	})

	log.Println("Client is running at 9094 port.")
	log.Fatal(http.ListenAndServe(":9094", nil))
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}
