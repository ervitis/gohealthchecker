package main

import (
	"errors"
	"github.com/ervitis/gohealthchecker"
	"net/http"
	"time"
)

func checkGithub() gohealthchecker.Healthfunc {
	const myUrl = "https://api.github.com/usrs/ervitis"

	return func() (code int, e error) {
		req, err := http.NewRequest(http.MethodGet, myUrl, nil)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		client := http.Client{Timeout: 10*time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if resp.StatusCode != http.StatusOK {
			return resp.StatusCode, errors.New(resp.Status)
		}
		return http.StatusOK, nil
	}
}

func main() {
	health := gohealthchecker.NewHealthchecker(http.StatusOK, http.StatusInternalServerError)

	health.Add(checkGithub())

	panic(http.ListenAndServe(":8085", health.ActivateHealthCheck("health")))
}