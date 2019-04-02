// admin contains extra http endpoints to perform administrative
// operations on the database. These can be enabled/disabled via the
// --admin command line flag
package admin

import (
	"github.com/go-chi/chi"
	. "github.com/pedro-gutierrez/form3/pkg/util"
	"net/http"
)

// Admin represents an admin service
type AdminService struct {
	// The database to operate with
	repo Repo
}

// New creates a new PaymentsService
func New(repo Repo) *AdminService {
	return &AdminService{repo: repo}
}

// Routes returns a router with all routes
// supported by this service
func (s *AdminService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Route("/repo", func(r chi.Router) {
		r.Delete("/", s.DeleteRepo)
		r.Get("/", s.GetRepo)
	})
	return router
}

// DeleteRepo deletes all data from the repo
func (s *AdminService) DeleteRepo(w http.ResponseWriter, r *http.Request) {
	err := s.repo.DeleteAll()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}
	RenderNoContent(w, r)
}

// GetRepo gets basic info from the repo and exposes it over http
func (s *AdminService) GetRepo(w http.ResponseWriter, r *http.Request) {
	info, err := s.repo.Info()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}
	RenderJSON(w, r, http.StatusOK, info)
}
