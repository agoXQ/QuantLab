#!/bin/bash
set -e

cd /Users/agoxq/Documents/Project/QuantLab

# Strategy proto
cat > api/strategy/v1/strategy.proto << 'EOF'
syntax = "proto3";

package strategy;

option go_package = "github.com/agoXQ/QuantLab/app/strategy/pb";

import "api/common/v1/common.proto";

service StrategyService {
  rpc CreateStrategy(CreateStrategyRequest) returns (CreateStrategyResponse);
  rpc GetStrategy(GetStrategyRequest) returns (GetStrategyResponse);
  rpc UpdateStrategy(UpdateStrategyRequest) returns (UpdateStrategyResponse);
  rpc DeleteStrategy(DeleteStrategyRequest) returns (DeleteStrategyResponse);
  rpc PublishStrategy(PublishStrategyRequest) returns (PublishStrategyResponse);
  rpc ArchiveStrategy(ArchiveStrategyRequest) returns (ArchiveStrategyResponse);
  rpc ForkStrategy(ForkStrategyRequest) returns (ForkStrategyResponse);
  rpc CreateVersion(CreateVersionRequest) returns (CreateVersionResponse);
  rpc ListVersions(ListVersionsRequest) returns (ListVersionsResponse);
  rpc GetVersion(GetVersionRequest) returns (GetVersionResponse);
  rpc SearchStrategies(SearchStrategiesRequest) returns (SearchStrategiesResponse);
  rpc ListStrategies(ListStrategiesRequest) returns (ListStrategiesResponse);
}

message Strategy {
  int64 id = 1;
  int64 author_id = 2;
  string title = 3;
  string description = 4;
  common.Status status = 5;
  common.Visibility visibility = 6;
  string category = 7;
  repeated string tags = 8;
  int64 current_version_id = 9;
  int64 view_count = 10;
  int64 favorite_count = 11;
  int64 fork_count = 12;
  int64 created_at = 13;
  int64 updated_at = 14;
}

message StrategyVersion {
  int64 id = 1;
  int64 strategy_id = 2;
  string version_no = 3;
  string formula_text = 4;
  string buy_rule = 5;
  string sell_rule = 6;
  string risk_rule = 7;
  string position_rule = 8;
  string rebalance_rule = 9;
  string change_log = 10;
  int64 created_by = 11;
  int64 created_at = 12;
}

message CreateStrategyRequest {
  string title = 1;
  string description = 2;
  string category = 3;
  repeated string tags = 4;
}
message CreateStrategyResponse { int64 strategy_id = 1; }

message GetStrategyRequest { int64 strategy_id = 1; }
message GetStrategyResponse { Strategy strategy = 1; }

message UpdateStrategyRequest {
  int64 strategy_id = 1;
  string title = 2;
  string description = 3;
  string category = 4;
  repeated string tags = 5;
}
message UpdateStrategyResponse {}

message DeleteStrategyRequest { int64 strategy_id = 1; }
message DeleteStrategyResponse {}

message PublishStrategyRequest {
  int64 strategy_id = 1;
  int64 version_id = 2;
}
message PublishStrategyResponse {}

message ArchiveStrategyRequest { int64 strategy_id = 1; }
message ArchiveStrategyResponse {}

message ForkStrategyRequest { int64 source_strategy_id = 1; }
message ForkStrategyResponse { int64 new_strategy_id = 1; }

message CreateVersionRequest {
  int64 strategy_id = 1;
  string formula_text = 2;
  string buy_rule = 3;
  string sell_rule = 4;
  string risk_rule = 5;
  string position_rule = 6;
  string rebalance_rule = 7;
  string change_log = 8;
}
message CreateVersionResponse { int64 version_id = 1; }

message ListVersionsRequest { int64 strategy_id = 1; }
message ListVersionsResponse { repeated StrategyVersion versions = 1; }

message GetVersionRequest { int64 version_id = 1; }
message GetVersionResponse { StrategyVersion version = 1; }

