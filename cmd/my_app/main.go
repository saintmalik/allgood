package main

import (
	"net/http"

	"github.com/saintmalik/allgood"
)

func main() {
	engine := allgood.NewEngine(
		allgood.WithCheckDiskSpace(90),
		allgood.WithCheckMemoryUsage(90, allgood.WithCheckName("Jiggy PC Memory usage")),
		allgood.WithCheckMemoryUsage(90),
		allgood.WithCheckCPUUsage(90),
	)

	http.HandleFunc("/healthcheck", engine.HealthCheckHandler())
	http.ListenAndServe(":8080", nil)
}
