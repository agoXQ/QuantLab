package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	httpHandler "github.com/agoXQ/QuantLab/app/formula/interfaces/http"
	"github.com/agoXQ/QuantLab/app/formula/internal/config"
	"github.com/agoXQ/QuantLab/app/formula/internal/server"
	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/formula.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// Start gRPC server
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterFormulaServiceServer(grpcServer, server.NewFormulaServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// Start Prometheus metrics HTTP server
	go func() {
		metricsAddr := ":9091"
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		fmt.Printf("Starting metrics server at %s...\n", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()

	// Start REST API HTTP server
	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())

		apiGroup := router.Group("/api/v1/formula")
		handler := httpHandler.NewHandlerWithEvaluator(
			ctx.FormulaService,
			ctx.EvaluatorService,
			ctx.DataPort,
		)
		handler.RegisterRoutes(apiGroup)

		addr := fmt.Sprintf("0.0.0.0:%d", c.HttpPort)
		fmt.Printf("Starting HTTP server at %s...\n", addr)
		if err := router.Run(addr); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
