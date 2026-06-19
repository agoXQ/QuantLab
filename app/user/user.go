package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	authmiddleware "github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
	httpHandler "github.com/agoXQ/QuantLab/app/user/interfaces/http"
	"github.com/agoXQ/QuantLab/app/user/internal/config"
	"github.com/agoXQ/QuantLab/app/user/internal/server"
	"github.com/agoXQ/QuantLab/app/user/internal/svc"
	"github.com/agoXQ/QuantLab/app/user/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	authInterceptor := authmiddleware.GRPCAuth(ctx.TokenIssuer())

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(authInterceptor)
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
		router.Use(authmiddleware.GinAuth(ctx.TokenIssuer(), false))

		apiGroup := router.Group("/api/v1/users")
		handler := httpHandler.NewHandler(ctx.UserSvc)
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
