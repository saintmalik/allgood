package allgood

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shirou/gopsutil/cpu"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthStatus struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Status  string `json:"status"`
}

// CheckInit is a function that creates a CheckConfig
type CheckInit func() CheckConfig

// CheckConfigModifier is a function that let you modify fields of a
// CheckConfig
type CheckConfigModifier func(*CheckConfig)

// CheckConfigModifierOptions are functions that serves as variadic parameters
// that can be passed with CheckInit functions to modify the resulting
// CheckConfig created
type CheckConfigModifierOption func() CheckConfigModifier

// WithCheckName lets you modify the default CheckConfig Name
func WithCheckName(name string) CheckConfigModifierOption {
	return func() CheckConfigModifier {
		return func(config *CheckConfig) {
			config.Name = name
		}
	}
}

// WithCheckStatus lets you modify the Enabled flag of a CheckConfig
func WithCheckStatus(enabled bool) CheckConfigModifierOption {
	return func() CheckConfigModifier {
		return func(config *CheckConfig) {
			config.Enabled = enabled
		}
	}
}

// WithCheckPostgresConnection creates a check initializer which creates a CheckConfig for
// checking postgres connections
func WithCheckPostgresConnection(pool *pgxpool.Pool, options ...CheckConfigModifierOption) CheckInit {

	handlerFunc := func() (bool, string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := pool.Ping(ctx)
		if err != nil {
			return false, "Postgres connection failed: " + err.Error()
		}
		return true, "Postgres connection successful"
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypePostgresConnection,
		Name:        "Postgres Connection",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckMongoConnection creates a check initializer which creates a CheckConfig for
// checking mongo connections
func WithCheckMongoConnection(client *mongo.Client, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := client.Ping(ctx, nil)
		if err != nil {
			return false, "MongoDB connection failed: " + err.Error()
		}
		return true, "MongoDB connection successful"
	}

	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeMongoConnection,
		Name:        "Mongo Connection",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckSupabaseDBConnection creates a check initializer which creates a CheckConfig for
// checking supabase database connections
func WithCheckSupabaseDBConnection(projectRef, secretToken string, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		url := fmt.Sprintf("https://api.supabase.com/v1/projects/%s/health?services=db", projectRef)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, fmt.Sprintf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+secretToken)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Sprintf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, fmt.Sprintf("Failed to read response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Sprintf("Supabase health check failed with status code: %d", resp.StatusCode)
		}

		var healthStatuses []HealthStatus
		if err := json.Unmarshal(body, &healthStatuses); err != nil {
			return false, fmt.Sprintf("Failed to parse health status: %v", err)
		}

		if len(healthStatuses) == 0 {
			return false, "No health status received"
		}

		dbStatus := healthStatuses[0]
		if dbStatus.Name != "db" {
			return false, fmt.Sprintf("Unexpected service name: %s", dbStatus.Name)
		}

		if dbStatus.Healthy {
			return true, fmt.Sprintf("Supabase DB is healthy. Status: %s", dbStatus.Status)
		} else {
			return false, fmt.Sprintf("Supabase DB is not healthy. Status: %s", dbStatus.Status)
		}
	}

	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeSupabaseDBConnection,
		Name:        "Supabase DB Connection",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckCPUUsage creates a check initializer which creates a CheckConfig for
// checking cpu usage
func WithCheckCPUUsage(threshold float64, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			return false, "Failed to get CPU usage: " + err.Error()
		}
		if percent[0] > threshold {
			return false, fmt.Sprintf("CPU usage is %.2f%%, which is above the threshold of %.2f%%", percent[0], threshold)
		}
		return true, fmt.Sprintf("CPU usage is %.2f%%, which is below the threshold of %.2f%%", percent[0], threshold)
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeCPUUsage,
		Name:        "CPU Usage",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckDatabaseConnection creates a check initializer which creates a CheckConfig for
// checking database connections
func WithCheckDatabaseConnection(db *sql.DB, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		err := db.Ping()
		if err != nil {
			return false, "Database connection failed"
		}
		return true, "Database connection successful"
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeDatabaseConnection,
		Name:        "Check Database Connection",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckDatabaseQuery creates a check initializer which creates a CheckConfig for
// checking database queries
func WithCheckDatabaseQuery(db *sql.DB, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		_, err := db.Exec("SELECT 1")
		if err != nil {
			return false, "Database query failed"
		}
		return true, "Database query successful"
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeDatabaseQuery,
		Name:        "Check Database Query",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckRedisConnection creates a check initializer which creates a CheckConfig for
// checking redis connections
func WithCheckRedisConnection(client *redis.Client, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			return false, "Redis connection failed"
		}
		return true, "Redis connection successful"
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeRedisConnection,
		Name:        "Check Redis Connection",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckDiskSpace creates a check initializer which creates a CheckConfig for
// checking disk space
func WithCheckDiskSpace(threshold float64, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		var stat syscall.Statfs_t
		syscall.Statfs("/", &stat)
		total := float64(stat.Blocks) * float64(stat.Bsize)
		free := float64(stat.Bfree) * float64(stat.Bsize)
		used := total - free
		usagePercent := (used / total) * 100

		if usagePercent > threshold {
			return false, fmt.Sprintf("Disk usage is %.2f%%, which is above the threshold of %.2f%%", usagePercent, threshold)
		}
		return true, fmt.Sprintf("Disk usage is %.2f%%, which is below the threshold of %.2f%%", usagePercent, threshold)
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeDiskSpace,
		Name:        "Check Disk Space",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}

// WithCheckMemoryUsage creates a check initializer which creates a CheckConfig for
// checking memory usage
func WithCheckMemoryUsage(threshold float64, options ...CheckConfigModifierOption) CheckInit {
	handlerFunc := func() (bool, string) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		usagePercent := float64(m.Alloc) / float64(m.Sys) * 100

		if usagePercent > threshold {
			return false, fmt.Sprintf("Memory usage is %.2f%%, which is above the threshold of %.2f%%", usagePercent, threshold)
		}
		return true, fmt.Sprintf("Memory usage is %.2f%%, which is below the threshold of %.2f%%", usagePercent, threshold)
	}
	config := CheckConfig{
		Id:          uuid.New(),
		Type:        CheckTypeMemoryUsage,
		Name:        "Check Memory Usage",
		Enabled:     true,
		HandlerFunc: handlerFunc,
	}

	for _, option := range options {
		modifier := option()
		modifier(&config)
	}
	return func() CheckConfig { return config }
}
