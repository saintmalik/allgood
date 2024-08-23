package allgood

import (
	"github.com/saintmalik/allgood/config"
	"github.com/saintmalik/allgood/engine"
	"github.com/saintmalik/allgood/healthcheck"
)

const Version = "0.1.0"

func NewEngine() *engine.Engine {
    cfg := config.NewConfiguration()
    hc := healthcheck.NewHealthCheck(cfg)
    return engine.NewEngine(hc)
}