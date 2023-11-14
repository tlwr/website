package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nlogrus "github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	nsecure "github.com/unrolled/secure"
	"github.com/urfave/negroni"
	nprom "github.com/zbindenren/negroni-prometheus"
)

func main() {
	renderer := render.New(render.Options{
		Directory: "templates",
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "healthy")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		_ = renderer.HTML(w, http.StatusNotFound, "404", nil)
	})

	mux.Handle("/metrics", promhttp.Handler())

	level := logrus.InfoLevel

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(nlogrus.NewCustomMiddleware(level, &logrus.JSONFormatter{}, "web"))
	n.Use(gzip.Gzip(gzip.DefaultCompression))
	n.Use(negroni.HandlerFunc(nsecure.New().HandlerFuncWithNext))
	n.Use(nprom.NewMiddleware("www.toby.codes"))
	n.UseHandler(mux)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	server := &http.Server{Addr: ":8080", Handler: n}

	go func() {
		_ = server.ListenAndServe()
	}()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	os.Exit(0)
}