message SearchStrategiesRequest {
  string keyword = 1;
  repeated string tags = 2;
  int64 author_id = 3;
  string category = 4;
  string sort = 5;
  common.Cursor cursor = 6;
  int32 limit = 7;
}
message SearchStrategiesResponse {
  repeated Strategy strategies = 1;
  common.Cursor cursor = 2;
}

message ListStrategiesRequest {
  int64 author_id = 1;
  common.Status status = 2;
  common.Cursor cursor = 3;
  int32 limit = 4;
}
message ListStrategiesResponse {
  repeated Strategy strategies = 1;
  common.Cursor cursor = 2;
}
EOF
echo "strategy.proto created"

# Portfolio proto
cat > api/portfolio/v1/portfolio.proto << 'EOF'
syntax = "proto3";

package portfolio;

option go_package = "github.com/agoXQ/QuantLab/app/portfolio/pb";

import "api/common/v1/common.proto";

service PortfolioService {
  rpc CreatePortfolio(CreatePortfolioRequest) returns (CreatePortfolioResponse);
  rpc GetPortfolio(GetPortfolioRequest) returns (GetPortfolioResponse);
  rpc UpdatePortfolio(UpdatePortfolioRequest) returns (UpdatePortfolioResponse);
  rpc DeletePortfolio(DeletePortfolioRequest) returns (DeletePortfolioResponse);
  rpc AddItem(AddItemRequest) returns (AddItemResponse);
  rpc RemoveItem(RemoveItemRequest) returns (RemoveItemResponse);
  rpc UpdateWeights(UpdateWeightsRequest) returns (UpdateWeightsResponse);
  rpc PublishPortfolio(PublishPortfolioRequest) returns (PublishPortfolioResponse);
  rpc GetAnalytics(GetAnalyticsRequest) returns (GetAnalyticsResponse);
  rpc GetEquityCurve(GetEquityCurveRequest) returns (GetEquityCurveResponse);
  rpc ListPortfolios(ListPortfoliosRequest) returns (ListPortfoliosResponse);
}

message Portfolio {
  int64 id = 1;
  int64 owner_id = 2;
  string name = 3;
  string description = 4;
  common.Status status = 5;
  common.Visibility visibility = 6;
  int32 current_version = 7;
  repeated PortfolioItem items = 8;
  RebalanceRule rebalance_rule = 9;
  int64 created_at = 10;
  int64 updated_at = 11;
}

message PortfolioItem {
  int64 id = 1;
  int64 strategy_id = 2;
  double weight = 3;
  int32 sort_order = 4;
}

message RebalanceRule {
  string frequency = 1;
  string method = 2;
  double threshold = 3;
}

message PortfolioAnalytics {
  double annual_return = 1;
  double total_return = 2;
  double sharpe_ratio = 3;
  double max_drawdown = 4;
  double volatility = 5;
  double win_rate = 6;
}

message CreatePortfolioRequest {
  string name = 1;
  string description = 2;
  common.Visibility visibility = 3;
}
message CreatePortfolioResponse { int64 portfolio_id = 1; }

message GetPortfolioRequest { int64 portfolio_id = 1; }
message GetPortfolioResponse { Portfolio portfolio = 1; }

message UpdatePortfolioRequest {
  int64 portfolio_id = 1;
  string name = 2;
  string description = 3;
  common.Visibility visibility = 4;
}
message UpdatePortfolioResponse {}

message DeletePortfolioRequest { int64 portfolio_id = 1; }
message DeletePortfolioResponse {}

message AddItemRequest {
  int64 portfolio_id = 1;
  int64 strategy_id = 2;
  double weight = 3;
}
message AddItemResponse { int64 item_id = 1; }

message RemoveItemRequest {
  int64 portfolio_id = 1;
  int64 item_id = 2;
}
message RemoveItemResponse {}

message UpdateWeightsRequest {
  int64 portfolio_id = 1;
  repeated WeightItem items = 2;
}
message WeightItem {
  int64 strategy_id = 1;
  double weight = 2;
}
message UpdateWeightsResponse {}

