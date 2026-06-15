-- ============================================================
-- QuantLab — Database Initialization
-- ============================================================
-- 参考：Documents/Standard/QuantLab Database Design Standard v1.0.md
-- 每个微服务拥有独立数据库（One Database Per Service）
-- ============================================================

-- 平台管理库（已由 POSTGRES_DB 自动创建）
-- quantlab_platform

-- 用户服务
CREATE DATABASE quantlab_user
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 策略服务
CREATE DATABASE quantlab_strategy
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 组合服务
CREATE DATABASE quantlab_portfolio
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 计费服务
CREATE DATABASE quantlab_billing
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 社区服务
CREATE DATABASE quantlab_community
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 排行榜服务
CREATE DATABASE quantlab_ranking
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 通知服务
CREATE DATABASE quantlab_notification
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- AI 服务
CREATE DATABASE quantlab_ai
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 回测引擎
CREATE DATABASE quantlab_backtest
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 公式引擎
CREATE DATABASE quantlab_formula
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';

-- 行情数据服务（使用 TimescaleDB）
CREATE DATABASE quantlab_market_data
  WITH ENCODING 'UTF8'
  LC_COLLATE = 'en_US.UTF-8'
  LC_CTYPE = 'en_US.UTF-8';
