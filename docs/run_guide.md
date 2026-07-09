# 本地运行指南

## 1. 初始化数据库

```powershell
cd F:\study\database\project1
Get-Content .\database\schema.sql | mysql -u root -p
```

## 2. 启动后端

```powershell
cd F:\study\database\project1\backend-go
$env:MYSQL_DSN="root:your_mysql_password@tcp(127.0.0.1:3306)/campus_takeout?parseTime=true&charset=utf8mb4&loc=Local"
$env:DB_MAX_OPEN_CONNS="50"
$env:DB_MAX_IDLE_CONNS="10"
$env:DB_CONN_MAX_LIFETIME_MINUTES="30"
$env:DB_CONN_MAX_IDLE_TIME_MINUTES="10"
$env:DB_QUERY_TIMEOUT_SECONDS="3"
$env:LOGIN_RATE_LIMIT_PER_MIN="30"
$env:ORDER_RATE_LIMIT_PER_MIN="20"
$env:UPLOAD_RATE_LIMIT_PER_MIN="20"
$env:ROUTE_CACHE_TTL_SECONDS="30"
$env:RECOMMEND_CACHE_TTL_SECONDS="60"
go run .
```

后端默认地址：http://localhost:8080

### 2.1 连接池参数说明

- `DB_MAX_OPEN_CONNS`：最大打开连接数，默认 `50`
- `DB_MAX_IDLE_CONNS`：最大空闲连接数，默认 `10`
- `DB_CONN_MAX_LIFETIME_MINUTES`：单连接最大生命周期，默认 `30` 分钟
- `DB_CONN_MAX_IDLE_TIME_MINUTES`：空闲连接最大保留时间，默认 `10` 分钟
- `DB_QUERY_TIMEOUT_SECONDS`：主要 SQL 操作超时时间，默认 `3` 秒

建议先使用默认值，本机压测后再微调。

### 2.2 限流说明

- 登录接口：`/api/auth/login`
- 下单接口：`/api/orders`
- 图片上传接口：`/api/merchant/upload/dish-image`

以上接口按 IP 做分钟级限流，超过阈值返回 `429 Too Many Requests`。

### 2.3 缓存说明

当前使用短时内存缓存的接口：

- `/api/dining-areas`
- `/api/dining-areas/:id/canteens`
- `/api/canteens/:id/stores`
- `/api/ai/recommend/dishes`

缓存时间默认 `30` 秒或 `60` 秒，可通过 `ROUTE_CACHE_TTL_SECONDS`、`RECOMMEND_CACHE_TTL_SECONDS` 调整。
购物车、下单、支付、评价等强一致接口不会缓存。

## 3. 启动前端

```powershell
cd F:\study\database\project1\frontend
npm install
npm run dev
```

前端默认地址：http://localhost:5173

## 4. 编译 C++ 工具

```powershell
cd F:\study\database\project1\cpp-tools
g++ -std=c++17 data_generator.cpp -o data_generator.exe
g++ -std=c++17 sales_analyzer.cpp -o sales_analyzer.exe
```

## 4.1 图片上传测试

1. 使用 `merchant01 / 123456` 登录商家端。
2. 在新增菜品区域选择一张 `jpg/png/webp` 图片。
3. 点击“上传图片”，成功后会回填 `image_url` 并显示预览。
4. 上传文件保存在 `backend-go/uploads/dishes/`。
5. 如果图片无法显示，先检查后端 `main.go` 是否已注册 `/uploads` 静态路由。

## 4.2 评价功能测试

1. 使用学生账号下单并支付。
2. 使用商家账号把订单流转到 `COMPLETED`。
3. 学生重新查看店铺详情页或订单列表。
4. 只有已完成且未评价订单会显示“评价”按钮。
5. 提交成功后会变成“已评价”，并刷新评论区。

## 5. AI 可选配置

如果希望使用 DeepSeek 生成菜品文案或图片提示词，可以在启动后端前配置环境变量：

```powershell
$env:DEEPSEEK_API_KEY="你的 API Key"
$env:DEEPSEEK_BASE_URL="https://api.deepseek.com"
$env:DEEPSEEK_MODEL="deepseek-chat"
```

说明：

- `DEEPSEEK_API_KEY` 为空时，后端不会报 500，而是自动使用本地 fallback 文案。
- `DEEPSEEK_BASE_URL` 和 `DEEPSEEK_MODEL` 都是可选项，不配置时会使用默认值。

## 6. 管理员批量导入用户

管理员登录后，在管理员主页上传 CSV 文件即可批量导入用户。CSV 表头必须严格为：

```text
username,password,email,phone,role,real_name,student_no,college,class_name
```

说明：

- `role` 只允许 `STUDENT` 或 `TEACHER`
- `password` 为空时默认使用 `123456`
- `real_name`、`student_no`、`college`、`class_name` 会写入 `student_profiles`
- 系统会逐行校验重复用户名、邮箱、手机号和学号，并返回成功/失败明细

## 7. 常见错误说明

### 6.1 8080 端口被占用

现象：后端启动时报 listen tcp :8080 bind failed。

处理：关闭占用进程，或设置 ADDR 为其他端口后再启动。

### 6.2 MYSQL_DSN 为空

现象：日志提示 MYSQL_DSN is empty，后端仅能跑健康检查。

处理：先设置环境变量再 go run .

### 6.3 MySQL Access denied

现象：连接数据库时报 Access denied for user。

处理：检查用户名、密码、主机和授权。

### 6.4 No database selected

现象：执行 SQL 脚本时报 No database selected。

处理：先在 MySQL 中创建并使用目标库，或在连接参数里指定数据库名。

### 6.5 PowerShell 不支持 < 重定向

现象：mysql < schema.sql 在 PowerShell 直接失败。

处理：改用 Get-Content .\database\schema.sql | mysql -u root -p。

### 6.6 missing bearer token

现象：访问受保护接口返回 missing bearer token。

处理：登录后在请求头携带 Authorization: Bearer <token>。

## 8. 压测方式

### 8.1 简单压测健康检查

```powershell
hey -n 1000 -c 100 http://localhost:8080/health
```

### 8.2 压测读接口

```powershell
hey -n 500 -c 50 -m GET -H "Authorization: Bearer <token>" http://localhost:8080/api/dining-areas
hey -n 500 -c 50 -m GET -H "Authorization: Bearer <token>" http://localhost:8080/api/ai/recommend/dishes
```

### 8.3 压测写接口注意事项

- 登录接口会被限流，不适合无限制高并发重复压。
- 下单接口依赖购物车和库存，压测前要准备好测试数据。
- 图片上传接口也有限流，压测时建议使用小文件并控制并发。

## 9. 部署到服务器的建议

- 使用环境变量而不是写死数据库地址和 JWT 密钥。
- 使用 Nginx 反向代理 `:8080`，并代理 `/uploads/` 静态资源。
- 给 MySQL 单独开放内网访问，不要直接暴露公网。
- 上线前根据机器配置调整连接池大小和超时参数。
- 如果后续要承受更高并发，可把缓存与限流迁移到 Redis / API 网关。
