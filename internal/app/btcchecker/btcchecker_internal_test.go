package btcchecker

import (
	"net/http"
	"net/http/httptest"
	"testing"

  	"github.com/stretchr/testify/assert"
)

func TestBTCChecker_HandleUserCreate(t *testing.T) {
	s := New(NewConfig())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/create", nil)

	s.HandleUserCreate().ServeHTTP(rec, req)
	assert.Equal(t, rec.Body.String(), "Create User")
}

func TestBTCChecker_HandleUserLogin(t *testing.T) {
	s := New(NewConfig())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/login", nil)

	s.HandleUserLogin().ServeHTTP(rec, req)
	assert.Equal(t, rec.Body.String(), "Login User")
}

func TestBTCChecker_HandleBTCRate(t *testing.T) {
	s := New(NewConfig())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/btcRate", nil)

	s.HandleBTCRate().ServeHTTP(rec, req)
	assert.Equal(t, rec.Body.String(), "BTC to grivna")
}
