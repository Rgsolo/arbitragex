# ArbitrageX 项目开发指南

**版本**: v2.2.0 (精简版)
**最后更新**: 2026-01-08
**维护人**: yangyangyang

---

## ⚠️ 重要经验教训（必读）

### Phase 2 问题总结（2026-01-08）

#### 问题发现

在 Phase 2 完成后，我们发现了两个**关键问题**：

**问题 1：未遵循 go-zero 最佳实践**
- ❌ **现象**：手动编写代码，未使用 goctl 工具生成代码结构
- ❌ **根本原因**：没有先创建 `.api` 文件，导致缺少 `internal/handler/`、`internal/logic/` 等标准层
- ❌ **后果**：代码不符合 go-zero 规范，无法充分利用框架生态工具

**问题 2：缺少阶段验证机制**
- ❌ **现象**：阶段完成后没有自动验证服务能否正常启动
- ❌ **根本原因**：没有建立自动化的验证流程
- ❌ **后果**：质量问题可能延续到后续阶段，修复成本更高

#### 经验教训

**1. go-zero 开发必须遵循的流程**（强制要求）

```bash
# ✅ 正确的流程（强制执行）
1. 编写 .api 文件（API 定义）
2. 使用 goctl 生成代码
   goctl api go -api api/price.api -dir ./cmd/price -style go_zero
3. 在生成的代码基础上添加业务逻辑
4. 运行验证确保质量

# ❌ 错误的流程（禁止执行）
1. 手动编写 main.go
2. 手动创建 internal/config、internal/svc 等
3. 手动实现 handler 和 logic
```

**关键原则**：
- ✅ **API 定义先行**：所有代码必须从 .api 文件开始
- ✅ **工具生成代码**：使用 goctl 工具生成标准结构
- ✅ **保留现有配置**：合并现有的 config、svc、types，不替换
- ✅ **最小化 API**：Phase 2 只创建健康检查接口，业务 API 延后到 Phase 3/4

**2. 每个阶段必须验证**（强制要求）

```bash
# ✅ 阶段完成后必须执行的验证
make verify-stage    # 完整验证（9 项检查）
make verify-quick    # 快速验证（编译+测试）
make check-startup   # 检查服务启动
```

**验证清单**：
- ✅ Go 版本检查（>= 1.21）
- ✅ goctl 工具检查（>= 1.9.0）
- ✅ 项目结构检查（必需目录存在）
- ✅ API 文件检查（.api 文件存在且语法正确）
- ✅ 依赖下载（go mod download）
- ✅ 代码格式化（gofmt）
- ✅ 编译检查（所有服务编译成功）
- ✅ 单元测试（所有测试通过）
- ✅ 测试覆盖率（>= 70%）

**3. 并行开发后的清理工作**

在生成新代码之前，必须：
1. ✅ 删除不符合规范的代码
2. ✅ 删除过时的文档和报告
3. ✅ 总结经验教训并更新 CLAUDE.md
4. ✅ 更新 .progress.json 记录问题

#### 预防措施

**为防止类似问题再次发生，采取以下措施**：

1. **流程控制**
   - 每个阶段开始前，先阅读 CLAUDE.md 的经验教训
   - 使用 goctl 工具生成代码，而不是手动编写
   - 阶段完成后立即运行验证

2. **代码审查**
   - 每个 Agent 完成任务后，检查是否符合最佳实践
   - 生成代码前，确认 .api 文件存在且正确

3. **文档更新**
   - 发现问题后，立即更新 CLAUDE.md
   - 定期回顾经验教训，避免重复错误

#### 参考资源

