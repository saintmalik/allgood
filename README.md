# allgood

![allgood](allgood.png)

```
go get github.com/saintmalik/allgood

import (
	"github.com/saintmalik/allgood"
	"github.com/saintmalik/allgood/checks"
)

func main() {
	engine := allgood.NewEngine()

	redisClient, err := initializeRedis(redisurl)
	if err != nil {
		fmt.Println("Redis initialization failed")
	}

	// mongoClient := // your MongoDB client
	// engine.SetCheck("MongoDB connection", allgood.CheckMongoConnection(mongoClient))

	ref := // your Supabase project ref
	apikey := // your Supabase apikey
	engine.SetCheck("Supabase connection", allgood.CheckSupabaseConnection(ref, apikey))
	// engine.SetCheck("Postgres connection", checks.CheckPostgresConnection(postgresPool))

	engine.SetCheck("Redis connection", checks.CheckRedisConnection(redisClient))

	engine.SetCheck("Disk space usage", checks.CheckDiskSpace(90))
	engine.SetCheck("Memory usage", checks.CheckMemoryUsage(90))
	engine.SetCheck("CPU usage", checks.CheckCPUUsage(90))


	http.HandleFunc("/healthcheck", engine.HealthCheckHandler())
	http.ListenAndServe(":8080", nil)
}
```