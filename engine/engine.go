package engine

import (
	"encoding/json"
	"net/http"

	"github.com/saintmalik/allgood/config"
	"github.com/saintmalik/allgood/healthcheck"
	"github.com/saintmalik/allgood/notify"
	"github.com/saintmalik/allgood/views"
)

type Engine struct {
	healthCheck *healthcheck.HealthCheck
	notifier    notify.Notifier
	config      *config.Configuration
}

func NewEngine(hc *healthcheck.HealthCheck) *Engine {
	return &Engine{
		healthCheck: hc,
		config:      hc.Config(),
	}
}

func (e *Engine) SetNotifier(n notify.Notifier) {
	e.notifier = n
}

func (e *Engine) SetCheck(name string, check healthcheck.CheckFunc) {
	e.config.AddCheck(name, check)
}

func (e *Engine) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := e.healthCheck.RunChecks()
		status := "ok"
		statusCode := http.StatusOK

		for _, result := range results {
			if !result.Success {
				status = "error"
				statusCode = http.StatusServiceUnavailable
				break
			}
		}

		switch r.Header.Get("Accept") {
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": status,
				"checks": results,
			})
		default:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(statusCode)
			err := views.HealthCheckPage(results, status).Render(r.Context(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}