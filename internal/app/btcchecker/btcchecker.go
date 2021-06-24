package btcchecker

import (
	"net/http"

	"github.com/dkushche/GoBTCChecker/storage"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func Start(config *Config) error {
	st, err := storage.New(config.StoragePath)
	if err != nil {
		return err
	}

	sessionStore := sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(64)))

	srv, err := NewServer(st, config.LogLevel, sessionStore)
	if err != nil {
		return err
	}

	return http.ListenAndServe(config.BindAddr, srv.router)
}