message PublishPortfolioRequest { int64 portfolio_id = 1; }
message PublishPortfolioResponse {}

message GetAnalyticsRequest { int64 portfolio_id = 1; }
message GetAnalyticsResponse { PortfolioAnalytics analytics = 1; }

message GetEquityCurveRequest { int64 portfolio_id = 1; }
message EquityPoint {
  string trade_date = 1;
  double nav = 2;
  double return_rate = 3;
}
message GetEquityCurveResponse { repeated EquityPoint points = 1; }

message ListPortfoliosRequest {
  int64 owner_id = 1;
  common.Status status = 2;
  common.Cursor cursor = 3;
  int32 limit = 4;
}
message ListPortfoliosResponse {
  repeated Portfolio portfolios = 1;
  common.Cursor cursor = 2;
}
EOF
echo "portfolio.proto created"

# Billing proto
cat > api/billing/v1/billing.proto << 'EOF'
syntax = "proto3";

package billing;

option go_package = "github.com/agoXQ/QuantLab/app/billing/pb";

import "api/common/v1/common.proto";

service BillingService {
  rpc GetMembership(GetMembershipRequest) returns (GetMembershipResponse);
  rpc PurchaseMembership(PurchaseMembershipRequest) returns (PurchaseMembershipResponse);
  rpc CancelMembership(CancelMembershipRequest) returns (CancelMembershipResponse);
  rpc CreateSubscription(CreateSubscriptionRequest) returns (CreateSubscriptionResponse);
  rpc CancelSubscription(CancelSubscriptionRequest) returns (CancelSubscriptionResponse);
  rpc ListSubscriptions(ListSubscriptionsRequest) returns (ListSubscriptionsResponse);
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
  rpc PaymentWebhook(PaymentWebhookRequest) returns (PaymentWebhookResponse);
  rpc GetCreatorRevenue(GetCreatorRevenueRequest) returns (GetCreatorRevenueResponse);
  rpc RequestSettlement(RequestSettlementRequest) returns (RequestSettlementResponse);
  rpc ListSettlements(ListSettlementsRequest) returns (ListSettlementsResponse);
}

message Membership {
  string membership_id = 1;
  int64 user_id = 2;
  string tier = 3;
  string status = 4;
  bool auto_renew = 5;
  int64 started_at = 6;
  int64 expired_at = 7;
}

message Subscription {
  string subscription_id = 1;
  int64 subscriber_id = 2;
  string resource_type = 3;
  string resource_id = 4;
  string status = 5;
  bool auto_renew = 6;
  int64 started_at = 7;
  int64 expired_at = 8;
}

message Order {
  string order_id = 1;
  int64 user_id = 2;
  string order_type = 3;
  double amount = 4;
  string currency = 5;
  string status = 6;
  int64 created_at = 7;
  int64 paid_at = 8;
}

message Revenue {
  string revenue_share_id = 1;
  string order_id = 2;
  int64 creator_id = 3;
  double gross_amount = 4;
  double platform_amount = 5;
  double creator_amount = 6;
  string status = 7;
}

message Settlement {
  string settlement_id = 1;
  int64 creator_id = 2;
  double amount = 3;
  string currency = 4;
  string status = 5;
  int64 requested_at = 6;
  int64 paid_at = 7;
}

message GetMembershipRequest { int64 user_id = 1; }
message GetMembershipResponse { Membership membership = 1; }

message PurchaseMembershipRequest {
  string tier = 1;
  string billing_cycle = 2;
}
message PurchaseMembershipResponse { string order_id = 1; }

message CancelMembershipRequest {}
message CancelMembershipResponse {}

message CreateSubscriptionRequest {
  string resource_type = 1;
  string resource_id = 2;
}
message CreateSubscriptionResponse { string subscription_id = 1; }

message CancelSubscriptionRequest { string subscription_id = 1; }
message CancelSubscriptionResponse {}

