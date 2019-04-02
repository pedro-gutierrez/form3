package main

import (
	"flag"
	"fmt"
	"github.com/766b/chi-prometheus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/pedro-gutierrez/form3/pkg/admin"
	"github.com/pedro-gutierrez/form3/pkg/health"
	"github.com/pedro-gutierrez/form3/pkg/logger"
	"github.com/pedro-gutierrez/form3/pkg/payments"
	"github.com/pedro-gutierrez/form3/pkg/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	listen         *string
	limit          *string
	httpLogs       *bool
	compress       *bool
	metrics        *bool
	repoDriver     *string
	repoUri        *string
	repoMigrations *string
	enableCors     *bool
	timeout        *int
	adminRoutes    *bool
	profiling      *bool
	apiVersion     *string
	externalUrl    *string
)

func init() {
	listen = flag.String("listen", ":8080", "the http interface to listen at")
	limit = flag.String("limit", "", "rate limit (eg. 5-S for 5 reqs/second)")
	httpLogs = flag.Bool("http-logs", false, "enable http logs")
	compress = flag.Bool("compress", false, "gzip responses")
	metrics = flag.Bool("metrics", false, "expose prometheus metrics")
	enableCors = flag.Bool("cors", false, "enable cors")
	timeout = flag.Int("timeout", 60, "request timeout")
	repoDriver = flag.String("repo", "sqlite3", "type of persistence repository to use, eg. sqlite3, postgres")
	repoUri = flag.String("repo-uri", "", "repo specific connection string")
	repoMigrations = flag.String("repo-migrations", "", "path to database migrations")
	adminRoutes = flag.Bool("admin", false, "enable admin endpoints")
	profiling = flag.Bool("profiling", false, "enable profiling")
	apiVersion = flag.String("api-version", "v1", "api version to expose our services at")
	externalUrl = flag.String("external-url", "http://localhost:8080", "url to access our microservice from the outside")
}

// Main entry point to the program. Connects to the database, configures
// all required routes, and starts listening for connections
func main() {
	flag.Parse()

	// compute the baseUrl as the externalUrl + apiVersion configured in this server
	baseUrl := fmt.Sprintf("%s/%s", *externalUrl, *apiVersion)

	// Setup our persistence. We do this first, since we want to exit
	// the program, in case the database is not available
	repo, err := util.NewRepo(util.RepoConfig{
		Driver:     *repoDriver,
		Uri:        *repoUri,
		Migrations: *repoMigrations,
	})

	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not setup repo"))
	}

	// Make sure we close the database on exit
	defer repo.Close()

	// Stop here if the repo is not ready
	if err := repo.Check(); err != nil {
		log.Fatal(errors.Wrap(err, "Could connect to the repo"))
	}

	router := chi.NewRouter()

	// Enable default middleware. Please move the ones you'd wish
	// to make optional further down.
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Timeout(time.Duration(*timeout)*time.Second),
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.RequestID,
		middleware.RealIP,
		middleware.AllowContentType("application/json", "text/plain"),
		middleware.NoCache,
	)

	// Maybe turn on Prometheus metrics
	if *metrics {
		router.Use(chiprometheus.NewMiddleware("payments"))
	}

	// Use our own structured logger middleware
	router.Use(logger.NewHttpLogger())

	// Maybe GZIP http responses
	if *compress {
		router.Use(middleware.DefaultCompress)
	}

	// Maybe set a rate limit, so that we keep things
	// controlled under high load. A 429 status code will
	// be returned once we reach the configured threshold.
	// For simplicity, we use the memory store but there are other
	// options (eg. Redis)
	if *limit != "" {
		rate, err := limiter.NewRateFromFormatted(*limit)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error setting rate limit"))
		}
		store := memory.NewStore()
		router.Use(stdlib.NewMiddleware(limiter.New(store, rate)).Handler)
	}

	// Set some CORS options, if enabled. I guess the options here could
	// be further customised or made configurable
	if *enableCors {
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		router.Use(cors.Handler)
	}

	// If metrics are enabled, then configure the http
	// handler. This has to be done after all middleware have been
	// set
	if *metrics {
		router.Mount("/metrics", prometheus.Handler())
	}

	// If profiling is enabled, enable net/http/pprof middleware
	if *profiling {
		router.Mount("/profiling", middleware.Profiler())
	}

	// Mount health probe
	router.Mount("/health", health.New(repo).Routes())

	// Admin features
	// (useful for testing, for example, but use with care in a production
	// environment
	if *adminRoutes {
		router.Route("/admin", func(adminRouter chi.Router) {
			adminRouter.Mount("/", admin.New(repo).Routes())
		})
	}

	// mount application logic
	router.Route("/v1", func(v1Router chi.Router) {

		// payments api
		v1Router.Mount("/", payments.New(repo, baseUrl).Routes())

		// more endpoints here...
	})

	// Print all routes mounted
	if err := chi.Walk(router, func(method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		logger.Info("Mounted route", &RouteInfo{Method: method, Path: route})
		return nil
	}); err != nil {
		log.Printf(err.Error())
	}

	// start the server
	logger.Info("Started server", &ServerInfo{
		ExternalUrl: *externalUrl,
		ApiVersion:  *apiVersion,
		Interface:   *listen,
	})
	log.Fatal(http.ListenAndServe(*listen, router))
}

// Simple route information
type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Simple port information
type ServerInfo struct {
	Interface   string `json:"interface"`
	ExternalUrl string `json:"externalUrl"`
	ApiVersion  string `json:"apiVersion"`
}
