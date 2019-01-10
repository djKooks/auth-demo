package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-session/session"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

func main() {
	manager := manage.NewDefaultManager()
	// token store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// this token configuration is just for test...
	// remove/change this part after test ===================================
	tokenCfg := manage.DefaultAuthorizeCodeTokenCfg
	tokenCfg.AccessTokenExp = time.Minute
	tokenCfg.RefreshTokenExp = time.Hour

	manager.SetAuthorizeCodeTokenCfg(tokenCfg)
	// remove/change this part after test ===================================

	clientStore := store.NewClientStore()
	clientStore.Set("delta-test", &models.Client{
		ID:     "delta-test",
		Secret: "delta-test-secret",
		Domain: "http://localhost:9094",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)
	srv.SetUserAuthorizationHandler(UserAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/auth", LoggedHandler)

	// 1. routed to 'authorize' by client
	// 13. call '/authorize POST' after press 'Allow' button in auth.html
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(nil, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var form url.Values
		v, ok := store.Get("ReturnUri")
		if ok {
			// 14. set form data in 'ReturnUri'
			form = v.(url.Values)
		}
		r.Form = form

		store.Delete("ReturnUri")
		store.Save()

		// 2. find auth request handler which set by 'SetUserAuthorizationHandler'
		// It is set as 'userAuthorizeHandler' here
		err = srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// send user information to client with token
	http.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		tokenInfo, err := srv.ValidationBearerToken(r)
		log.Println("user info token:", tokenInfo)
		if err != nil {
			// if fail on getting token info
			log.Println("fail getting token info:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO: Get user information from client ID, and set in `userData`
		userData := map[string]interface{}{
			"expireRemain": int64(tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
			"clientID":     tokenInfo.GetClientID(),
			"userID":       tokenInfo.GetUserID(),
			"plan":         "trial",
		}

		// jsonData, _ := json.Marshal(userData)

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(userData)

		// uncomment this part when you need to return `userData` to client
		/*
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(jsonData)

			return
		*/
	})

	log.Println("Server is running at 9096 port.")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
