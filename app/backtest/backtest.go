package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/backtest/interfaces/http"
	"github.com/agoXQ/QuantLab/app/backtest/internal/config"
	"github.com/agoXQ/QuantLab/app/backtest/internal/server"
	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/backtest.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// gRPC server: kept for the eventual queued worker path. The MVP
	// surface is HTTP, but we keep the goctl-generated gRPC server alive
	// so service discovery (etcd) can still register the backtest service.
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterBacktestServiceServer(grpcServer, server.NewBacktestServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())

		apiGroup := router.Group("/api/v1/backtests")
		handler := httpHandler.NewHandler(ctx.BacktestSvc)
		handler.RegisterRoutes(apiGroup)

		addr := fmt.Sprintf("0.0.0.0:%d", c.HttpPort)
		fmt.Printf("Starting HTTP server at %s...\n", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
