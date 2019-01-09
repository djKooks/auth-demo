package main

import (
	"log"
	"net/http"
	"net/url"

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
		log.Println("Authorize process")
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

	log.Println("Server is running at 9096 port.")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
