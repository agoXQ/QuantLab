// Package main is the QuantLab API Gateway entry point. It exposes a
// single REST surface (/api/v1/*) backed by gRPC clients to every
// backend service, so the frontend dials one address and the per-
// service HTTP handlers can be retired.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/core/conf"

	"github.com/agoXQ/QuantLab/app/gateway/internal/config"
	"github.com/agoXQ/QuantLab/app/gateway/internal/svc"
	"github.com/agoXQ/QuantLab/app/gateway/interfaces/handler"
	"github.com/agoXQ/QuantLab/app/gateway/interfaces/middleware"
)

var configFile = flag.String("f", "app/gateway/etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	sc := svc.NewServiceContext(c)
	log.Printf("[gateway] %s", sc.Summary())

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(middleware.CORS())
	// Optional auth on every route: resolves the caller id when a token
	// is present without rejecting anonymous traffic. Mutating handlers
	// that need a caller read it via middleware.CallerID.
	router.Use(middleware.Auth(sc.TokenVerifier, false))

	api := router.Group("/api/v1")
	h := handler.NewHandler(sc)
	h.RegisterAll(api)

	// Health check for docker / k8s liveness probes.
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", c.HttpPort)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.Printf("[gateway] listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[gateway] server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("[gateway] shutting down...")
	_ = srv.Close()
}