message ListSubscriptionsRequest {
  int64 user_id = 1;
  common.Cursor cursor = 2;
  int32 limit = 3;
}
message ListSubscriptionsResponse {
  repeated Subscription subscriptions = 1;
  common.Cursor cursor = 2;
}

message CreateOrderRequest {
  string order_type = 1;
  string resource_id = 2;
  double amount = 3;
  string currency = 4;
}
message CreateOrderResponse { string order_id = 1; }

message GetOrderRequest { string order_id = 1; }
message GetOrderResponse { Order order = 1; }

message ListOrdersRequest {
  int64 user_id = 1;
  common.Cursor cursor = 2;
  int32 limit = 3;
}
message ListOrdersResponse {
  repeated Order orders = 1;
  common.Cursor cursor = 2;
}

message CreatePaymentRequest {
  string order_id = 1;
  string channel = 2;
}
message CreatePaymentResponse { string payment_url = 1; }

message PaymentWebhookRequest {
  string channel = 1;
  bytes raw_body = 2;
}
message PaymentWebhookResponse {}

message GetCreatorRevenueRequest { int64 creator_id = 1; }
message GetCreatorRevenueResponse {
  double total_revenue = 1;
  double pending_revenue = 2;
  double settled_revenue = 3;
  repeated Revenue recent_revenues = 4;
}

message RequestSettlementRequest {
  double amount = 1;
  string currency = 2;
}
message RequestSettlementResponse { string settlement_id = 1; }

message ListSettlementsRequest {
  int64 creator_id = 1;
  common.Cursor cursor = 2;
  int32 limit = 3;
}
message ListSettlementsResponse {
  repeated Settlement settlements = 1;
  common.Cursor cursor = 2;
}
EOF
echo "billing.proto created"

# Community proto
cat > api/community/v1/community.proto << 'EOF'
syntax = "proto3";

package community;

option go_package = "github.com/agoXQ/QuantLab/app/community/pb";

import "api/common/v1/common.proto";

service CommunityService {
  rpc CreateContent(CreateContentRequest) returns (CreateContentResponse);
  rpc GetContent(GetContentRequest) returns (GetContentResponse);
  rpc GetFeed(GetFeedRequest) returns (GetFeedResponse);
  rpc LikeContent(LikeContentRequest) returns (LikeContentResponse);
  rpc UnlikeContent(UnlikeContentRequest) returns (UnlikeContentResponse);
  rpc FavoriteContent(FavoriteContentRequest) returns (FavoriteContentResponse);
  rpc UnfavoriteContent(UnfavoriteContentRequest) returns (UnfavoriteContentResponse);
  rpc CreateComment(CreateCommentRequest) returns (CreateCommentResponse);
  rpc ListComments(ListCommentsRequest) returns (ListCommentsResponse);
  rpc DeleteComment(DeleteCommentRequest) returns (DeleteCommentResponse);
  rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
  rpc ListUserContents(ListUserContentsRequest) returns (ListUserContentsResponse);
}

message Content {
  int64 id = 1;
  string content_type = 2;
  int64 object_id = 3;
  int64 author_id = 4;
  string title = 5;
  string summary = 6;
  common.Visibility visibility = 7;
  common.Status status = 8;
  int64 like_count = 9;
  int64 favorite_count = 10;
  int64 comment_count = 11;
  int64 view_count = 12;
  int64 created_at = 13;
}

message Comment {
  int64 id = 1;
  int64 content_id = 2;
  int64 user_id = 3;
  int64 parent_id = 4;
  string body = 5;
  string status = 6;
  int64 created_at = 7;
}

message FeedItem {
  Content content = 1;
  double score = 2;
  int64 created_at = 3;
}

message UserProfile {
  int64 user_id = 1;
  string nickname = 2;
  string avatar_url = 3;
  string bio = 4;
  int64 follower_count = 5;
  int64 following_count = 6;
  int64 content_count = 7;
}

message CreateContentRequest {
  string content_type = 1;
  int64 object_id = 2;
  string title = 3;
  string summary = 4;
}
message CreateContentResponse { int64 content_id = 1; }

