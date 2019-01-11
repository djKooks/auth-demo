package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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

		log.Println("Access token:", globalToken.AccessToken)
		resp, err := http.Get(fmt.Sprintf("%s/userinfo?access_token=%s", AuthServerURL, globalToken.AccessToken))
		if err != nil {
			log.Println("get user information error: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		io.Copy(w, resp.Body)

		// uncomment this part when you need to receive user data from auth server
		/*
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			var responseMap map[string]interface{}

			// get user information with JSON format
			json.Unmarshal(bodyBytes, &responseMap)
			log.Println("get user information with access token: ", responseMap["clientID"])
		*/

	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		globalToken.Expiry = time.Now()
		token, err := config.TokenSource(context.Background(), globalToken).Token()

		if err != nil {
			log.Println("token refreshing err: ", err.Error())
			// TODO:
			// need to check detail, to see error code is `invalid_grant`(refresh token expired) or not,
			// and redirect to root only it is expired
			globalToken = nil
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		log.Println("refreshed token: ", token)
		globalToken = token

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
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
