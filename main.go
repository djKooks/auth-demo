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

	// store client
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	firstClientStore := store.NewClientStore()
	firstClientStore.Set("11111", &models.Client{
		ID:     "1111",
		Secret: "111111",
		Domain: "",
	})

	manager.MapClientStorage(firstClientStore)

	secondClientStore := store.NewClientStore()
	secondClientStore.Set("22222", &models.Client{
		ID:     "2222",
		Secret: "222222",
		Domain: "",
	})

	manager.MapClientStorage(secondClientStore)

	srv := server.NewServer(server.NewConfig(), manager)
	srv.SetUserAuthorizationHandler(UserAuthorizeHandler)
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Error:", re.Error.Error())
	})

	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/auth", LoggedHandler)
	http.HandleFunc("/authorize", func(w http.ResponseWriter, req *http.Request) {
		store, err := session.Start(nil, w, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var form url.Values
		if v, ok := store.Get("ReturnUri"); ok {
			form = v.(url.Values)
		}
		req.Form = form

		store.Delete("ReturnUri")
		store.Save()

		err = srv.HandleAuthorizeRequest(w, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	http.HandleFunc("/token", func(w http.ResponseWriter, req *http.Request) {
		err := srv.HandleTokenRequest(w, req)
		if err != nil {

		}
	})

	log.Println("Server runs in localhost:9096")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
