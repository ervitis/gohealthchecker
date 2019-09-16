package gohealthchecker

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type (
	apiHealthCheck struct {
		healthchecker
	}

	info struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Service string `json:"service"`
	}

	responseError struct {
		Info []info `json:"info"`
		Code int    `json:"code"`
	}

	errInfo struct {
		code    int
		message string
		service string
	}

	FnNode struct {
		next *FnNode
		e    *errInfo
		fn   Healthfunc
	}

	healthchecker struct {
		mtx      sync.Mutex
		fns      *FnNode
		statusOk int
		statusKo int
		nErrors  uint
		// TODO add logger
	}

	Healthfunc func() (code int, e error)
)

func NewHealthchecker(statusOk, statusKo int) *healthchecker {
	return &healthchecker{statusKo: statusKo, statusOk: statusOk, nErrors: 0}
}

func (h *healthchecker) executeHealthChecker() {
	c := h.fns

	h.mtx.Lock()
	defer h.mtx.Unlock()

	for c != nil {
		if code, err := c.fn(); err != nil {
			nm := runtime.FuncForPC(reflect.ValueOf(c.fn).Pointer()).Name()
			svc := strings.Split(nm, ".")
			if len(svc) > 1 {
				nm = svc[1]
			}
			c.e = &errInfo{message: err.Error(), code: code, service: nm}
			h.nErrors++
		}
		c = c.next
	}
}

func (h *healthchecker) Add(healthfunc Healthfunc) {
	if h.fns == nil {
		h.fns = &FnNode{fn: healthfunc}
		return
	}

	c := h.fns.next
	for c.next != nil {
		c = c.next
	}
	c.next = &FnNode{fn: healthfunc}
}

func (h *healthchecker) clearError() {
	defer func() {
		h.nErrors = 0
	}()
}

func (a *apiHealthCheck) healthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	a.executeHealthChecker()
	if a.nErrors == 0 {
		w.WriteHeader(a.statusOk)
		return
	}

	responseError := &responseError{Code: a.statusKo}

	c := a.fns
	for c != nil {
		info := info{Code: c.e.code, Message: c.e.message, Service: c.e.service}
		responseError.Info = append(responseError.Info, info)

		c = c.next
	}
	a.healthchecker.clearError()

	b, err := json.Marshal(responseError)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(a.statusKo)
	_, _ = w.Write(b)
}

func (h *healthchecker) ActivateHealthCheck(routePath string) *mux.Router {
	r := mux.NewRouter()

	api := &apiHealthCheck{healthchecker: *h}

	if string(routePath[0]) != "/" {
		routePath = "/" + routePath
	}

	r.HandleFunc(routePath, api.healthCheckerHandler).Methods(http.MethodGet)
	return r
}
