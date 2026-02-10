package restapp

import (
	"context"
	"net/http"
	"strconv"
	restmiddleware "usersservice/internal/app/rest/middleware"
	"usersservice/internal/handlers"
	"usersservice/internal/handlers/rest/users"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type App struct {
	log     *zap.Logger
	service handlers.Service
	port    int
	server  *http.Server
}

func New(log *zap.Logger, service handlers.Service, port int) *App {
	return &App{
		log:     log,
		service: service,
		port:    port,
	}
}

func (a *App) Start() error {
	const op = "app.rest.Start"
	log := a.log.With(zap.String("op", op))

	usersHandler := users.New(log, a.service)

	base := mux.NewRouter()
	base.Use(restmiddleware.CORS)
	base.Use(restmiddleware.RequestLoggerMiddleware(log))
	base.Handle("/metrics", promhttp.Handler())

	router := base.PathPrefix("/api/v1").Subrouter()

	router.HandleFunc("/users", usersHandler.GetUsers).Methods(http.MethodGet)
	router.HandleFunc("/users/{id}", usersHandler.GetUserById).Methods(http.MethodGet)
	router.HandleFunc("/users", usersHandler.Insert).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", usersHandler.Update).Methods(http.MethodPut)
	router.HandleFunc("/users/{id}/password", usersHandler.ChangePassword).Methods(http.MethodPatch)
	router.HandleFunc("/users/{id}", usersHandler.Delete).Methods(http.MethodDelete)

	log.Info("starting rest server", zap.String("op", op), zap.Int("port", a.port))

	a.server = &http.Server{
		Addr:    ":" + strconv.Itoa(a.port),
		Handler: base,
	}

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", zap.String("op", op), zap.Error(err))
		return err
	}

	return nil
}

func (a *App) Shutdown() error {
	return a.server.Shutdown(context.Background())
}
