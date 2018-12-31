package main

import (
	"log"
	"net/http"

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
	http.HandleFunc("/authorize", AuthorizeHandler)
	http.HandleFunc("/token", func(w http.ResponseWriter, req *http.Request) {
		err := srv.HandleTokenRequest(w, req)
		if err != nil {

		}
	})

	log.Println("Server runs in localhost:9096")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