- [go-zero API 服务开发指南](https://go-zero.dev/docs/tutorials/cli/api)
- [goctl RPC 工具使用](https://go-zero.dev/docs/tutorials/cli/rpc)
- [goctl Model 工具使用](https://go-zero.dev/docs/tutorials/cli/model)
- [go-zero 官方文档](https://go-zero.dev/en/docs/concepts/overview)
- [go-zero GitHub 仓库](https://github.com/zeromicro/go-zero)

---

## 📘 go-zero 完整开发流程（基于官方文档）

### 概述

本章节基于 go-zero 官方文档总结了完整的开发流程，包括：
1. **API 服务开发**（HTTP REST API）
2. **RPC 服务开发**（gRPC 微服务）
3. **数据库 Model 开发**（MySQL/PostgreSQL/Mongo）

### 一、API 服务开发流程

#### 1.1 安装 goctl 工具

```bash
# 安装最新版本的 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 验证安装
goctl --version
# 输出示例：goctl version 1.9.2 darwin/arm64
```

**重要提示**：
- ✅ goctl 版本应 >= 1.9.0（与 go-zero v1.9.4 配套）
- ✅ 确保 `$GOPATH/bin` 在 PATH 中

#### 1.2 创建 API 定义文件

**步骤**：

1. **创建 api 目录**
   ```bash
   mkdir -p api
   ```

2. **编写 .api 文件**
   ```api
   syntax = "v1"

   info(
       title: "Price Monitor API"
       desc: "价格监控服务"
       author: "yangyangyang"
       version: "v1.0"
   )

   type (
       // Request 请求结构体
       Request {
           Name string `json:"name"`
       }

       // Response 响应结构体
       Response {
           Message string `json:"message"`
       }
   )

   @server(
       prefix: /api
   )
   service price-api {
       @doc "健康检查"
       @handler healthCheck
       get /health(Request) returns(Response)
   }
   ```

**.api 文件语法说明**：
- `syntax`: API 语法版本（固定为 "v1"）
- `info`: 服务元信息（title、desc、author、version）
- `type`: 定义请求和响应的结构体
- `@server`: 服务级别配置（prefix、middleware 等）
- `@handler`: 处理器函数名
- `@doc`: 接口文档说明

#### 1.3 使用 goctl 生成 API 服务代码

```bash
# 生成 API 服务代码
goctl api go -api api/price.api -dir ./cmd/price -style go_zero

# 参数说明：
# -api: API 定义文件路径（必需）
# -dir: 代码输出目录（默认当前目录）
# -style: 文件命名风格（默认 gozero，可选 go_zero、goZero）
```

**生成的目录结构**：
```
cmd/price/
├── main.go                 # 主入口文件
├── etc/
│   └── price.yaml         # 配置文件
└── internal/
    ├── config/
    │   └── config.go      # 配置定义
    ├── handler/
    │   ├── routes.go      # 路由注册
    │   └── healthcheckhandler.go  # 处理器
    ├── logic/
    │   └── healthchecklogic.go    # 业务逻辑
    ├── svc/
    │   └── servicecontext.go      # 服务上下文
    └── types/
        └── types.go       # 类型定义
```

**生成的关键文件说明**：

1. **main.go** - 服务入口
   ```go
   func main() {
       flag.Parse()

       var c config.Config
       conf.MustLoad(*configFile, &c)

       ctx := svc.NewServiceContext(c)
       server := rest.MustNewServer(c.RestConf)

       // 注册路由
       handler.RegisterHandlers(server, ctx)

       server.Start()
   }
   ```

2. **handler/routes.go** - 路由注册
   ```go
   func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
       server.AddRoutes(
           []rest.Route{
               {
                   Method:  http.MethodGet,
                   Path:    "/api/health",
                   Handler: healthCheckHandler(serverCtx),
               },
           },
       )
   }
   ```

3. **logic/xxxlogic.go** - 业务逻辑
   ```go
   type HealthCheckLogic struct {
       logx.Logger
       ctx context.Context
       svcCtx *svc.ServiceContext
   }

   func (l *HealthCheckLogic) HealthCheck(req *types.Request) (resp *types.Response, err error) {
       // 业务逻辑实现
       return
   }
   ```

#### 1.4 验证 API 文件语法

```bash
# 验证 .api 文件语法
goctl api validate --api api/price.api

# 格式化 .api 文件
goctl api format --dir api/
```

#### 1.5 生成 Swagger 文档（可选）

```bash
# 生成 Swagger 文档
goctl api swagger --api api/price.api --dir ./docs
```

---

### 二、RPC 服务开发流程

#### 2.1 创建 Proto 文件

**步骤**：

1. **创建 rpc 目录**
   ```bash
   mkdir -p rpc
   ```

2. **编写 .proto 文件**
   ```protobuf
   syntax = "proto3";

   package greet;
   option go_package = "./greet";

   // 请求消息
   message Request {
       string name = 1;
   }

   // 响应消息
   message Response {
       string message = 2;
   }

   // Greeting 服务
   service Greeter {
       rpc SayHello (Request) returns (Response);
   }
   ```

#### 2.2 使用 goctl 生成 RPC 服务代码

```bash
# 方式 1：从 proto 文件生成（推荐）
goctl rpc protoc greet.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --client=true

# 方式 2：快速创建 RPC 服务
goctl rpc new greet

# 参数说明：
# --go_out: protobuf Go 代码输出目录
# --go-grpc_out: gRPC Go 代码输出目录
# --zrpc_out: go-zero RPC 代码输出目录
# --client: 是否生成客户端代码（默认 true）
```

**生成的目录结构**：
```
.
├── greet.proto          # proto 定义文件
├── greet.go             # pb 代码（--go_out）
├── greet_grpc.pb.go     # pb grpc 代码（--go-grpc_out）
└── greet/               # RPC 服务代码（--zrpc_out）
    ├── etc/
    │   └── greet.yaml   # 配置文件
    ├── greet.go        # 服务入口
    ├── internal/
    │   ├── config/
    │   ├── logic/
    │   ├── server/
    │   └── types/
    └── pb/              # 生成的 pb 代码
```

#### 2.3 RPC 服务配置示例

**greet.yaml**：
```yaml
Name: greet.rpc
ListenOn: 0.0.0.0:8080

# 数据库配置
Mysql:
  DataSource: user:password@tcp(127.0.0.1:3306)/dbname

# Redis 配置
Redis:
  Host: localhost:6379
  Type: node
  Pass: ""
```

---

### 三、数据库 Model 开发流程

#### 3.1 MySQL Model 生成

**方式 1：从 SQL 文件生成**

```bash
# 从 SQL DDL 文件生成 Model
goctl model mysql ddl \
  --src=./scripts/mysql/*.sql \
  --dir=./internal/model \
  --cache=true \
  --strict=false

# 参数说明：
# --src: SQL 文件路径或通配符（必需）
# --dir: 代码输出目录（必需）
# --cache: 是否生成带缓存的代码（默认 false）
# --strict: 严格模式（default false）
# --ignore-columns: 忽略的字段（create_at, update_at 等）
# --prefix: 缓存 key 前缀（默认 "cache"）
```

**方式 2：从数据库连接生成**

```bash
# 从数据库连接生成 Model
goctl model mysql datasource \
  --url="user:password@tcp(127.0.0.1:3306)/database" \
  --table="user,trade,order" \
  --dir=./internal/model \
  --cache=true

# 参数说明：
# --url: 数据库连接字符串（必需）
# --table: 表名或通配符（必需）
# --dir: 代码输出目录（默认当前目录）
# --cache: 是否生成带缓存的代码
```

**生成的文件**（每张表）：

1. **xxxmodel.go** - Model 定义（可编辑）
   ```go
   type (
       UserModel interface {
           userModel
           // 自定义方法接口
           FindByName(ctx context.Context, name string) (*User, error)
       }

       customUserModel struct {
               *defaultUserModel
       }
   )

   // NewUserModel 创建 Model
   func NewUserModel(conn sqlx.SqlConn) UserModel {
       return &customUserModel{
           defaultUserModel: newUserModel(conn),
       }
   }
   ```

2. **xxxmodelgen.go** - Model 生成（DO NOT EDIT）
   ```go
   // Code generated by goctl. DO NOT EDIT.
   type defaultUserModel struct {
       conn sqlx.SqlConn
       table string
   }

   func (m *defaultUserModel) Insert(ctx context.Context, data *User) error {
       // 自动生成的插入方法
   }
   ```

3. **types.go** - 类型定义（可编辑）
   ```go
   type User struct {
       Id       int64  `json:"id"`
       Name     string `json:"name"`
       Age      int32  `json:"age"`
       CreateTime time.Time `json:"createTime"`
       UpdateTime time.Time `json:"updateTime"`
   }
   ```

4. **error.go** - 错误定义

#### 3.2 PostgreSQL Model 生成

```bash
# 从数据库连接生成
goctl model pg datasource \
  --url="postgres://user:password@127.0.0.1:5432/dbname?sslmode=disable" \
  --table="user,trade" \
  --schema="public" \
  --dir=./internal/model \
  --cache=true

# 参数说明：
# --schema: schema 名称（默认 "public"）
```

#### 3.3 Mongo Model 生成

```bash
# 生成 Mongo Model
goctl model mongo \
  --type=User,Order \
  --dir=./internal/model \
  --cache=false

# 参数说明：
# --type: 结构体类型名称（必需，多个用逗号分隔）
# --easy: 是否暴露集合名称变量（default false）
```

#### 3.4 MySQL 类型映射关系

**strict = true 时**：

| MySQL 类型 | 是否 nullable | Golang 类型 |
|-----------|-------------|------------|
| tinyint | NO | uint64 |
| tinyint | YES | sql.NullInt64 |
| int | NO | uint64 |
| int | YES | sql.NullInt64 |
| bigint | NO | uint64 |
| bigint | YES | sql.NullInt64 |
| float | NO | float64 |
| double | NO | float64 |
| decimal | NO | float64 |
| date | NO | time.Time |
| datetime | NO | time.Time |
| timestamp | NO | time.Time |
| varchar | NO | string |
| text | NO | string |
| json | NO | string |
| bool | NO | bool |

**注意事项**：
- `strict = true`：unsigned 字段映射到 uint64
- `strict = false`：unsigned 字段也映射到 int64
- 默认忽略字段：`create_at`, `created_at`, `create_time`, `update_at`, `updated_at`

#### 3.5 自定义 Model 方法

**场景**：需要添加自定义查询方法

**步骤**：

1. 编辑 `xxxmodel.go`
2. 在接口中添加方法签名
3. 实现 `customXxxModel` 结构体

**示例**：
```go
type (
    UserModel interface {
        userModel
        // 自定义方法
        FindByStatus(ctx context.Context, status int) ([]*User, error)
    }

    customUserModel struct {
        *defaultUserModel
    }
)

func (m *customUserModel) FindByStatus(ctx context.Context, status int) ([]*User, error) {
    var resp []*User
    query := fmt.Sprintf("SELECT %s FROM %s WHERE status = ?", userRows, m.table)
    err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
    return resp, err
}
```

---

### 四、完整开发流程示例

#### 4.1 场景：开发一个用户管理 API 服务

**步骤 1：编写 API 定义**

```api
# api/user.api
syntax = "v1"

info(
    title: "User Management API"
    desc: "用户管理服务"
    author: "yangyangyang"
    version: "v1.0"
)

type (
    // CreateUserRequest 创建用户请求
    CreateUserRequest {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    // CreateUserResponse 创建用户响应
    CreateUserResponse {
        Id   int64  `json:"id"`
        Name string `json:"name"`
    }

    // GetUserRequest 获取用户请求
    GetUserRequest {
        Id int64 `path:"id"`
    }

    // GetUserResponse 获取用户响应
    GetUserResponse {
        Id    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
)

@server(
    prefix: /api
    group: user
    middleware: Auth
)
service user-api {
    @doc "创建用户"
    @handler createUser
    post /user(CreateUserRequest) returns(CreateUserResponse)

    @doc "获取用户"
    @handler getUser
    get /user/:id(GetUserRequest) returns(GetUserResponse)
}
```

**步骤 2：生成 API 服务代码**

```bash
goctl api go -api api/user.api -dir ./cmd/user -style go_zero
```

**步骤 3：初始化 goctl 配置（可选）**

```bash
# 初始化配置文件
goctl config init

# 编辑 goctl.yaml 自定义类型映射
# model:
#   types_map:
#     varchar:
#       type: string
```

**步骤 4：生成数据库 Model**

```bash
# 从 SQL 文件生成
goctl model mysql ddl \
  --src=./scripts/mysql/user.sql \
  --dir=./internal/model \
  --cache=true
```

**步骤 5：实现业务逻辑**

编辑 `internal/logic/createuserlogic.go`：
```go
func (l *CreateUserLogic) CreateUser(req *types.CreateUserRequest) (resp *types.CreateUserResponse, err error) {
    // 1. 参数验证
    if req.Name == "" {
        return nil, errors.New("name cannot be empty")
    }

    // 2. 调用 Model 层插入数据
    userId, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password, // 实际应该加密
    })
    if err != nil {
        return nil, err
    }

    // 3. 返回结果
    return &types.CreateUserResponse{
        Id:   userId,
        Name: req.Name,
    }, nil
}
```

**步骤 6：运行验证**

```bash
# 编译
make build

# 运行测试
make test

# 启动服务
./bin/user -f cmd/user/etc/user.yaml

# 测试 API
curl -X POST http://localhost:8888/api/user \
  -H "Content-Type: application/json" \
  -d '{"name":"test","email":"test@example.com","password":"123456"}'
```

---

### 五、goctl 常用命令速查表

#### 5.1 API 相关

| 命令 | 说明 | 示例 |
|------|------|------|
| `goctl api new` | 快速创建 API 服务 | `goctl api new service-name` |
| `goctl api go` | 生成 API 服务代码 | `goctl api go -api api/user.api -dir .` |
| `goctl api validate` | 验证 API 文件 | `goctl api validate --api api/user.api` |
| `goctl api format` | 格式化 API 文件 | `goctl api format --dir api/` |
| `goctl api doc` | 生成 Markdown 文档 | `goctl api doc --dir api/ -o ./docs` |
| `goctl api swagger` | 生成 Swagger 文档 | `goctl api swagger --api api/user.api` |

#### 5.2 RPC 相关

| 命令 | 说明 | 示例 |
|------|------|------|
| `goctl rpc new` | 快速创建 RPC 服务 | `goctl rpc new greet` |
| `goctl rpc protoc` | 生成 RPC 服务代码 | `goctl rpc protoc greet.proto --go_out=./types --go-grpc_out=./types --zrpc_out=.` |
| `goctl rpc template` | 生成 proto 模板 | `goctl rpc template -o greet.proto` |

#### 5.3 Model 相关

| 命令 | 说明 | 示例 |
|------|------|------|
| `goctl model mysql datasource` | 从数据库生成 MySQL Model | `goctl model mysql datasource --url="..." --table="user" --dir ./model` |
| `goctl model mysql ddl` | 从 SQL 文件生成 MySQL Model | `goctl model mysql ddl --src=./sql/*.sql --dir ./model` |
| `goctl model pg datasource` | 从数据库生成 PostgreSQL Model | `goctl model pg datasource --url="..." --table="user" --dir ./model` |
| `goctl model mongo` | 生成 Mongo Model | `goctl model mongo --type=User --dir ./model` |

#### 5.4 配置相关

| 命令 | 说明 | 示例 |
|------|------|------|
| `goctl config init` | 初始化配置文件 | `goctl config init` |
| `goctl env` | 查看 goctl 环境变量 | `goctl env` |
| `goctl env -w` | 设置环境变量 | `goctl env -w GOCTL_EXPERIMENTAL=on` |

---

### 六、最佳实践总结

#### 6.1 API 服务开发最佳实践

✅ **DO（推荐做法）**：
1. 先编写 .api 文件定义接口
2. 使用 goctl 生成代码结构
3. 在生成的 logic 层编写业务逻辑
4. 使用 goctl api validate 验证 API 语法
5. 定期使用 goctl api format 格式化 API 文件

❌ **DON'T（避免做法）**：
1. 手动创建 main.go 和 handler
2. 手动定义路由
3. 跳过 .api 文件直接编写代码
4. 修改 goctl 生成的代码（标记为 DO NOT EDIT 的文件）

#### 6.2 RPC 服务开发最佳实践

✅ **DO（推荐做法）**：
1. 使用 protobuf 定义服务接口
2. 使用 goctl rpc protoc 生成代码
3. 在生成的逻辑层实现业务
4. 合理拆分 proto 文件（按业务域）

❌ **DON'T（避免做法）**：
1. 手动编写 gRPC 服务代码
2. 在不同 proto 文件中引用 message（不支持）
3. 忘记添加 -client 参数（如需客户端）

#### 6.3 Model 开发最佳实践

✅ **DO（推荐做法）**：
1. 使用 goctl model 生成 Model 代码
2. 在 xxxmodel.go 中添加自定义方法
3. 不要修改 xxxmodelgen.go（DO NOT EDIT）
4. 使用 --cache 参数生成带缓存的代码
5. 使用 --strict 模式确保类型安全

❌ **DON'T（避免做法）**：
1. 手动编写 CRUD 操作
2. 在 xxxmodelgen.go 中添加代码
3. 忘记处理 nullable 类型
4. 忽略字段命名冲突

#### 6.4 项目结构最佳实践

**推荐的微服务架构**：
```
project/
├── api/                    # API 定义文件
│   ├── user.api
│   ├── order.api
│   └── product.api
├── cmd/                   # 服务入口
│   ├── user/              # 用户服务（API Gateway）
│   ├── user-rpc/          # 用户服务（RPC）
│   ├── order/             # 订单服务
│   └── order-rpc/         # 订单服务（RPC）
├── internal/              # 内部实现
│   ├── config/            # 配置
│   ├── model/             # 数据库 Model
│   ├── middleware/        # 中间件
│   └── types/             # 类型定义
├── rpc/                   # RPC 定义
│   ├── user.proto
│   └── order.proto
└── scripts/               # 脚本
    └── mysql/            # SQL 文件
```

---

**文档版本**: v2.3.0
**最后更新**: 2026-01-08
**更新内容**: 添加完整的 go-zero 开发流程（API + RPC + Model）

---

## 📌 快速启动

**每次启动项目时，按顺序执行**（2 分钟）：

### 1. 检查项目进度 ⏱️ (30 秒)
```bash
# 查看项目整体进度
cat .progress.json | jq '.current_phase, .overall_progress'

# 查看并行任务状态
cat .parallel-tasks.json | jq '.parallel_tasks[] | {task_id, name, status}'
```

### 2. 检查环境 ⏱️ (30 秒)
```bash
# 检查 Docker 容器
docker-compose ps

# 检查数据库连接
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex -e "SHOW TABLES;"
```

### 3. 恢复未完成任务 ⏱️ (根据任务数量)
```bash
# 如果有未完成的并行任务，告诉 Claude Code：
# "请恢复 .parallel-tasks.json 中的未完成任务"
```

### 4. 开始工作 🚀
```bash
# 查看当前任务
cat .progress.json | jq '.next_steps'

# 开始开发
# （根据 next_steps 指示进行）
```

---

## 1. 项目简介

**ArbitrageX** 是一个专业的加密货币跨交易所套利交易系统，支持在 CEX 和 DEX 之间进行自动化套利交易。

### 开发者信息
- **角色**: 区块链后端开发工程师
- **主要语言**: Go 1.21+, Java, TypeScript
- **框架**: Go 使用 go-zero v1.9.4+
- **交流语言**: 中文
- **工作目录**: `/Users/yangyangyang/code/cc/ArbitrageX`

### 核心技术栈
- **后端**: Go 1.21+ + go-zero v1.9.4+
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **区块链**: Ethereum, BSC
- **CEX**: Binance, OKX, Bybit
- **DEX**: Uniswap, SushiSwap
- **部署**: Docker, Docker Compose, Kubernetes

---

## 2. 项目文档结构

```
docs/
├── requirements/           # PRD 文档（已重构）
│   ├── PRD_Core.md
│   ├── PRD_Technical.md
│   └── Strategies/        # 策略文档
├── design/                # 技术设计文档（25 个文档，已重构）
│   ├── Architecture/      # 系统架构
│   ├── TechStack/         # 技术栈详情
│   ├── Modules/           # 模块设计
│   ├── Database/          # 数据库设计
│   ├── Deployment/        # 部署设计
│   └── Monitoring/        # 监控设计
├── development/           # 开发相关文档（新增）
│   ├── PARALLEL_DEVELOPMENT.md  # 并行开发框架
│   ├── TASK_RECOVERY.md          # 任务恢复机制
│   └── CODING_STANDARDS.md       # 详细代码规范（待创建）
├── risk/                  # 风险管理文档
└── config/                # 配置文件设计（已更新 MySQL + go-zero）
```

**📖 文档阅读顺序**：
1. 新手入门：`docs/design/Architecture/README.md`
2. 技术栈：`docs/design/TechStack/README.md`
3. 模块设计：`docs/design/Modules/README.md`
4. 并行开发：`docs/development/PARALLEL_DEVELOPMENT.md`

---

## 3. 代码规范（精简版）

### 命名规范

**Go 语言**：
- 包名：小写单词，不使用下划线或驼峰
  ```go
  package price  // ✓
  package priceMonitor  // ✗
  ```
- 常量：驼峰命名或全大写+下划线
- 变量/函数：驼峰命名
- 接口：通常以 -er 结尾（如 `PriceMonitorer`）

### 格式规范
- **Go**: 使用 `gofmt` 或 `goimports`
- 缩进：Go 使用 tab，其他语言 2-4 空格
- 每行最大长度：120 字符

### 注释规范
- **必须添加注释的场景**：
  1. 所有公开的 API（函数、方法、结构体）
  2. 复杂的业务逻辑
  3. 关键算法和数据处理
  4. TODO 和 FIXME
  5. 文件级别注释
- **注释语言**：中文（专业术语保留英文）

### 测试要求
- **所有代码必须编写单元测试**
- 核心业务逻辑测试覆盖率 ≥ 80%
- 使用表驱动测试（Table-Driven Tests）

**📖 详细规范**：参考 `docs/development/CODING_STANDARDS.md`（待创建）

---

## 4. 并行开发工作模式

### 概述

**ArbitrageX 使用多 Agent 并行协作开发模式**，模拟真实团队协作，提高开发效率。

**核心理念**：
- ✅ 多个 Agent 同时工作，互不干扰
- ✅ 接口先行，确保模块独立
- ✅ 频繁集成，快速迭代
- ✅ 任务持久化，支持中断恢复

### 快速检查清单

**启动并行任务前**：
1. ✅ 读取 `.parallel-tasks.json` 检查未完成任务
2. ✅ 读取 `.progress.json` 检查项目阶段
3. ✅ 定义清晰的接口（如果并行开发不同模块）
4. ✅ 启动并行任务（建议 3-5 个同时并行）
5. ✅ 每启动/完成一个任务就保存进度

**恢复中断的任务**：
1. ✅ 读取 `.parallel-tasks.json`
2. ✅ 检查任务状态
3. ✅ 重新启动 `pending` 和 `in_progress` 的任务
4. ✅ 验证 `completed` 任务的结果

### Agent 使用

**可用 Agent**：
- `general-purpose` ⭐ **最常用**：通过 prompt 指定角色
- `go-developer`：Go 代码实现
- `test-engineer`：测试用例编写
- `code-reviewer`：代码审查
- `blockchain-expert`：区块链相关
- `devops-engineer`：Docker、数据库、部署（使用 general-purpose 模拟）

**使用示例**：
```python
# 启动并行任务
Task(
    subagent_type="general-purpose",
    prompt="你是 DevOps 工程师，配置 Docker 环境...",
    run_in_background=True
)
```

**📖 详细文档**：
- `docs/development/PARALLEL_DEVELOPMENT.md` - 并行开发框架
- `docs/development/TASK_RECOVERY.md` - 任务恢复机制
- `CLAUDE.md` 第 1724-2338 行 - 完整的并行开发指南

---

## 5. 开发流程

### 新功能开发
1. 阅读相关文档（需求、设计）
2. 创建功能分支
   ```bash
   git checkout -b feature/price-monitor
   ```
3. 编写代码和测试
4. 运行测试确保通过
5. 提交代码
   ```bash
   git add .
   git commit -m "feat(price): 实现价格监控功能"
   ```
6. 推送到远程
   ```bash
   git push origin feature/price-monitor
   ```

### Bug 修复
1. 定位问题
2. 编写复现用例
3. 修复 Bug
4. 添加测试防止回归
5. 提交修复

### 代码审查清单

提交代码前检查：
- [ ] 代码已通过 `gofmt` 格式化
- [ ] 所有公开 API 有清晰的中文注释
- [ ] 核心逻辑有对应的单元测试
- [ ] 测试覆盖率符合要求
- [ ] 没有硬编码的配置值
- [ ] 错误处理完善，不忽略错误
- [ ] 日志记录合理，使用结构化日志
- [ ] 没有明显的性能问题
- [ ] 敏感信息不暴露
- [ ] Git 提交信息符合规范

---

## 6. 常用命令

### 开发命令
```bash
# 格式化代码
go fmt ./...
goimports -w .

# 运行测试
go test -v ./...
go test -cover ./...

# 生成依赖
go mod tidy
go mod vendor
```

### 构建和运行
```bash
# 构建
go build -o bin/arbitragex cmd/arbitragex/main.go

# 运行
./bin/arbitragex -config config/config.yaml

# 使用 make
make build
make run
make test
```

### Docker 命令
```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose stop

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f price-monitor

# 进入 MySQL 容器
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex
```

### goctl 命令
```bash
# 生成 API 服务代码
goctl api go -api api/arbitragex.api -dir .

# 生成 Model 代码
goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="*" -dir="./model"
```

---

## 7. 项目特定规范

### 搬砖业务相关

1. **价格处理**
   - 所有价格使用 `float64` 存储
   - 金额计算使用整数（USDT 精确到分）
   ```go
   // ✓ 正确
   amountUsdt := int64(100.50 * 100)  // 10050 分
   // ✗ 错误
   amountUsdt := 100.50
   ```

2. **交易对格式**
   - 统一使用 `BTC/USDT` 格式（斜杠分隔）
   - 内部转换各交易所格式

3. **时间处理**
   - 统一使用毫秒时间戳
   - 使用 UTC 时区

4. **错误处理**
   - 所有关键操作必须处理错误
   - 交易相关错误需要记录详细日志

### 安全相关

1. **敏感信息**
   - API 密钥必须加密存储
   - 日志中脱敏显示
   ```go
   // ✓ 正确
   logger.Info("API key", log.String("key", maskAPIKey(key)))
   // ✗ 错误
   logger.Info("API key", log.String("key", key))
   ```

2. **资金安全**
   - 严格遵循风险控制规则
   - 余额不足时不执行交易
   - 大额交易需要分批

### 性能指标
- 价格更新延迟 ≤ 100ms
- 套利识别延迟 ≤ 50ms
- 订单下单延迟 ≤ 100ms
- CPU 使用率 ≤ 70%
- 内存使用 ≤ 2GB

---

## 8. Git 提交规范

### Commit Message 格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具链

### 示例
```
feat(price): 实现价格监控模块

- 添加价格监控器
- 实现多交易所价格获取
- 添加价格缓存机制

Closes #123
```

---

## 9. 参考文档索引

### 技术栈文档
- **后端技术栈**: `docs/design/TechStack/Backend_TechStack.md` (764 行)
- **数据库技术栈**: `docs/design/TechStack/Database_TechStack.md` (411 行)
- **区块链技术栈**: `docs/design/TechStack/Blockchain_TechStack.md` (425 行)

### 设计文档
- **系统架构**: `docs/design/Architecture/System_Architecture.md`
- **模块结构**: `docs/design/Architecture/Module_Structure.md`
- **数据库设计**: `docs/design/Database/Schema_Design.md`
- **数据访问层**: `docs/design/Database/Data_Access_Layer.md`

### 部署文档
- **Docker 部署**: `docs/design/Deployment/Docker_Deployment.md` (700 行)
- **生产环境部署**: `docs/design/Deployment/Production_Deployment.md` (750 行)
- **监控指标**: `docs/design/Monitoring/Metrics_Design.md` (600 行)
- **告警策略**: `docs/design/Monitoring/Alerting_Strategy.md` (550 行)

### 开发文档
- **并行开发框架**: `docs/development/PARALLEL_DEVELOPMENT.md`
- **任务恢复机制**: `docs/development/TASK_RECOVERY.md`
- **代码规范**: `docs/development/CODING_STANDARDS.md` (待创建)
- **配置文件设计**: `docs/config/config_design.md` (v1.1, 已更新 MySQL + go-zero)

### 外部资源
- [go-zero 官方文档](https://go-zero.dev/en/docs/concepts/overview)
- [go-zero GitHub](https://github.com/zeromicro/go-zero)
- [go-zero-looklook 最佳实践](https://github.com/Mikaelemmmm/go-zero-looklook)
- [MySQL 8.0 官方文档](https://dev.mysql.com/doc/refman/8.0/en/)
- [Docker 官方文档](https://docs.docker.com/)

---

## 10. 联系方式

如有问题或建议，请：
1. 查阅项目文档
2. 提交 Issue
3. 在代码 Review 时讨论

---

**文档版本**: v2.1.0 (精简版)
**完整版**: `CLAUDE_FULL.md` (2372 行，包含详细教程)
**最后更新**: 2026-01-08
**维护人**: yangyangyang

---

## 附录：快速参考

### go-zero 快速参考

**项目初始化**：
```bash
# 创建 API 服务
goctl api init -o api/arbitragex.api

# 生成代码
goctl api go -api api/arbitragex.api -dir .

# 生成 Model
goctl model mysql datasource -url="user:password@tcp(localhost:3306)/arbitragex" -table="*" -dir="./model"
```

**配置结构**：
```go
type Config struct {
    rest.RestConf
    Mysql struct {
        DataSource string
    }
    Redis struct {
        Host string
        Type int
    }
}
```

### Docker 快速参考

**启动服务**：
```bash
# 启动所有服务
docker-compose up -d

# 重启单个服务
docker-compose restart price-monitor

# 查看日志
docker-compose logs -f price-monitor
```

**数据库操作**：
```bash
# 连接 MySQL
docker exec -it arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex

# 备份数据库
docker exec arbitragex-mysql mysqldump -uarbitragex_user -pArbitrageX2025! arbitragex > backup.sql

# 恢复数据库
docker exec -i arbitragex-mysql mysql -uarbitragex_user -pArbitrageX2025! arbitragex < backup.sql
```

---

## 💡 经验积累与最佳实践

### 目录结构重构经验（2026-01-08）

#### 问题 1：目录结构不符合 go-zero 规范

**发现的问题**：
- ❌ 原结构：所有服务在 `cmd/` 目录下
- ❌ 不符合 go-zero 官方推荐的工程维度结构
- ❌ 与 go-zero-looklook 最佳实践不一致

**正确的结构**：
```
ArbitrageX/
├── restful/          # HTTP 服务（go-zero 官方推荐）
│   ├── price/
│   ├── engine/
│   └── trade/
├── service/          # RPC 服务（未来扩展）
├── job/              # 定时任务（未来扩展）
└── consumer/         # 消息消费（未来扩展）
```

**重构步骤**：
1. 创建 `restful/` 目录
2. 移动服务：`mv cmd/price restful/`（等）
3. 更新 Makefile 所有路径
4. 更新验证脚本所有路径
5. 批量更新 import 路径
6. 删除重复文件（goctl 生成的额外 main.go）
7. 验证编译

**参考**：
- [go-zero 官方项目结构](https://go-zero.dev/docs/concepts/layout)
- [go-zero-looklook 项目](https://github.com/Mikaelemmmm/go-zero-looklook)

---

### Go 依赖管理经验（2026-01-08）

#### 问题 2：genproto 版本冲突

**错误信息**：
```
go: google.golang.org/genproto@v0.0.0-20221024183307-1bcac889d1e3: invalid version: unknown revision 1bcac889d1e3
```

**根本原因**：
- `go.mod` 中引用了已删除的 genproto 版本
- 该版本的 commit hash 已从仓库中删除

**解决方案**：
```bash
# 1. 从 go.mod 中手动删除无效版本
# 编辑 go.mod，删除这一行：
# google.golang.org/genproto v0.0.0-20221024183307-1bcac889d1e3

# 2. 运行 go mod tidy 自动选择新版本
go mod tidy

# 成功后会自动选择有效版本：
# google.golang.org/genproto/googleapis/api v0.0.0-20240711142825-46eb208f015d
# google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094
```

**预防措施**：
- ✅ 定期运行 `go mod tidy` 保持依赖更新
- ✅ 使用 `go get -u` 更新依赖到最新稳定版本
- ✅ 不要在 `go.mod` 中手动指定伪版本（pseudo-version）

---

### 批量更新代码经验（2026-01-08）

#### 问题 3：重构后 import 路径全部错误

**问题规模**：
- 3 个服务 × 多个文件 = 50+ 个文件需要更新
- 手动更新耗时且容易出错

**解决方案 - 使用 sed 批量替换**：
```bash
# 批量替换 import 路径
find restful/ -name "*.go" -type f -exec sed -i '' 's|arbitragex/cmd/price/|arbitragex/restful/price/|g' {} \;
find restful/ -name "*.go" -type f -exec sed -i '' 's|arbitragex/cmd/engine/|arbitragex/restful/engine/|g' {} \;
find restful/ -name "*.go" -type f -exec sed -i '' 's|arbitragex/cmd/trade/|arbitragex/restful/trade/|g' {} \;
```

**注意事项**：
- ⚠️ macOS 使用 `sed -i ''`，Linux 使用 `sed -i`
- ⚠️ 使用 `|` 作为分隔符而非 `/`，避免路径中的 `/` 冲突
- ✅ 先用 `grep` 查找确认，再批量替换
- ✅ 替换后立即编译验证

**其他批量操作工具**：
- `find + xargs`：更灵活的批量操作
- `grep -r + sed`：基于搜索结果批量替换
- IDE 全局替换：适合有图形界面的情况

---

#### 问题 4：goctl 生成重复的 main 函数

**问题现象**：
```
restful/price/price.go:18:5: configFile redeclared in this block
restful/price/main.go:15:5: other declaration of configFile
```

**原因**：
- goctl 生成了 `main.go` 和 `price.go`
- 两个文件都在 `package main` 中
- 都定义了 `main()` 函数和 `configFile` 变量

**解决方案**：
```bash
# 删除 goctl 生成的额外文件
rm -f restful/price/price.go
rm -f restful/engine/engine.go
rm -f restful/trade/trade.go
```

**预防措施**：
- ✅ 使用 `goctl api go` 时，检查是否生成了重复文件
- ✅ 如果已存在 `main.go`，goctl 可能会生成额外的 main 文件
- ✅ 保留自己编写的 `main.go`，删除 goctl 生成的

---

### 经验积累机制（自动学习）

#### 建立持续改进机制

**核心思想**：
> 每次解决问题后，总结经验并写入文档，建立知识库，防止重复犯错。

**实施步骤**：

**1. 问题发现时**
- ✅ 记录问题的现象和错误信息
- ✅ 分析问题的根本原因
- ✅ 尝试多种解决方案

**2. 问题解决后**
- ✅ 记录最终有效的解决方案
- ✅ 总结经验教训（DO/DON'T）
- ✅ 提取可复用的通用经验

**3. 经验文档化**
- ✅ 将经验写入 CLAUDE.md 或项目文档
- ✅ 使用清晰的标题和分类
- ✅ 包含代码示例和命令
- ✅ 添加参考链接

**4. 定期回顾**
- ✅ 每个阶段完成后回顾经验
- ✅ 项目复盘时更新最佳实践
- ✅ 分享经验给团队成员

**经验文档模板**：
```markdown
### [问题名称]（日期）

#### 问题描述
- **现象**：[具体的错误或问题]
- **影响**：[对项目的影响]
- **发现时机**：[何时发现此问题]

#### 根本原因
[分析为什么会发生此问题]

#### 解决方案
```bash
# 具体解决步骤和命令
[命令示例]
```

#### 经验教训
**DO**（应该做的）：
- ✅ [正确的做法]

**DON'T**（不应该做的）：
- ❌ [避免的做法]

#### 预防措施
[如何防止类似问题再次发生]

#### 参考资源
- [相关文档链接]
- [GitHub Issue 或讨论]
```

**自动化建议**：
- 考虑使用 Git hooks 在提交时检查常见问题
- 创建脚本自动检测代码质量问题
- 使用 linter 和 formatter 强制代码规范

---

### 关键经验总结

#### DO（应该做的）

1. **遵循框架最佳实践**
   - ✅ 使用 goctl 工具生成代码结构
   - ✅ 采用官方推荐的目录结构（`restful/`）
   - ✅ API 定义先行（.api 文件）

2. **依赖管理**
   - ✅ 定期运行 `go mod tidy` 更新依赖
   - ✅ 及时删除无效的依赖版本
   - ✅ 使用 Go 1.21+ 和 go-zero v1.9.4+

3. **代码重构**
   - ✅ 使用批量工具（sed、find）提高效率
   - ✅ 重构后立即编译验证
   - ✅ 删除重复和过时的文件

4. **经验积累**
   - ✅ 解决问题后立即总结经验
   - ✅ 将通用经验写入文档
   - ✅ 定期回顾和更新最佳实践

#### DON'T（不应该做的）

1. **违反框架规范**
   - ❌ 不要手动编写 go-zero 标准结构代码
   - ❌ 不要使用不符合规范的目录结构
   - ❌ 不要跳过阶段验证

2. **依赖管理错误**
   - ❌ 不要手动指定伪版本（pseudo-version）
   - ❌ 不要忽略依赖版本冲突
   - ❌ 不要使用过时的依赖版本

3. **低效操作**
   - ❌ 不要手动逐个文件更新 import 路径
   - ❌ 不要重复造轮子
   - ❌ 不要在重构后忘记验证

4. **忽视经验积累**
   - ❌ 不要重复犯错
   - ❌ 不要忽视文档更新
   - ❌ 不要跳过问题总结

---

### 工具和脚本

**批量操作脚本示例**：
```bash
# 批量更新 import 路径
update_imports() {
    local old_path=$1
    local new_path=$2
    find . -name "*.go" -type f -exec sed -i '' "s|$old_path|$new_path|g" {} \;
}

# 使用示例
update_imports "arbitragex/cmd/" "arbitragex/restful/"

# 批量删除文件
cleanup_duplicate_files() {
    rm -f restful/*/price.go
    rm -f restful/*/engine.go
    rm -f restful/*/trade.go
}
```

**验证脚本**：
- `scripts/verify-stage.sh` - 9 步自动验证
- `Makefile` - `make verify-stage` 快速验证

---

**版本**: v2.3.0
**最后更新**: 2026-01-08
**维护人**: yangyangyang
