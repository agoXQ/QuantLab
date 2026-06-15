# QuantLab Frontend Development Standard v1.0

**Document Version:** v1.0
**Project:** QuantLab
**Document Type:** Frontend Development Standard
**Status:** Approved
**Last Updated:** 2026-06

---

# Introduction

## 1.1 Purpose

本文档定义 QuantLab 前端工程开发规范。

目标：

- 保证 UI 风格一致
- 降低前后端协作成本
- 保证前端性能与可维护性
- 支撑量化数据可视化需求
- 规范组件复用与状态管理

---

## 1.2 Scope

适用于：

```Plain Text
Web Application (React SPA)
Admin Dashboard
Strategy Editor
Backtest Visualization
Community Pages
Mobile Web (Responsive)
```

---

## 1.3 Design Principles

### Principle 1：Data-Dense, Not Decorative

量化平台以数据展示为核心。

优先信息密度，而非视觉装饰。

---

### Principle 2：Component First

所有 UI 元素必须封装为可复用组件。

禁止页面内联重复结构。

---

### Principle 3：Type Safety

TypeScript strict mode 强制执行。

禁止 `any` 类型（除非有明确理由并注释）。

---

### Principle 4：Performance Budget

首屏加载 < 3s（3G 网络）。

图表渲染 < 500ms（1000 数据点）。

---

### Principle 5：Accessibility

所有交互元素必须可键盘访问。

图表必须提供数据表格降级方案。

---

# Tech Stack

## 2.1 Core

```Plain Text
React 18+
TypeScript 5+
Vite
```

---

## 2.2 UI Framework

```Plain Text
Ant Design 5+
```

禁止混用其他 UI 库。

---

## 2.3 Visualization

```Plain Text
ECharts 5+
```

统一图表库。

---

## 2.4 State Management

```Plain Text
Zustand (全局状态)
React Query / TanStack Query (服务端状态)
```

---

## 2.5 Routing

```Plain Text
React Router 6+
```

---

## 2.6 HTTP Client

```Plain Text
Axios
```

统一拦截器处理认证、错误、重试。

---

## 2.7 Form Management

```Plain Text
React Hook Form + Zod
```

---

## 2.8 Code Editor

```Plain Text
Monaco Editor
```

用于策略 DSL 编辑。

---

## 2.9 Build & Dev

```Plain Text
Vite
pnpm
ESLint + Prettier
```

---

# Project Structure

## 3.1 Monorepo Structure

```Plain Text
frontend/
├── apps/
│   ├── web/              # 主站
│   └── admin/            # 管理后台
├── packages/
│   ├── ui/               # 共享 UI 组件
│   ├── charts/           # 图表组件
│   ├── hooks/            # 共享 Hooks
│   ├── utils/            # 工具函数
│   ├── types/            # 共享类型定义
│   └── api/              # API Client
├── docs/
└── scripts/
```

---

## 3.2 Page Module Structure

每个页面模块：

```Plain Text
src/pages/strategy/
├── index.tsx             # 页面入口
├── components/           # 页面级组件
├── hooks/                # 页面级 Hooks
├── types.ts              # 页面级类型
└── constants.ts          # 页面级常量
```

---

## 3.3 Component Structure

```Plain Text
src/components/StrategyCard/
├── index.tsx
├── StrategyCard.module.css
├── StrategyCard.test.tsx
└── types.ts
```

---

# Naming Convention

## 4.1 Files

```Plain Text
Component: PascalCase  (StrategyCard.tsx)
Hook:      camelCase   (useStrategyList.ts)
Util:      camelCase   (formatReturn.ts)
Type:      camelCase   (strategy.ts)
Constant:  UPPER_CASE  (RANKING_TYPES.ts)
Style:     *.module.css
```

---

## 4.2 Components

```TypeScript
// 正确
export function StrategyCard() {}
export function BacktestChart() {}
export function RankingTable() {}

// 错误
export function strategyCard() {}
export function Card() {}
```

---

## 4.3 Props

```TypeScript
interface StrategyCardProps {
  strategy: Strategy;
  onCopy?: (id: string) => void;
  isLoading?: boolean;
}
```

