package main

import (
	"flag"
	"github.com/766b/chi-prometheus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/pedro-gutierrez/form3/pkg/admin"
	"github.com/pedro-gutierrez/form3/pkg/payments"
	"github.com/pedro-gutierrez/form3/pkg/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	l "github.com/treastech/logger"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"
	"go.uber.org/zap"
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
	repoLogs       *bool
	enableCors     *bool
	timeout        *int
	adminRoutes    *bool
	profiling      *bool
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
	repoLogs = flag.Bool("repo-logs", false, "enable repo logs")
	adminRoutes = flag.Bool("admin", false, "enable admin endpoints")
	profiling = flag.Bool("profiling", false, "enable profiling")
}

// Main entry point to the program. Connects to the database, configures
// all required routes, and starts listening for connections
func main() {
	flag.Parse()

	// Setup our persistence. We do this first, since we want to exit
	// the program, in case the database is not available
	repo, err := util.NewRepo(util.RepoConfig{
		Driver:     *repoDriver,
		Uri:        *repoUri,
		Debug:      *repoLogs,
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

	log.Printf("Using repo: %s", repo.Description())

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
		log.Printf("Exposing Prometheus metrics")
		router.Use(chiprometheus.NewMiddleware("payments"))
	}

	// Maybe turn on logging of http requests/responses
	if *httpLogs {
		// JSON logger, so that it is easier to consume
		// logs (eg. with Elasticsearch)

		logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
		router.Use(l.Logger(logger))

		//	router.Use(zerochi.NewStructuredLogger(zerochi.NewLogger()))
		log.Printf("Logging responses")
	}

	// Maybe GZIP http responses
	if *compress {
		log.Printf("Compressing responses")
		router.Use(middleware.DefaultCompress)
	}

	// Maybe set a rate limit, so that we keep things
	// controlled under high load. A 429 status code will
	// be returned once we reach the configured threshold.
	// For simplicity, we use the memory store but there are other
	// options (eg. Redis)
	if *limit != "" {
		log.Printf("Setting rate-limit to %v", *limit)
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
		log.Printf("Enabling CORS")
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
		router.Handle("/metrics", prometheus.Handler())
	}

	// If profiling is enabled, enable net/http/pprof middleware
	if *profiling {
		router.Mount("/profiling", middleware.Profiler())
	}

	// mount application logic
	router.Route("/v1", func(v1Router chi.Router) {

		// Admin features
		// (useful for testing, for example, but use with care in a production
		// environment
		if *adminRoutes {
			v1Router.Route("/admin", func(adminRouter chi.Router) {
				adminRouter.Mount("/", admin.New(repo).Routes())
			})
			log.Printf("Admin routes added")
		}

		// payments api
		v1Router.Mount("/", payments.New(repo).Routes())

		// more endpoints here...
	})

	// Print all routes mounted
	if err := chi.Walk(router, func(method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		log.Printf("Mounted: %s %s", method, route)
		return nil
	}); err != nil {
		log.Printf(err.Error())
	}

	// start the server
	log.Printf("Listening on %v", *listen)
	log.Fatal(http.ListenAndServe(*listen, router))
}