message GetContentRequest { int64 content_id = 1; }
message GetContentResponse { Content content = 1; }

message GetFeedRequest {
  common.Cursor cursor = 1;
  int32 limit = 2;
}
message GetFeedResponse {
  repeated FeedItem items = 1;
  common.Cursor cursor = 2;
}

message LikeContentRequest { int64 content_id = 1; }
message LikeContentResponse {}

message UnlikeContentRequest { int64 content_id = 1; }
message UnlikeContentResponse {}

message FavoriteContentRequest { int64 content_id = 1; }
message FavoriteContentResponse {}

message UnfavoriteContentRequest { int64 content_id = 1; }
message UnfavoriteContentResponse {}

message CreateCommentRequest {
  int64 content_id = 1;
  int64 parent_id = 2;
  string body = 3;
}
message CreateCommentResponse { int64 comment_id = 1; }

message ListCommentsRequest {
  int64 content_id = 1;
  common.Cursor cursor = 2;
  int32 limit = 3;
}
message ListCommentsResponse {
  repeated Comment comments = 1;
  common.Cursor cursor = 2;
}

message DeleteCommentRequest { int64 comment_id = 1; }
message DeleteCommentResponse {}

message GetProfileRequest { int64 user_id = 1; }
message GetProfileResponse { UserProfile profile = 1; }

message ListUserContentsRequest {
  int64 user_id = 1;
  common.Cursor cursor = 2;
  int32 limit = 3;
}
message ListUserContentsResponse {
  repeated Content contents = 1;
  common.Cursor cursor = 2;
}
EOF
echo "community.proto created"

# Ranking proto
cat > api/ranking/v1/ranking.proto << 'EOF'
syntax = "proto3";

package ranking;

option go_package = "github.com/agoXQ/QuantLab/app/ranking/pb";

import "api/common/v1/common.proto";

service RankingService {
  rpc GetRanking(GetRankingRequest) returns (GetRankingResponse);
  rpc GetStrategyRank(GetStrategyRankRequest) returns (GetStrategyRankResponse);
  rpc GetAuthorRank(GetAuthorRankRequest) returns (GetAuthorRankResponse);
  rpc GetHistoryRanking(GetHistoryRankingRequest) returns (GetHistoryRankingResponse);
  rpc ListSnapshots(ListSnapshotsRequest) returns (ListSnapshotsResponse);
}

enum RankingType {
  RANKING_TYPE_UNSPECIFIED = 0;
  RETURN = 1;
  SHARPE = 2;
  WIN_RATE = 3;
  DRAWDOWN = 4;
  STABILITY = 5;
  OVERALL = 6;
  POPULARITY = 7;
  AUTHOR = 8;
}

enum RankingPeriod {
  RANKING_PERIOD_UNSPECIFIED = 0;
  DAILY = 1;
  WEEKLY = 2;
  MONTHLY = 3;
  QUARTERLY = 4;
  YEARLY = 5;
  ALL_TIME = 6;
}

message RankingItem {
  int32 rank = 1;
  int64 strategy_id = 2;
  string strategy_name = 3;
  int64 author_id = 4;
  string author_name = 5;
  double score = 6;
  double trust_score = 7;
  double total_return = 8;
  double sharpe_ratio = 9;
  double max_drawdown = 10;
  double win_rate = 11;
  int32 rank_change = 12;
}

message AuthorRankingItem {
  int32 rank = 1;
  int64 author_id = 2;
  string author_name = 3;
  double author_score = 4;
  int32 strategy_count = 5;
  double avg_return = 6;
  double avg_sharpe = 7;
  int64 follower_count = 8;
}

message RankingSnapshot {
  int64 id = 1;
  RankingType ranking_type = 2;
  RankingPeriod ranking_period = 3;
  int64 snapshot_time = 4;
}

message GetRankingRequest {
  RankingType type = 1;
  RankingPeriod period = 2;
  common.Cursor cursor = 3;
  int32 limit = 4;
}
message GetRankingResponse {
  repeated RankingItem items = 1;
  common.Cursor cursor = 2;
}