---

## 4.4 Event Handlers

```TypeScript
// Props: on + 事件
onSubmit
onCancel
onStrategySelect

// Internal: handle + 事件
handleSubmit
handleStrategyClick
```

---

# TypeScript Standard

## 5.1 Strict Mode

`tsconfig.json`:

```JSON
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true
  }
}
```

---

## 5.2 Type Definition

优先使用 `interface`，需要 union / intersection 时使用 `type`。

```TypeScript
interface Strategy {
  id: string;
  name: string;
  status: StrategyStatus;
  metrics: BacktestMetrics;
}

type StrategyStatus = 'draft' | 'published' | 'archived';
```

---

## 5.3 API Response Types

所有 API 响应必须定义类型：

```TypeScript
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface PaginatedResponse<T> extends ApiResponse<T[]> {
  meta: {
    nextCursor?: string;
    hasMore: boolean;
  };
}
```

---

## 5.4 Discriminated Unions

状态驱动的 UI 使用 discriminated union：

```TypeScript
type BacktestState =
  | { status: 'idle' }
  | { status: 'running'; jobId: string }
  | { status: 'completed'; report: BacktestReport }
  | { status: 'failed'; error: string };
```

---

# Component Design Standard

## 6.1 Component Categories

```Plain Text
Atom:    Button, Input, Tag, Icon
Molecule: StrategyCard, UserAvatar, MetricBadge
Organism: StrategyEditor, BacktestPanel, RankingList
Template: PageLayout, DashboardLayout
Page:     StrategyListPage, BacktestPage
```

---

## 6.2 Component Rules

- 每个组件只做一件事
- Props 不超过 8 个（超过则拆分子组件）
- 禁止在组件内直接调用 API（使用 Hook 封装）
- 展示组件与容器组件分离

---

## 6.3 Compound Components

复杂组件使用 Compound Pattern：

```TypeScript
<StrategyEditor>
  <StrategyEditor.Formula />
  <StrategyEditor.BuyRule />
  <StrategyEditor.SellRule />
  <StrategyEditor.RiskRule />
</StrategyEditor>
```

---

# State Management Standard

## 7.1 Server State

所有 API 数据使用 TanStack Query：

```TypeScript
const { data, isLoading, error } = useQuery({
  queryKey: ['strategy', id],
  queryFn: () => strategyApi.getById(id),
  staleTime: 5 * 60 * 1000,
});
```

---

## 7.2 Client State

全局 UI 状态使用 Zustand：

```TypeScript
interface AppStore {
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}
```

---

## 7.3 Form State

使用 React Hook Form + Zod：

```TypeScript
const schema = z.object({
  name: z.string().min(1).max(100),
  category: z.enum(['value', 'growth', 'momentum']),
});

type FormData = z.infer<typeof schema>;
```

---

## 7.4 State Location Rule

```Plain Text
仅一个组件使用：useState
父子组件共享：props drilling 或 Context
跨页面共享：Zustand
服务端数据：TanStack Query
URL 参数：React Router search params
```

---

# API Integration Standard

## 8.1 API Client

统一 Axios 实例：

```TypeScript
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE,
  timeout: 30000,
});

apiClient.interceptors.request.use(attachAuthToken);
apiClient.interceptors.response.use(handleSuccess, handleError);
```

---

## 8.2 API Module Organization

```Plain Text
src/api/
├── client.ts           # Axios 实例
├── strategy.ts         # 策略相关 API
├── backtest.ts         # 回测相关 API
├── ranking.ts          # 排行榜相关 API
├── community.ts        # 社区相关 API
├── user.ts             # 用户相关 API
├── billing.ts          # 账单相关 API
└── types.ts            # API 类型
```

---

## 8.3 Error Handling

统一错误处理：

```TypeScript
interface ApiError {
  code: number;
  message: string;
  requestId: string;
}

// 全局错误边界
// 业务错误码映射为用户提示
// 网络错误显示重试按钮
```

---

## 8.4 Loading States

所有异步操作必须覆盖三种状态：

