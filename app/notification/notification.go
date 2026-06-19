package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/notification/interfaces/http"
	"github.com/agoXQ/QuantLab/app/notification/internal/config"
	"github.com/agoXQ/QuantLab/app/notification/internal/server"
	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/notification.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterNotificationServiceServer(grpcServer, server.NewNotificationServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()
	defer ctx.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("shutdown signal received, releasing resources...")
		ctx.Close()
		s.Stop()
		os.Exit(0)
	}()

	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())

		apiGroup := router.Group("/api/v1/notifications")
		handler := httpHandler.NewHandler(ctx.Service)
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
