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

	private := s.router.NewRoute().Subrouter()
	private.Use(s.setRequestID)
	private.Use(s.logRequest)
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
		Status bool `json:"status"`
		BtcUah struct {
			Sell          string `json:"sell"`
			CurrencyTrade string `json:"currency_trade"`
			BuyUsd        string `json:"buy_usd"`
			Buy           string `json:"buy"`
			Last          string `json:"last"`
			Updated       int    `json:"updated"`
			Vol           string `json:"vol"`
			SellUsd       string `json:"sell_usd"`
			LastUsd       string `json:"last_usd"`
			CurrencyBase  string `json:"currency_base"`
			VolCur        string `json:"vol_cur"`
			High          string `json:"high"`
			Low           string `json:"low"`
			VolCurUsd     string `json:"vol_cur_usd"`
			Avg           string `json:"avg"`
			UsdRate       string `json:"usd_rate"`
		} `json:"btc_uah"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		json_resp, err := http.Get("https://btc-trade.com.ua/api/ticker/btc_uah")
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errors.New("can't get the price"))
		}

		resp := &response{}
		if err := json.NewDecoder(json_resp.Body).Decode(resp); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if resp.Status {
			s.respond(w, r, http.StatusOK, resp.BtcUah.Sell)
			return
		}

		s.error(w, r, http.StatusInternalServerError, errors.New("incorrect status"))
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
