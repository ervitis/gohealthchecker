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
		healthchecker *Healthchecker
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

	fnNode struct {
		next *fnNode
		e    *errInfo
		fn   Healthfunc
		name string
	}

	// Healthchecker type
	Healthchecker struct {
		mtx      sync.Mutex
		fns      *fnNode
		statusOk int
		statusKo int
		nErrors  uint
		// TODO add logger
	}

	// Healthfunc is used to define the health check functions
	// Example:
	// func checkPort() gohealthchecker.Healthfunc {
	//	return func() (code int, e error) {
	//		conn, err := net.Dial("tcp", ":8185")
	//		if err != nil {
	//			return http.StatusInternalServerError, err
	//		}
	//
	//		_ = conn.Close()
	//		return http.StatusOK, nil
	//	}
	//}
	Healthfunc func() (code int, e error)
)

// NewHealthchecker Constructor
// Pass the http statuses when it's ok or ko to tell if your microservice is not ready
func NewHealthchecker(statusOk, statusKo int) *Healthchecker {
	return &Healthchecker{statusKo: statusKo, statusOk: statusOk, nErrors: 0}
}

func (h *Healthchecker) executeHealthChecker() {
	c := h.fns

	h.mtx.Lock()
	defer h.mtx.Unlock()

	for c != nil {
		if code, err := c.fn(); err != nil {
			if c.name == "" {
				nm := runtime.FuncForPC(reflect.ValueOf(c.fn).Pointer()).Name()
				svc := strings.Split(nm, ".")
				if len(svc) > 1 {
					c.name = svc[1]
				}
			}

			c.e = &errInfo{message: err.Error(), code: code, service: c.name}
			h.nErrors++
		}
		c = c.next
	}
}

func toString(ts []string) string {
	if len(ts) > 0 {
		return ts[0]
	}
	return ""
}

// Add appends to the healthchecker instance a function of type Healthfunc and an optional
// name for the service which it prints the JSON result
func (h *Healthchecker) Add(healthfunc Healthfunc, nameFunction ...string) {
	nf := toString(nameFunction)
	if h.fns == nil {
		h.fns = &fnNode{fn: healthfunc, name: nf}
		return
	}

	c := h.fns
	for c.next != nil {
		c = c.next
	}
	c.next = &fnNode{fn: healthfunc, name: nf}
}

func (h *Healthchecker) clearError() {
	defer func() {
		h.nErrors = 0
	}()
}

func (a *apiHealthCheck) healthCheckerHandler(w http.ResponseWriter, r *http.Request) {
	a.healthchecker.executeHealthChecker()
	if a.healthchecker.nErrors == 0 {
		w.WriteHeader(a.healthchecker.statusOk)
		return
	}

	responseError := &responseError{Code: a.healthchecker.statusKo}

	c := a.healthchecker.fns
	for c != nil {
		if c.e != nil {
			info := &info{Code: c.e.code, Message: c.e.message, Service: c.e.service}
			responseError.Info = append(responseError.Info, *info)
		}

		c = c.next
	}
	a.healthchecker.clearError()

	b, err := json.Marshal(responseError)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(a.healthchecker.statusKo)
	_, _ = w.Write(b)
}

// ActivateHealthCheck returns a Router used for the handler
// it needs a routePath to setup the url of the health check
func (h *Healthchecker) ActivateHealthCheck(routePath string) *mux.Router {
	r := mux.NewRouter()

	api := &apiHealthCheck{healthchecker: h}

	if string(routePath[0]) != "/" {
		routePath = "/" + routePath
	}

	r.HandleFunc(routePath, api.healthCheckerHandler).Methods(http.MethodGet)
	return r
}
