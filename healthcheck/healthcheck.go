package healthcheck

import (
    "github.com/saintmalik/allgood/config"
)

type CheckFunc func() (bool, string)

type Result struct {
    Name     string
    Success  bool
    Message  string
}

type HealthCheck struct {
    config *config.Configuration
}

func NewHealthCheck(cfg *config.Configuration) *HealthCheck {
    return &HealthCheck{
        config: cfg,
    }
}

func (hc *HealthCheck) Config() *config.Configuration {
    return hc.config
}

func (hc *HealthCheck) RunChecks() []Result {
    var results []Result
    for name, check := range hc.config.Checks {
        if check.Enabled {
            success, message := check.Func()
            results = append(results, Result{
                Name:    name,
                Success: success,
                Message: message,
            })
        }
    }
    return results
}