message GetStrategyRankRequest {
  int64 strategy_id = 1;
  RankingType type = 2;
}
message GetStrategyRankResponse {
  int32 rank = 1;
  double score = 2;
  double trust_score = 3;
  RankingType type = 4;
  RankingPeriod period = 5;
}

message GetAuthorRankRequest { int64 author_id = 1; }
message GetAuthorRankResponse {
  int32 rank = 1;
  double author_score = 2;
}

message GetHistoryRankingRequest {
  int64 strategy_id = 1;
  RankingType type = 2;
  int64 start_time = 3;
  int64 end_time = 4;
}
message HistoryPoint {
  int64 snapshot_time = 1;
  int32 rank = 2;
  double score = 3;
}
message GetHistoryRankingResponse { repeated HistoryPoint history = 1; }

message ListSnapshotsRequest {
  RankingType type = 1;
  RankingPeriod period = 2;
  common.Cursor cursor = 3;
  int32 limit = 4;
}
message ListSnapshotsResponse {
  repeated RankingSnapshot snapshots = 1;
  common.Cursor cursor = 2;
}
EOF
echo "ranking.proto created"

# Notification proto
cat > api/notification/v1/notification.proto << 'EOF'
syntax = "proto3";

package notification;

option go_package = "github.com/agoXQ/QuantLab/app/notification/pb";

import "api/common/v1/common.proto";

service NotificationService {
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
  rpc MarkRead(MarkReadRequest) returns (MarkReadResponse);
  rpc MarkAllRead(MarkAllReadRequest) returns (MarkAllReadResponse);
  rpc DeleteNotification(DeleteNotificationRequest) returns (DeleteNotificationResponse);
  rpc GetPreferences(GetPreferencesRequest) returns (GetPreferencesResponse);
  rpc UpdatePreferences(UpdatePreferencesRequest) returns (UpdatePreferencesResponse);
  rpc CreateSubscription(CreateSubscriptionRequest) returns (CreateSubscriptionResponse);
  rpc CancelSubscription(CancelSubscriptionRequest) returns (CancelSubscriptionResponse);
  rpc ListSubscriptions(ListSubscriptionsRequest) returns (ListSubscriptionsResponse);
}

enum NotificationType {
  NOTIFICATION_TYPE_UNSPECIFIED = 0;
  SYSTEM = 1;
  COMMENT = 2;
  LIKE = 3;
  FOLLOW = 4;
  MENTION = 5;
  RANKING = 6;
  STRATEGY = 7;
  PORTFOLIO = 8;
  BACKTEST = 9;
  MEMBERSHIP = 10;
}

message Notification {
  int64 id = 1;
  int64 user_id = 2;
  NotificationType type = 3;
  string title = 4;
  string content = 5;
  string status = 6;
  int64 read_at = 7;
  int64 created_at = 8;
}

message NotificationPreference {
  int64 user_id = 1;
  bool in_app_enabled = 2;
  bool email_enabled = 3;
  bool webhook_enabled = 4;
  bool push_enabled = 5;
}

message NotificationSubscription {
  int64 id = 1;
  int64 subscriber_id = 2;
  string object_type = 3;
  int64 object_id = 4;
  int64 created_at = 5;
}

message ListNotificationsRequest {
  common.Cursor cursor = 1;
  int32 limit = 2;
}
message ListNotificationsResponse {
  repeated Notification notifications = 1;
  common.Cursor cursor = 2;
}

message GetUnreadCountRequest {}
message GetUnreadCountResponse { int32 count = 1; }

message MarkReadRequest { int64 notification_id = 1; }
message MarkReadResponse {}

message MarkAllReadRequest {}
message MarkAllReadResponse {}

message DeleteNotificationRequest { int64 notification_id = 1; }
message DeleteNotificationResponse {}

message GetPreferencesRequest {}
message GetPreferencesResponse { NotificationPreference preferences = 1; }

