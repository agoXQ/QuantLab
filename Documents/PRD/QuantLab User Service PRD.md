# QuantLab User Service PRD

Version：1\.0

Module：User Service

Priority：P0

Status：Draft

Owner：Product Team

Last Update：2026

---

# 模块定位

User Service 是 QuantLab 的统一身份与用户中心。

负责：

- 用户身份管理 

- 认证授权 

- 用户资料管理 

- 创作者体系 

- 会员体系 

- 权限体系 

- 用户荣誉体系 

不负责：

- 策略管理 

- 回测执行 

- 排行榜计算 

---

# 产品目标

建立完整的量化创作者生态。

实现：

```Plain Text
用户
↓
策略创作者
↓
策略作者
↓
优秀作者
↓
策略大师
↓
社区KOL
```

---

# 用户体系架构

```Plain Text
User Service

├── Identity Center
├── Profile Center
├── Permission Center
├── Membership Center
├── Creator Center
├── Honor Center
├── Follow Center
└── Security Center
```

---

# 用户生命周期

```Plain Text
游客

↓

注册用户

↓

活跃用户

↓

策略作者

↓

优秀作者

↓

认证作者

↓

策略大师
```

---

# Identity Center

身份中心

负责：

- 注册 

- 登录 

- 第三方登录 

- Token管理 

- Session管理 

---

# 注册方式

支持：

邮箱注册

手机号注册

用户名注册

---

未来支持：

微信登录

Google登录

GitHub登录

Apple登录

---

# 登录方式

支持：

密码登录

验证码登录

扫码登录（V2）

OAuth登录

---

# 用户资料

## 基础资料

昵称

头像

简介

所在地

个人主页

注册时间

---

## 投资资料

投资年限

投资风格

擅长市场

擅长策略

---

示例：

```Plain Text
投资风格：

价值投资

ETF轮动

趋势交易
```

---

# 用户等级体系

目的：

鼓励持续活跃

---

等级：

```Plain Text
Lv1 新人

Lv2 研究员

Lv3 策略作者

Lv4 高级研究员

Lv5 策略大师

Lv6 Quant Expert
```

---

# 经验值体系

经验来源：

创建策略

执行回测

发布策略

获得点赞

获得收藏

进入榜单

---

# 创作者中心

核心模块。

---

用户成为：

Strategy Creator

后获得：

发布策略

创建专栏

参与榜单

获得粉丝

---

# 作者主页

展示：

头像

简介

粉丝数

策略数

回测次数

获赞数

上榜次数

---

# 作者统计

统计：

公开策略数量

总回测次数

平均收益率

平均Sharpe

策略复制次数

---

# 策略大师认证

认证条件：

公开策略 ≥ 10

粉丝 ≥ 1000

历史上榜 ≥ 20次

综合评分 ≥ 90

---

认证标识：

```Plain Text
✓ Strategy Master
```

---

# Follow Center

用户关注体系（数据归属方）。

Community Service 基于本模块的关注关系驱动 Feed 与通知。

---

支持：

关注作者

取消关注

查看粉丝

查看关注

---

# 关注后获得

策略发布通知

排行榜变化通知

文章更新通知

---

# 用户荣誉体系

荣誉类型：

新人王

月度冠军

年度冠军

百强策略作者

人气作者

策略大师

---

# 勋章系统

示例：

```Plain Text
🥇 百强策略

🏆 年度冠军

🔥 热门作者

⭐ 策略大师
```

---

# 权限体系

RBAC模型

---

角色：

Guest

User

Pro

Creator

Master

Admin

SuperAdmin

---

# Guest权限

允许：

浏览公开策略

查看排行榜

查看作者主页

---

禁止：

创建策略

回测策略

评论

收藏

---

# User权限

允许：

创建策略

执行回测

评论

点赞

收藏

关注

---

# Pro权限

额外获得：

高级回测

更长历史数据

更多策略数量

AI功能

参数优化

---

# Creator权限

额外获得：

发布策略

进入排行榜

创建专栏

策略推广

---

# Master权限

额外获得：

认证标识

优先推荐

专属主页

高级统计分析

---

# Membership Center

会员体系

---

等级：

Free

Pro

Master

Enterprise

---

# Free

限制：

策略数 20

回测次数 50/月

历史数据 3年

---

# Pro

限制：

策略数 200

回测次数 1000/月

历史数据 10年

---

# Master

限制：

无限策略

无限回测

高级因子

高级AI

---

# Enterprise

支持：

团队管理

策略共享

团队排行榜

私有部署

API接入

---

# 配额体系

控制：

策略数量

回测次数

AI调用次数

数据下载次数

API调用次数

---

# 安全中心

支持：

密码管理

登录记录

设备管理

安全验证

双因素认证（V2）

---

# 风险控制

检测：

异常登录

异常设备

异常API调用

批量注册

机器人行为

---

# 用户行为统计

记录：

登录次数

活跃天数

策略创建数

回测次数

评论数

收藏数

---

# 用户画像

构建：

投资风格画像

策略偏好画像

市场偏好画像

风险偏好画像

---

用途：

推荐系统

AI助手

策略推荐

---

# 用户通知偏好

支持：

站内信

邮件

短信

Webhook（V2）

---

# 用户数据导出

支持：

策略导出

回测导出

收藏导出

关注导出

---

满足：

数据可携带性

---

# 用户注销

支持：

申请注销

冷静期

数据归档

账号删除

---

# MVP范围

必须实现：

√ 注册登录

√ 用户资料

√ 权限体系

√ 会员体系

√ 作者主页

√ 粉丝体系

√ 荣誉体系

√ 配额管理

---

暂缓实现：

× 企业组织

× 双因素认证

× API开放平台

× 团队空间

---

# KPI

新增用户数

活跃用户数

创作者数量

认证作者数量

会员转化率

用户留存率

关注关系数

---

# 核心原则

原则1

用户不仅是投资者，更是策略创作者

---

原则2

策略资产属于用户

---

原则3

所有荣誉必须来源于真实数据

---

原则4

权限体系与会员体系解耦

---

原则5

用户成长路径必须清晰

---

原则6

鼓励创作优于鼓励消费

---

# 用户模型（核心实体）

这里建议在架构层面直接定义：

```Plain Text
User:
  id:
  nickname:
  avatar:
  level:
  membership:
  creator_status:
  verified_status:
  followers:
  following:
  strategy_count:
  backtest_count:
  ranking_count:
  honor_count:
  created_at:
```



