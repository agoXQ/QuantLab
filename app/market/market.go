package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/market/interfaces/http"
	"github.com/agoXQ/QuantLab/app/market/internal/config"
	"github.com/agoXQ/QuantLab/app/market/internal/server"
	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/market.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// gRPC server.
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterMarketDataServiceServer(grpcServer, server.NewMarketDataServiceServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// HTTP REST server.
	if ctx.MarketService != nil {
		go func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.Use(gin.Recovery())

			apiGroup := router.Group("/api/v1/market")
			handler := httpHandler.NewHandler(ctx.MarketService)
			handler.RegisterRoutes(apiGroup)

			router.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

			addr := fmt.Sprintf("0.0.0.0:%d", c.HttpPort)
			fmt.Printf("Starting HTTP server at %s...\n", addr)
			if err := router.Run(addr); err != nil {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()
	}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