```TypeScript
{isLoading && <Skeleton />}
{error && <ErrorState onRetry={refetch} />}
{data && <DataView data={data} />}
```

空数据状态单独处理：

```TypeScript
{data?.length === 0 && <EmptyState />}
```

---

# Visualization Standard

## 9.1 Chart Component Pattern

```TypeScript
interface BacktestChartProps {
  data: ChartDataPoint[];
  benchmark?: ChartDataPoint[];
  height?: number;
  loading?: boolean;
  onTimeRangeChange?: (range: DateRange) => void;
}
```

---

## 9.2 Chart Types

```Plain Text
收益曲线:     Line Chart
回撤曲线:     Area Chart
月收益热力图: Heatmap
持仓分布:     Pie / Treemap
行业暴露:     Bar Chart
排名变化:     Line Chart (多系列)
相关性矩阵:   Heatmap
```

---

## 9.3 ECharts Configuration

统一主题与样式：

```TypeScript
const defaultChartTheme = {
  backgroundColor: 'transparent',
  textStyle: { fontFamily: 'inherit' },
  grid: { top: 40, right: 20, bottom: 30, left: 50 },
  tooltip: { trigger: 'axis' },
  toolbox: {
    feature: {
      dataZoom: {},
      saveAsImage: { pixelRatio: 2 },
    },
  },
};
```

---

## 9.4 Performance

- 超过 5000 数据点使用 `sampling: 'lttb'`
- 超过 10000 数据点使用 Web Worker 预处理
- 图表懒加载（Intersection Observer）
- 禁止同时渲染超过 4 个图表

---

## 9.5 Accessibility

每个图表必须提供数据表格降级：

```TypeScript
<ChartWithTable>
  <ECharts option={option} />
  <Collapse>
    <Table dataSource={rawData} />
  </Collapse>
</ChartWithTable>
```

---

# Strategy Editor Standard

## 10.1 Editor Layout

```Plain Text
+------------------+------------------+
|                  |                  |
|   Formula        |   Context Help   |
|   Editor         |   (Formula Docs) |
|   (Monaco)       |                  |
|                  |                  |
+------------------+------------------+
|   Buy Rule       |   Sell Rule      |
+------------------+------------------+
|   Risk Rule      |   Position Rule  |
+------------------+------------------+
```

---

## 10.2 Monaco Configuration

```TypeScript
const editorOptions = {
  language: 'quantlab-dsl',
  minimap: { enabled: false },
  lineNumbers: 'on',
  fontSize: 14,
  tabSize: 2,
  wordWrap: 'on',
  suggestOnTriggerCharacters: true,
};
```

---

## 10.3 DSL Language Support

- 语法高亮（关键字、函数、变量）
- 自动补全（函数名、变量名、参数提示）
- 实时语法校验
- 悬停文档提示
- 错误标记与快速修复

---

## 10.4 AI Copilot Integration

编辑器内嵌 AI 辅助：

```TypeScript
<FormulaEditor>
  <AICopilot
    onGenerate={(prompt) => aiApi.generateFormula(prompt)}
    onExplain={(formula) => aiApi.explainFormula(formula)}
  />
</FormulaEditor>
```

---

# Responsive Design Standard

## 11.1 Breakpoints

```Plain Text
Mobile:  < 768px
Tablet:  768px - 1024px
Desktop: > 1024px
Wide:    > 1440px
```

---

## 11.2 Layout Strategy

```Plain Text
Desktop:  Sidebar + Content (多列)
Tablet:   Collapsible Sidebar + Content
Mobile:   Bottom Nav + Content (单列)
```

---

## 11.3 Data Table Responsive

移动端表格：

```Plain Text
固定首列（策略名称）
横向滚动
或切换为卡片列表
```

---

# Performance Standard

## 12.1 Bundle Optimization

- 路由级代码分割（`React.lazy`）
- ECharts 按需引入
- Ant Design 按需引入（tree shaking）
- 图片使用 WebP + 懒加载

---

## 12.2 Rendering Optimization

