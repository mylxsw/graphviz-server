package api

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/graphviz-server/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServiceProvider struct{}

func (s ServiceProvider) Register(app container.Container) {}

func (s ServiceProvider) Boot(app infra.Glacier) {
	app.MustResolve(func(conf *config.Config) {
		app.WebAppRouter(routers(app.Container()))
		app.WebAppMuxRouter(func(router *mux.Router) {
			// prometheus metrics
			router.PathPrefix("/metrics").Handler(promhttp.Handler())
			// health check
			router.PathPrefix("/health").Handler(HealthCheck{})

			router.PathPrefix("/resources").Handler(http.StripPrefix("/resources", http.FileServer(http.Dir(filepath.Join(conf.TempDir, "graphviz")))))
			router.PathPrefix("/dashboard").Handler(http.StripPrefix("/dashboard", http.FileServer(FS(conf.Debug))))
		})
	})
}

type HealthCheck struct{}

func (h HealthCheck) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(`{"status": "UP"}`))
}
