# QuantLab

量化策略研究平台 —— 面向个人投资者的 AI 原生量化研究平台。

帮助用户将投资想法快速转化为可验证、可复用、可分享的策略资产。

## 产品定位

QuantLab 是一个面向个人投资者、量化研究者和策略开发者的策略研究平台。用户无需编写复杂代码，即可完成选股策略构建、买卖规则定义、历史回测验证、策略分析优化、策略分享与交流，形成"策略开发 → 策略验证 → 策略分享 → 策略成长"的完整闭环。

## 架构概览

QuantLab 采用 Go + DDD + Clean Architecture + Event Driven Architecture 的微服务架构，所有服务共享统一的 IDL 定义和基础设施层。

```
api/              # IDL 定义（protobuf）
├── common/       # 共享类型（Cursor、Visibility、Status）
├── user/         # 用户服务接口
├── strategy/     # 策略服务接口
├── portfolio/    # 组合服务接口
├── billing/      # 计费服务接口
├── community/    # 社区服务接口
├── ranking/      # 排行榜服务接口
├── notification/ # 通知服务接口
├── ai/           # AI 服务接口
├── backtest/     # 回测引擎接口
├── formula/      # 公式引擎接口
└── market/       # 行情数据服务接口

app/              # 微服务实现
├── user/         # 用户服务
├── strategy/     # 策略服务
├── portfolio/    # 组合服务
├── billing/      # 计费服务
├── community/    # 社区服务
├── ranking/      # 排行榜服务
├── notification/ # 通知服务
├── ai/           # AI 服务
├── backtest/     # 回测引擎
├── formula/      # 公式引擎
└── market/       # 行情数据服务

pkg/              # 共享基础设施
├── response/     # 统一响应格式
├── errors/       # 标准化错误码
├── postgresql/   # 数据库连接
├── redis/        # 缓存客户端
├── kafka/        # 消息队列
├── middleware/   # HTTP 中间件
├── pagination/   # 分页工具
└── logger/       # 结构化日志

Documents/        # 项目文档
├── PRD/          # 产品需求文档
├── DDD/          # 领域驱动设计
├── TD/           # 技术设计文档
├── Standard/     # 开发规范标准
├── RoadMap/      # 路线图
├── Deployment/   # 部署架构
└── API/          # API 规范
```

## 技术栈

| 层 | 技术 |
|---|---|
| 语言 | Go >= 1.24 |
| 框架 | go-zero (zrpc) |
| IDL | Protocol Buffers (gRPC) |
| 数据库 | PostgreSQL / TimescaleDB |
| 缓存 | Redis |
| 消息队列 | Kafka |
| 认证 | JWT |
| 可观测性 | OpenTelemetry |

## 微服务列表

| 服务 | 模块路径 | 优先级 | 说明 |
|---|---|---|---|
| user-service | `app/user` | P0 | 用户注册、登录、认证、关注系统 |
| strategy-service | `app/strategy` | P0 | 策略 CRUD、版本管理、搜索、发布 |
| formula-engine | `app/formula` | P0 | DSL 解析、编译、AST 生成、函数注册中心 |
| market-data-service | `app/market` | P0 | 行情数据、财务数据、因子数据、复权处理 |
| backtest-engine | `app/backtest` | P0 | 回测任务调度、撮合、绩效分析、报告生成 |
| ranking-service | `app/ranking` | P0 | 策略/作者排行榜、历史排名、快照 |
| portfolio-service | `app/portfolio` | P1 | 策略组合管理、权重配置、收益分析 |
| community-service | `app/community` | P1 | 内容 Feed、点赞、收藏、评论 |
| notification-service | `app/notification` | P1 | 站内通知、偏好设置、订阅 |
| ai-service | `app/ai` | P2 | AI 策略生成、解释、优化、回测分析 |
| billing-service | `app/billing` | P2 | 会员、订阅、支付、创作者收益、结算 |

## 快速开始

### 前置条件

- Go >= 1.24
- protoc (Protocol Buffers compiler)
- protoc-gen-go
- protoc-gen-go-grpc
- goctl (go-zero 工具链)

