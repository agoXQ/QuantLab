// Package svc is the composition root for the API Gateway. It dials
// every backend service over gRPC and exposes the typed clients the
// handler layer translates REST requests into.
package svc

import (
	"fmt"
	"log"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/agoXQ/QuantLab/app/ai/pb"
	backtestpb "github.com/agoXQ/QuantLab/app/backtest/pb"
	billingpb "github.com/agoXQ/QuantLab/app/billing/pb"
	communitypb "github.com/agoXQ/QuantLab/app/community/pb"
	"github.com/agoXQ/QuantLab/app/gateway/internal/config"
	formulapb "github.com/agoXQ/QuantLab/app/formula/pb"
	marketpb "github.com/agoXQ/QuantLab/app/market/pb"
	notificationpb "github.com/agoXQ/QuantLab/app/notification/pb"
	portfoliopb "github.com/agoXQ/QuantLab/app/portfolio/pb"
	rankingpb "github.com/agoXQ/QuantLab/app/ranking/pb"
	strategypb "github.com/agoXQ/QuantLab/app/strategy/pb"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
	userpb "github.com/agoXQ/QuantLab/app/user/pb"
)

// ServiceContext holds the JWT verifier and one gRPC client per
// backend service. A nil client means the service was not configured;
// handlers that touch it return a 503 so a missing service never looks
// like a normal empty result.
type ServiceContext struct {
	TokenVerifier *token.JWTIssuer

	User         userpb.UserServiceClient
	Strategy     strategypb.StrategyServiceClient
	Formula      formulapb.FormulaServiceClient
	// FormulaHTTPAddr is the formula service HTTP address for the
	// evaluate proxy (Evaluate is not on gRPC).
	FormulaHTTPAddr string
	Backtest     backtestpb.BacktestServiceClient
	Market       marketpb.MarketDataServiceClient
	Ranking      rankingpb.RankingServiceClient
	Portfolio    portfoliopb.PortfolioServiceClient
	Community    communitypb.CommunityServiceClient
	AI           pb.AIServiceClient
	Billing      billingpb.BillingServiceClient
	Notification notificationpb.NotificationServiceClient
}

// NewServiceContext dials every configured service and builds the
// gRPC client set. Unconfigured services log a warning and keep a nil
// client; the gateway still boots so the frontend can talk to whatever
// is running.
func NewServiceContext(c config.Config) *ServiceContext {
	sc := &ServiceContext{TokenVerifier: buildVerifier(c.Token)}
	sc.User = dial(c.User, userpb.NewUserServiceClient)
	sc.Strategy = dial(c.Strategy, strategypb.NewStrategyServiceClient)
	sc.Formula = dial(c.Formula, formulapb.NewFormulaServiceClient)
	sc.FormulaHTTPAddr = c.FormulaHTTP
	sc.Backtest = dial(c.Backtest, backtestpb.NewBacktestServiceClient)
	sc.Market = dial(c.Market, marketpb.NewMarketDataServiceClient)
	sc.Ranking = dial(c.Ranking, rankingpb.NewRankingServiceClient)
	sc.Portfolio = dial(c.Portfolio, portfoliopb.NewPortfolioServiceClient)
	sc.Community = dial(c.Community, communitypb.NewCommunityServiceClient)
	sc.AI = dial(c.AI, pb.NewAIServiceClient)
	sc.Billing = dial(c.Billing, billingpb.NewBillingServiceClient)
	sc.Notification = dial(c.Notification, notificationpb.NewNotificationServiceClient)
	return sc
}

func buildVerifier(c config.TokenConfig) *token.JWTIssuer {
	if len(c.Keys) == 0 {
		log.Printf("[gateway] token keys empty; protected routes will reject every request")
		return &token.JWTIssuer{}
	}
	keys := make([]token.SigningKey, 0, len(c.Keys))
	for _, k := range c.Keys {
		keys = append(keys, token.SigningKey{ID: k.ID, Secret: k.Secret})
	}
	keySet, err := token.NewKeySet(c.ActiveKeyID, keys)
	if err != nil {
		log.Fatalf("[gateway] invalid token key set: %v", err)
	}
	issuer, err := token.NewJWTIssuer(token.Config{Keys: keySet, Issuer: c.Issuer})
	if err != nil {
		log.Fatalf("[gateway] build jwt verifier: %v", err)
	}
	return issuer
}

// dial builds a zrpc client from a ServiceConfig and wraps it with the
// supplied gRPC constructor. An empty config yields a nil client (typed
// zero value) so handlers can detect the missing dependency. Every
// generated NewXxxServiceClient takes a grpc.ClientConnInterface, and
// zrpc.Client.Conn() returns the underlying *grpc.ClientConn.
func dial[C any](sc config.ServiceConfig, newClient func(grpc.ClientConnInterface) C) C {
	var zero C
	cli, ok := buildClient(sc)
	if !ok {
		return zero
	}
	return newClient(cli.Conn())
}

func buildClient(sc config.ServiceConfig) (zrpc.Client, bool) {
	if len(sc.Endpoints) == 0 {
		return nil, false
	}
	conf := zrpc.RpcClientConf{
		Endpoints: sc.Endpoints,
		Timeout:   sc.Timeout,
		NonBlock:  sc.NonBlock,
	}
	cli, err := zrpc.NewClient(conf)
	if err != nil {
		log.Printf("[gateway] dial %v: %v", sc.Endpoints, err)
		return nil, false
	}
	return cli, true
}

// Summary is a startup log helper.
func (sc *ServiceContext) Summary() string {
	return fmt.Sprintf(
		"gateway ready: user=%v strategy=%v formula=%v backtest=%v market=%v ranking=%v portfolio=%v community=%v ai=%v billing=%v notification=%v",
		sc.User != nil, sc.Strategy != nil, sc.Formula != nil, sc.Backtest != nil,
		sc.Market != nil, sc.Ranking != nil, sc.Portfolio != nil, sc.Community != nil,
		sc.AI != nil, sc.Billing != nil, sc.Notification != nil,
	)
}