message UpdatePreferencesRequest {
  bool in_app_enabled = 1;
  bool email_enabled = 2;
  bool webhook_enabled = 3;
  bool push_enabled = 4;
}
message UpdatePreferencesResponse {}

message CreateSubscriptionRequest {
  string object_type = 1;
  int64 object_id = 2;
}
message CreateSubscriptionResponse { int64 subscription_id = 1; }

message CancelSubscriptionRequest { int64 subscription_id = 1; }
message CancelSubscriptionResponse {}

message ListSubscriptionsRequest {
  common.Cursor cursor = 1;
  int32 limit = 2;
}
message ListSubscriptionsResponse {
  repeated NotificationSubscription subscriptions = 1;
  common.Cursor cursor = 2;
}
EOF
echo "notification.proto created"

# AI proto
cat > api/ai/v1/ai.proto << 'EOF'
syntax = "proto3";

package ai;

option go_package = "github.com/agoXQ/QuantLab/app/ai/pb";

import "api/common/v1/common.proto";

service AIService {
  rpc GenerateStrategy(GenerateStrategyRequest) returns (GenerateStrategyResponse);
  rpc ExplainStrategy(ExplainStrategyRequest) returns (ExplainStrategyResponse);
  rpc OptimizeStrategy(OptimizeStrategyRequest) returns (OptimizeStrategyResponse);
  rpc GeneratePortfolio(GeneratePortfolioRequest) returns (GeneratePortfolioResponse);
  rpc OptimizePortfolio(OptimizePortfolioRequest) returns (OptimizePortfolioResponse);
  rpc AnalyzeBacktest(AnalyzeBacktestRequest) returns (AnalyzeBacktestResponse);
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc GetReport(GetReportRequest) returns (GetReportResponse);
}

enum TaskType {
  TASK_TYPE_UNSPECIFIED = 0;
  GENERATE_STRATEGY = 1;
  EXPLAIN_STRATEGY = 2;
  OPTIMIZE_STRATEGY = 3;
  GENERATE_PORTFOLIO = 4;
  OPTIMIZE_PORTFOLIO = 5;
  ANALYZE_BACKTEST = 6;
  ASK_QUANTLAB = 7;
}

enum TaskStatus {
  TASK_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  RUNNING = 2;
  COMPLETED = 3;
  FAILED = 4;
  CANCELLED = 5;
}

message AITask {
  int64 id = 1;
  int64 user_id = 2;
  TaskType task_type = 3;
  TaskStatus status = 4;
  string input_json = 5;
  string output_json = 6;
  int64 created_at = 7;
  int64 updated_at = 8;
}

message AIReport {
  int64 id = 1;
  string object_type = 2;
  int64 object_id = 3;
  string summary = 4;
  string strengths = 5;
  string risks = 6;
  string suggestions = 7;
  int64 created_at = 8;
}

message GenerateStrategyRequest { string prompt = 1; }
message GenerateStrategyResponse { int64 task_id = 1; }

message ExplainStrategyRequest { int64 strategy_id = 1; }
message ExplainStrategyResponse { int64 task_id = 1; }

message OptimizeStrategyRequest { int64 strategy_id = 1; }
message OptimizeStrategyResponse { int64 task_id = 1; }

message GeneratePortfolioRequest { string prompt = 1; }
message GeneratePortfolioResponse { int64 task_id = 1; }

message OptimizePortfolioRequest { int64 portfolio_id = 1; }
message OptimizePortfolioResponse { int64 task_id = 1; }

message AnalyzeBacktestRequest { int64 backtest_id = 1; }
message AnalyzeBacktestResponse { int64 task_id = 1; }

message ChatRequest {
  string message = 1;
  int64 conversation_id = 2;
}
message ChatResponse {
  string reply = 1;
  int64 conversation_id = 2;
}

message GetTaskRequest { int64 task_id = 1; }
message GetTaskResponse { AITask task = 1; }

message GetReportRequest {
  string object_type = 1;
  int64 object_id = 2;
}
message GetReportResponse { AIReport report = 1; }
EOF
echo "ai.proto created"
