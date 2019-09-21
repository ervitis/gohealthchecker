package gohealthchecker

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealthchecker(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusNotFound)

	if h.statusOk != http.StatusOK && h.statusKo != http.StatusNotFound {
		t.Error("statusOk and statusKo are not correctly set")
	}
}

func TestHealthchecker_Add(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health1 := func() Healthfunc {
		return func() (code int, e error) {
			return 200, nil
		}
	}

	h.Add(health1())

	if h.fns.fn == nil {
		t.Error("not set function for healthchecking")
	}
}

func TestHealthchecker_Add2(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health1 := func() Healthfunc {
		return func() (code int, e error) {
			return 200, nil
		}
	}
	health2 := func() Healthfunc {
		return func() (code int, e error) {
			return 200, nil
		}
	}

	h.Add(health1())
	h.Add(health2())

	if h.fns.next.fn == nil {
		t.Error("error setting two handler functions for healthcheck")
	}
}

func TestHealthchecker_ActivateHealthCheck(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	count := 0

	health1 := func() Healthfunc {
		return func() (code int, e error) {
			count++
			return 200, nil
		}
	}

	health2 := func() Healthfunc {
		return func() (code int, e error) {
			count++
			return 500, errors.New("oh my god")
		}
	}

	h.Add(health1())
	h.Add(health2())

	r := h.ActivateHealthCheck("/healthtest")
	go func() {
		_ = http.ListenAndServe(":8082", r)
	}()

	time.Sleep(2 * time.Second)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://localhost:8082/healthtest", nil))

	if w.Code != http.StatusInternalServerError {
		t.Error("code is not internal server error")
	}

	if count != 2 {
		t.Error("the healthcheckers handlers were not activated")
	}

	count = 0
	h2 := NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health3 := func() Healthfunc {
		return func() (code int, e error) {
			count++
			return 200, nil
		}
	}

	h2.Add(health1())
	h2.Add(health3())

	r = h2.ActivateHealthCheck("/healthtest2")

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://localhost:8082/healthtest2", nil))

	if w.Code != http.StatusOK {
		t.Error("code is not status ok")
	}

	if count != 2 {
		t.Error("the healthcheckers handlers were not activated")
	}
}
