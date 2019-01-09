package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	config = oauth2.Config{
		ClientID:     "delta-test",
		ClientSecret: "delta-test-secret",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:9094/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:9096/authorize",
			TokenURL: "http://localhost:9096/token",
		},
	}
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: need to generate random string for this
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

		log.Println("code 1: ", code)

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.Client(context.Background(), token)
		log.Println("client: ", client)
		/*
			userInfoResp, err := client.Get(UserInfoAPIEndpoint)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer userInfoResp.Body.Close()
			userInfo, err := ioutil.ReadAll(userInfoResp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var authUser User
			json.Unmarshal(userInfo, &authUser)
		*/
		// e := json.NewEncoder(w)
		// e.SetIndent("", "  ")
		// e.Encode(*token)
		log.Println("redirect: ", w.Header())
		w.Header().Set("Location", "/main")
		w.WriteHeader(http.StatusFound)
		return
	})

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		outputHTML(w, r, "static/client.html")
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
