# milk_tea_go

奶茶点评后端服务（Go）：**Gin + PostgreSQL（pgx 连接池）+ Redis（热点缓存）+ JWT（Access）+ DB 持久化 Refresh（轮换/吊销）+ bcrypt 密码摘要**。

面向 **单机约 5000 并发（读多写少）** 的典型 Web/API 形态：通过 **连接池参数、Redis 缓存、按 IP 限流、优雅关停** 等工程化手段为后续水平扩展打基础；真正极限需结合 **压测、PostgreSQL 调优、前置 Nginx/网关、读写分离** 等继续演进。

## 功能概览

- 用户注册 / 登录 / 刷新 Token / 登出（Refresh 轮换）
- 品类列表、奶茶单品列表/详情、评论列表
- 提交/更新点评（每用户每单品一条，数据库 `UNIQUE(tea_id, user_id)`）
- 排行榜（按 `tea_stats.avg_rating` + `review_count`）
- 健康检查：`/healthz`、`/readyz`（检查 PG + Redis）
- Redis 缓存：单品详情、列表、排行榜（写点评后失效相关键）

## 数据库表（字段完备）

| 表 | 说明 |
|----|------|
| `users` | 用户主表：账号、邮箱、密码摘要、资料、状态、最近登录等 |
| `refresh_tokens` | Refresh Token **仅存 SHA-256 摘要**，支持 family 轮换与吊销 |
| `categories` | 奶茶品类：slug、排序、封面、扩展 JSON 等 |
| `teas` | 单品：品牌/店铺、价格区间、标签数组、糖度、热量、咖啡因、季节款等 |
| `reviews` | 点评：星级、标题、正文、图片 JSON、 helpful、审核状态、IP 等 |
| `tea_stats` | 单品统计：评论数、总分、均分、最近评论时间（由触发器维护） |

迁移脚本位于 `migrations/*.up.sql`，启动时自动执行（已记录在 `schema_migrations`）。

## API（`/api/v1`）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/auth/register` | 否 | 注册 |
| POST | `/auth/login` | 否 | 登录，返回 `access_token` + `refresh_token` |
| POST | `/auth/refresh` | 否 | 刷新（轮换 refresh） |
| POST | `/auth/logout` | 否 | 吊销 refresh |
| GET | `/categories` | 否 | 品类列表 |
| GET | `/teas` | 否 | `page`/`page_size`/`category_id` |
| GET | `/teas/:id` | 否 | 单品详情（异步累计 `view_count`） |
| GET | `/teas/:id/reviews` | 否 | 评论分页 |
| GET | `/leaderboard` | 否 | `limit` |
| GET | `/me` | Bearer | 当前用户 |
| POST | `/teas/:id/reviews` | Bearer | 提交/更新点评 |

请求头：`Authorization: Bearer <access_token>`

## 本地开发

前置：**Go 1.22+**、**PostgreSQL 16+**、**Redis 7+**。

```bash
cp .env.example .env
# 修改 JWT 密钥与 DSN

go run ./cmd/server
```

默认监听 `HTTP_ADDR`（默认 `:8080`）。

## Docker Compose（推荐一键起）

```bash
docker compose up -d --build
```

- API：`http://127.0.0.1:8080`
- PostgreSQL：`localhost:5432`（`milktea/milktea`，库名 `milktea`）
- Redis：`localhost:6379`

**生产环境务必**通过环境变量覆盖 `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET`。

## 云服务器部署（概要）

1. 安装 Docker（或手动安装 Go + PostgreSQL + Redis）。
2. 克隆仓库，配置 `.env` 或 `docker-compose.yml` 环境变量。
3. `docker compose up -d --build`。
4. 前置 **Nginx** 反代到 `8080`，并配置 TLS、超时与 `client_max_body_size`。
5. 依据压测结果调节：`PG_MAX_CONNS`、PostgreSQL `max_connections`、Redis `maxmemory` 与淘汰策略、以及 `RATE_LIMIT_*`。

### 单机 5000 并发（方向性建议）

- **PostgreSQL**：`pgbouncer` 连接池、合理索引、避免 N+1；本服务已用 **单条 SQL + 统计表**。
- **应用**：`PG_MAX_CONNS` / `REDIS_POOL_SIZE` 与机器 CPU/内存匹配；前置 **限流** 保护。
- **Redis**：热点缓存 + 失效策略；必要时排行榜可改为 **定时任务** 从 DB 刷新到 Redis。
- **观测**：接入 Prometheus/OpenTelemetry（后续迭代可加 metrics 中间件）。

## 与前端 `Hello_milkTea` 对接

前端设置 `VITE_API_BASE=https://你的域名`，`VITE_USE_MOCK=false`，并按前端预留路径对齐（如需可加反向代理把 `/api` 指到本服务）。

## 目录结构（精简）

```
cmd/server          # 入口
internal/config     # 环境配置
internal/db         # PostgreSQL 连接池
internal/cache      # Redis 客户端与缓存键
internal/migrate    # SQL 迁移执行器
internal/repo       # 数据访问
internal/handler    # HTTP
internal/middleware # JWT / 限流 / CORS
pkg/jwtutil         # JWT
pkg/password        # bcrypt
migrations          # SQL 脚本
```

## 许可证

MIT
