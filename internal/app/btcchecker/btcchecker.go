package btcchecker

import (
	"fmt"
	"net/http"
	
	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type BTCChecker struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
}

func New(config *Config) *BTCChecker {
	return &BTCChecker {
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *BTCChecker) Start() error {
	if err := s.ConfigureLogger(); err != nil {
		return err
	}
	s.ConfigureRouter()

	s.logger.Info("starting server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *BTCChecker) ConfigureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *BTCChecker) ConfigureRouter() {
	s.router.HandleFunc("/user/create", s.HandleUserCreate())
	s.router.HandleFunc("/user/login", s.HandleUserLogin())
	s.router.HandleFunc("/btcRate", s.HandleBTCRate())
}

func (s *BTCChecker) HandleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Create User")
	}
}

func (s *BTCChecker) HandleUserLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Login User")
	}
}
func (s *BTCChecker) HandleBTCRate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "BTC to grivna")
	}
}
