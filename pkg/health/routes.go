package health

import (
	"github.com/go-chi/chi"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"net/http"
)

// HealthService represents a simple health check service
// it defines the routes that we expose for readiness and
// liveness probes
type HealthService struct {
	// The database to operate with
	repo Repo
}

// New creates a new HealthService
func New(repo Repo) *HealthService {
	return &HealthService{repo: repo}
}

// Routes returns a router with all routes
// supported by this service
func (s *HealthService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", s.Get)
	return router
}

// Health returns the health status of the
// payments service. This relies on the status of the database, if reachable,
// then a 200 status code is returned, otherwise, we return a 503
// and we rely on the orchestration to supervise us, and shut us down if needed
func (s *HealthService) Get(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	statusMsg := "up"
	if s.repo.Check() != nil {
		statusCode = http.StatusServiceUnavailable
		statusMsg = "down"
	}
	RenderJSON(w, r, statusCode, &Health{
		Status: statusMsg,
	})
}
