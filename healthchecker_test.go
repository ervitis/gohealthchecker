package gohealthchecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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

	h.Add(health1(), "health1")

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

	h.Add(health1(), "health1")
	h.Add(health2(), "health2")

	if h.fns.next.fn == nil {
		t.Error("error setting two handler functions for healthcheck")
	}
}

func TestHealthchecker_ActivateHealthCheck_Router(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusBadRequest)
	r := h.ActivateHealthCheck("healths")

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		ts, _ := route.GetPathTemplate()
		if ts != "/healths" {
			return fmt.Errorf("not correct path")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestHealthchecker_Add3(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusBadRequest)

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
	health3 := func() Healthfunc {
		return func() (code int, e error) {
			return 200, nil
		}
	}

	h.Add(health1(), "health1")
	h.Add(health2(), "health2")
	h.Add(health3(), "health3")

	count := 0
	n := h.fns
	for n != nil {
		count++
		n = n.next
	}
	if count != 3 {
		t.Errorf("I have added 3 functions so count should be 3, not %d", count)
	}
}

func TestHealthchecker_ActivateHealthCheck_WithoutName(t *testing.T) {
	h := NewHealthchecker(http.StatusOK, http.StatusBadRequest)

	health1 := func() Healthfunc {
		return func() (code int, e error) {
			return 500, fmt.Errorf("something happened")
		}
	}

	h.Add(health1())
	r := h.ActivateHealthCheck("health")

	muxRouter := http.NewServeMux()
	muxRouter.Handle("/healthtest", r)

	srv := httptest.NewUnstartedServer(muxRouter)

	l, _ := net.Listen("tcp", "localhost:8082")
	srv.Listener = l

	srv.Start()
	defer srv.Close()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://localhost:8082/health", nil))

	var body map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&body)

	if body == nil {
		t.Errorf("the body should not be nil")
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

	h.Add(health1(), "health1")
	h.Add(health2(), "health2")

	r := h.ActivateHealthCheck("/healthtest")

	mux := http.NewServeMux()
	mux.Handle("/healthtest", r)

	srv := httptest.NewUnstartedServer(mux)

	l, _ := net.Listen("tcp", "localhost:8082")
	srv.Listener = l

	srv.Start()
	defer srv.Close()

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

	h2.Add(health1(), "health1")
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

func TestToString(t *testing.T) {
	if "hello" != toString([]string{"hello"}) {
		t.Error("it should be a string but returned empty")
	}
}

func TestHealthchecker_CustomFunctionName(t *testing.T) {
	funcName := "funcName"
	h := NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health := func() Healthfunc {
		return func() (code int, e error) {
			return 200, nil
		}
	}

	h.Add(health(), funcName)

	if h.fns.name != funcName {
		t.Error("error setting custom function name")
	}
}