- 列表虚拟滚动（> 100 条）
- `React.memo` 用于纯展示组件
- `useMemo` / `useCallback` 用于计算密集型
- 避免在渲染中创建新对象 / 数组

---

## 12.3 Network Optimization

- API 响应缓存（TanStack Query）
- 预加载关键页面
- 请求去重与合并
- WebSocket 用于实时数据（V2）

---

## 12.4 Performance Monitoring

```Plain Text
LCP < 2.5s
FID < 100ms
CLS < 0.1
```

使用 Lighthouse CI 持续监控。

---

# Testing Standard

## 13.1 Test Types

```Plain Text
Unit Test:       Vitest + React Testing Library
Component Test:  Storybook + Chromatic
E2E Test:        Playwright
```

---

## 13.2 Coverage

```Plain Text
核心业务组件: ≥ 80%
工具函数:     ≥ 90%
页面:         关键路径覆盖
```

---

## 13.3 What to Test

```Plain Text
必须测试:
- 用户交互流程（创建策略 → 回测 → 查看结果）
- 表单验证
- 错误状态展示
- 空数据状态
- 加载状态

不必须测试:
- 第三方库内部逻辑
- 纯样式变更
```

---

## 13.4 Component Test Example

```TypeScript
test('StrategyCard shows backtest metrics when available', async () => {
  render(<StrategyCard strategy={mockStrategy} />);
  expect(screen.getByText('+42.5%')).toBeVisible();
  expect(screen.getByText('Sharpe: 1.85')).toBeVisible();
});

test('StrategyCard shows empty state when no backtest', () => {
  render(<StrategyCard strategy={mockDraftStrategy} />);
  expect(screen.getByText('未回测')).toBeVisible();
});
```

---

# Accessibility Standard

## 14.1 Minimum Requirements

- 所有交互元素可键盘访问
- 表单有明确的 label 关联
- 图表提供数据表格降级
- 颜色不是唯一的信息传递方式
- 页面有合理的 heading 层级

---

## 14.2 ARIA

```TypeScript
<button aria-label="收藏策略" onClick={handleFavorite}>
  <StarIcon />
</button>

<table role="table">
  <thead role="rowgroup">
    <tr role="row">
      <th role="columnheader">策略名称</th>
    </tr>
  </thead>
</table>
```

---

# Security Standard

## 15.1 XSS Prevention

- 禁止 `dangerouslySetInnerHTML`
- 用户输入必须转义
- 使用 Content Security Policy

---

## 15.2 Token Management

- JWT 存储在 httpOnly cookie（优先）
- 或内存中（避免 localStorage）
- 过期自动刷新

---

## 15.3 Sensitive Data

- 禁止在 URL 中传递敏感参数
- 禁止在 console.log 中输出用户数据
- 禁止在前端存储明文密码

---

# Build & Deployment

## 16.1 Build

```Plain Text
pnpm build
```

输出至 `dist/`。

---

## 16.2 Environment Variables

```Plain Text
VITE_API_BASE=https://api.quantlab.io
VITE_WS_BASE=wss://ws.quantlab.io
VITE_SENTRY_DSN=...
```

---

## 16.3 Deployment

```Plain Text
CDN + Nginx 静态托管
Docker + Nginx（可选）
```

---

## 16.4 CI/CD

```Plain Text
Lint → Type Check → Test → Build → Deploy
```

---

# Code Review Checklist

合并前必须检查：

✓ TypeScript strict 通过
✓ 无 `any` 类型（或有注释说明）
✓ 组件有 Loading / Error / Empty 状态
✓ API 调用使用 TanStack Query
✓ 图表有数据表格降级
✓ 表单有完整验证
✓ 列表 > 100 条使用虚拟滚动
✓ 无 console.log 残留
✓ 无硬编码的 API 地址
✓ 响应式布局在 Mobile / Tablet / Desktop 均正常

---

# Final Principle

前端是量化平台的数据展示层。

设计优先服务于数据的清晰呈现与高效交互，而非视觉效果。

任何前端决策必须满足：

Data Clarity > Interaction Efficiency > Visual Consistency > Aesthetic Preference

未经 Code Review 不得合并至主分支。
