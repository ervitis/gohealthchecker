package gohealthchecker

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
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

	response struct {
		Info       []info                    `json:"info,omitempty"`
		Code       int                       `json:"code"`
		SystemInfo SystemInformationResponse `json:"systemInformation"`
	}

	errInfo struct {
		code    int
		message string
		service string
	}

	// SystemInformationResponse body response struct data
	SystemInformationResponse struct {
		ProcessStatus string `json:"processStatus"`
		ProcessActive bool   `json:"processActive"`
		Pid           uint64 `json:"pid"`
		StartTime     string `json:"startTime"`
		Memory        struct {
			Total     uint64 `json:"total"`
			Free      uint64 `json:"free"`
			Available uint64 `json:"available"`
		} `json:"memory"`
		IpAddress      string `json:"ipAddress"`
		RuntimeVersion string `json:"runtimeVersion"`
		CanAcceptWork  bool   `json:"canAcceptWork"`
	}

	fnNode struct {
		next *fnNode
		e    *errInfo
		fn   Healthfunc
		name string
	}

	// Healthchecker type
	Healthchecker struct {
		mtx        sync.Mutex
		fns        *fnNode
		statusOk   int
		statusKo   int
		nErrors    uint
		systemInfo *SystemInformation
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
	return &Healthchecker{statusKo: statusKo, statusOk: statusOk, nErrors: 0, systemInfo: &SystemInformation{startTime: time.Now()}}
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

	// do check system
	h.recoverFromPanic()
	if err := h.systemInfo.GetSystemInfo(); err != nil {
		h.nErrors++
		panic(err.Error())
	}
}

func (h *Healthchecker) responseSystemFactory() *response {
	r := &response{}
	r.SystemInfo.ProcessActive = h.systemInfo.processActive
	r.SystemInfo.ProcessStatus = h.systemInfo.processStatus
	r.SystemInfo.Memory.Available = h.systemInfo.memory.available
	r.SystemInfo.Memory.Total = h.systemInfo.memory.total
	r.SystemInfo.Memory.Free = h.systemInfo.memory.free
	r.SystemInfo.CanAcceptWork = h.systemInfo.canAcceptWork
	r.SystemInfo.Pid = h.systemInfo.pid
	r.SystemInfo.IpAddress = h.systemInfo.ipAddress.ipAddress
	r.SystemInfo.RuntimeVersion = h.systemInfo.runtimeVersion
	r.SystemInfo.StartTime = h.systemInfo.startTime.Format(time.RFC3339)

	return r
}

func (h *Healthchecker) recoverFromPanic() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = fmt.Fprintf(os.Stdout, "panic at %v", time.Now().Format(time.RFC3339))
		}
	}()
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

func (a *apiHealthCheck) healthCheckerHandler(w http.ResponseWriter, _ *http.Request) {
	a.healthchecker.executeHealthChecker()
	resp := a.healthchecker.responseSystemFactory()

	if a.healthchecker.nErrors == 0 {
		resp.Code = a.healthchecker.statusOk
		b, _ := json.Marshal(resp)
		w.WriteHeader(a.healthchecker.statusOk)
		_, _ = w.Write(b)
		return
	}

	resp.Code = a.healthchecker.statusKo

	c := a.healthchecker.fns
	for c != nil {
		if c.e != nil {
			info := &info{Code: c.e.code, Message: c.e.message, Service: c.e.service}
			resp.Info = append(resp.Info, *info)
		}

		c = c.next
	}
	a.healthchecker.clearError()

	b, err := json.Marshal(resp)
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
