package checks

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
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shirou/gopsutil/cpu"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthStatus struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Status  string `json:"status"`
}

func CheckPostgresConnection(pool *pgxpool.Pool) func() (bool, string) {
	return func() (bool, string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := pool.Ping(ctx)
		if err != nil {
			return false, "Postgres connection failed: " + err.Error()
		}
		return true, "Postgres connection successful"
	}
}
func CheckMongoConnection(client *mongo.Client) func() (bool, string) {
	return func() (bool, string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := client.Ping(ctx, nil)
		if err != nil {
			return false, "MongoDB connection failed: " + err.Error()
		}
		return true, "MongoDB connection successful"
	}
}

func CheckSupabaseDBConnection(projectRef, secretToken string) func() (bool, string) {
	return func() (bool, string) {
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
}

func CheckCPUUsage(threshold float64) func() (bool, string) {
	return func() (bool, string) {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			return false, "Failed to get CPU usage: " + err.Error()
		}
		if percent[0] > threshold {
			return false, fmt.Sprintf("CPU usage is %.2f%%, which is above the threshold of %.2f%%", percent[0], threshold)
		}
		return true, fmt.Sprintf("CPU usage is %.2f%%, which is below the threshold of %.2f%%", percent[0], threshold)
	}
}

func CheckDatabaseConnection(db *sql.DB) func() (bool, string) {
	return func() (bool, string) {
		err := db.Ping()
		if err != nil {
			return false, "Database connection failed"
		}
		return true, "Database connection successful"
	}
}

func CheckDatabaseQuery(db *sql.DB) func() (bool, string) {
	return func() (bool, string) {
		_, err := db.Exec("SELECT 1")
		if err != nil {
			return false, "Database query failed"
		}
		return true, "Database query successful"
	}
}

func CheckRedisConnection(client *redis.Client) func() (bool, string) {
	return func() (bool, string) {
		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			return false, "Redis connection failed"
		}
		return true, "Redis connection successful"
	}
}

func CheckDiskSpace(threshold float64) func() (bool, string) {
	return func() (bool, string) {
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
}

func CheckMemoryUsage(threshold float64) func() (bool, string) {
	return func() (bool, string) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		usagePercent := float64(m.Alloc) / float64(m.Sys) * 100

		if usagePercent > threshold {
			return false, fmt.Sprintf("Memory usage is %.2f%%, which is above the threshold of %.2f%%", usagePercent, threshold)
		}
		return true, fmt.Sprintf("Memory usage is %.2f%%, which is below the threshold of %.2f%%", usagePercent, threshold)
	}
}
