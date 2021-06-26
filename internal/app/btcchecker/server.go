package btcchecker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dkushche/GoBTCChecker/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

const (
	sessionName            = "btcchecker"
	contextKey      ctxKey = iota
	ctxKeyRequestID ctxKey = iota
)

type ctxKey int8

type server struct {
	logger       *logrus.Logger
	router       *mux.Router
	storage      *storage.Storage
	sessionStore sessions.Store
}

func NewServer(storage *storage.Storage, log_level string,
	sessionStore sessions.Store) (*server, error) {
	s := &server{
		logger:       logrus.New(),
		router:       mux.NewRouter(),
		storage:      storage,
		sessionStore: sessionStore,
	}

	level, err := logrus.ParseLevel(log_level)
	if err != nil {
		return nil, err
	}
	s.logger.SetLevel(level)

	s.configureRouter()

	return s, nil
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/user/create", s.handleUserCreate()).Methods("POST")
	s.router.HandleFunc("/user/login", s.handleUserLogin()).Methods("POST")

	private := s.router.PathPrefix("/").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc("/btcRate", s.handleBTCRate()).Methods("GET")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()

		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof("completed with code %d(%s) in %v", rw.code, http.StatusText(rw.code), time.Since(start))
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		email, exists := session.Values["user_email"]
		if !exists {
			s.error(w, r, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		_, err = s.storage.Find(email.(string))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey, email)))
	})
}

func (s *server) handleUserCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := s.storage.AddUser(req.Email, req.Password); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, "Success")
	}
}

func (s *server) handleUserLogin() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		if err := s.storage.UserAuth(req.Email, req.Password); err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_email"] = req.Email
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleBTCRate() http.HandlerFunc {
	type response struct {
		Data struct {
			Base     string `json:"base"`
			Currency string `json:"currency"`
			Amount   string `json:"amount"`
		} `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		json_resp, err := http.Get("https://api.coinbase.com/v2/prices/BTC-UAH/buy")

		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errors.New("can't get the price"))
		}

		resp := &response{}
		if err := json.NewDecoder(json_resp.Body).Decode(resp); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if resp.Data.Base == "BTC" && resp.Data.Currency == "UAH" {
			s.respond(w, r, http.StatusOK, resp.Data.Amount)
			return
		}

		s.error(w, r, http.StatusInternalServerError, errors.New("incorrect exchange response"))
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request,
	code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request,
	code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