### 安装工具链

```bash
# 安装 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 安装 protoc 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 启动服务

```bash
# 启动 user-service
cd app/user
go run user.go -f etc/user.yaml

# 启动 strategy-service
cd app/strategy
go run strategy.go -f etc/strategy.yaml
```

### 生成代码

```bash
# 从 proto 生成 pb 文件
protoc --proto_path=. \
  --go_out=. --go_opt=module=github.com/agoXQ/QuantLab \
  --go-grpc_out=. --go-grpc_opt=module=github.com/agoXQ/QuantLab \
  api/user/v1/user.proto

# 生成 go-zero zrpc 层
goctl rpc protoc api/user/v1/user.proto \
  --go_out=. --go_opt=module=github.com/agoXQ/QuantLab \
  --go-grpc_out=. --go-grpc_opt=module=github.com/agoXQ/QuantLab \
  --zrpc_out=app/user --style=gozero --module=github.com/agoXQ/QuantLab
```

## 开发规范

项目遵循统一的工程开发标准，详见 [Documents/Standard/](Documents/Standard/)：

- [API 设计标准](Documents/Standard/QuantLab%20API%20Design%20Standard%20v1.0.md)
- [架构治理标准](Documents/Standard/QuantLab%20Architecture%20Governance%20Standard%20v1.0.md)
- [工程开发标准](Documents/Standard/QuantLab%20Engineering%20Development%20Standard%20v1.0.md)
- [数据库设计标准](Documents/Standard/QuantLab%20Database%20Design%20Standard%20v1.0.md)
- [前端开发标准](Documents/Standard/QuantLab%20Frontend%20Development%20Standard%20v1.0.md)

### Git 工作流

```
main
└── develop
    ├── feature/*
    ├── bugfix/*
    ├── release/*
    └── hotfix/*
```

Commit 遵循 Conventional Commit 格式：`feat:` / `fix:` / `refactor:` / `docs:` / `test:` / `chore:`

### 代码分层

每个服务遵循 go-zero 标准 4 层结构：

```
app/<service>/
├── <service>.go                  # main 入口
├── etc/<service>.yaml            # 配置
├── internal/
│   ├── config/config.go          # 配置结构
│   ├── logic/                    # 业务逻辑（每个 RPC 一个文件）
│   ├── server/                   # gRPC 服务注册
│   └── svc/servicecontext.go     # 服务上下文
├── pb/                           # protoc 生成的 gRPC 代码
└── <service>service/             # 客户端调用封装
```

## 项目路线图

### Phase 1 — MVP（4~5 个月）

核心闭环：策略 DSL → 选股 → 回测 → 输出收益率

- [ ] Formula Engine：DSL 解析、指标计算、选股执行
- [ ] Market Data Service：股票基础信息、日线行情、复权处理
- [ ] Backtest Engine：策略回测、收益率、回撤、夏普比率
- [ ] User Service：注册、登录、JWT 认证、个人主页
- [ ] Strategy Service：策略 CRUD、版本管理、发布
- [ ] Ranking Service：收益率排行、胜率排行
- [ ] Portfolio Service：策略组合、组合收益统计

### Phase 2 — 社区与 AI

- [ ] Community Service：策略评论、内容 Feed
- [ ] Notification Service：站内通知
- [ ] AI Service：AI 解释策略、AI 分析回测

### Phase 3 — 商业化

- [ ] Billing Service：会员订阅、支付、创作者收益

## 相关文档

- [平台 PRD](Documents/PRD/QuantLab%20平台PRD.md) — 完整产品需求
- [MVP 路线图](Documents/RoadMap/QuantLab%20MVP%20Roadmap%20v1.0.md) — 研发计划
- [领域模型](Documents/DDD/QuantLab%20Domain%20Model%20v1.0.md) — DDD 聚合设计
- [部署架构](Documents/Deployment/QuantLab%20Deployment%20Architecture%20v1.0.md) — 基础设施规划

## 许可证

MIT
