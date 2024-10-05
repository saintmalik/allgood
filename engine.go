package allgood

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/saintmalik/allgood/internal/models"
	"github.com/saintmalik/allgood/internal/notify"
	"github.com/saintmalik/allgood/internal/views"
)

const Version = "0.1.1"

var (
	CheckTypePostgresConnection   CheckType   = "postgres"
	CheckTypeMongoConnection      CheckType   = "mongo"
	CheckTypeSupabaseDBConnection CheckType   = "supabaseDb"
	CheckTypeCPUUsage             CheckType   = "cpuUsage"
	CheckTypeDatabaseConnection   CheckType   = "databaseConnection"
	CheckTypeDatabaseQuery        CheckType   = "databaseQuery"
	CheckTypeRedisConnection      CheckType   = "redisConnection"
	CheckTypeDiskSpace            CheckType   = "diskSpace"
	CheckTypeMemoryUsage          CheckType   = "memoryUsage"
	AvoidDuplicateFor             []CheckType = []CheckType{CheckTypeCPUUsage, CheckTypeDiskSpace, CheckTypeMemoryUsage}
)

// CheckFunc is the function that performs checks
type CheckFunc func() (bool, string)

// CheckType is the category a check belongs to
type CheckType string

// CheckConfig holds all the configurations required
// to setup a check
type CheckConfig struct {
	// Id is the identifier of the created check
	Id          uuid.UUID
	// Type is the category the check to be be created belongs to
	// an Engine may perform checks on e.g. multiple postgres connection
	// checks
	Type        CheckType
	// Name is the name of the check. a default name is set is a check
	// initializer is used and may be overridden in the same check initializer.
	Name        string
	// Enabled is a flag to determine is the check should be executed.
	Enabled     bool
	// HandlerFunc is the function that performs the check
	HandlerFunc CheckFunc
}

// Engine
type Engine struct {
	checks   map[uuid.UUID]CheckConfig
	notifier notify.Notifier
	mu       sync.Mutex
}

// NewEngine creates a new Engine.
// 
// it accepts variable check initializers used to create CheckConfig.
// 
// # Example
// 	import (
// 	"github.com/saintmalik/allgood"
// 	"net/http"
// )
// 	engine := NewEngine(allgood.WithCheckMemoryUsage(90))
// 	http.HandleFunc("/healthcheck", engine.HealthCheckHandler())
// 	http.ListenAndServe(":8080")
func NewEngine(checkInitializers ...CheckInit) *Engine {
	checks := make(map[uuid.UUID]CheckConfig, len(checkInitializers))
	for _, initializer := range checkInitializers {
		config := initializer()
		// TODO: ignore adding check config for times in
		// AvoidDuplicateFor it types already exist in the
		// checks.
		checks[config.Id] = config
	}
	return &Engine{
		checks: checks,
	}
}



// HealthCheckHandler provides the http.HandlerFunc that
// you can bind to your go web app to view the status of
// all the checks.
func (e *Engine) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := e.runChecks()
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
			json.NewEncoder(w).Encode(map[string]any{
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

// AddChecks lets you add more CheckConfig to already existing
// checks
func (e *Engine) AddChecks(checkInitializers ...CheckInit) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, initializer := range checkInitializers {
		config := initializer()
		// TODO: ignore adding check config for times in
		// AvoidDuplicateFor it types already exist in the
		// checks. 
			e.checks[config.Id] = config
	}
}

// EnableCheck lets you disable a check.
func (e *Engine) EnableCheck(id uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if check, exists := e.checks[id]; exists {
		check.Enabled = true
		e.checks[id] = check
	}
}
// DisableCheck lets you disable a check.
func (e *Engine) DisableCheck(id uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if check, exists := e.checks[id]; exists {
		check.Enabled = false
		e.checks[id] = check
	}
}

func (e *Engine) SetNotifier(n notify.Notifier) {
	e.notifier = n
}

func (e *Engine) runChecks() []models.Result {
	var results []models.Result
	for id, check := range e.checks {
		if check.Enabled {
			success, message := check.HandlerFunc()
			results = append(results, models.Result{
				Id:      id.String(),
				Name:    check.Name,
				Success: success,
				Message: message,
			})
		}
	}
	return results
}
