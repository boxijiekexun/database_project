# 校园外卖管理系统 - Go 后端

当前后端继续使用 Gin + `database/sql` + MySQL，不引入 Docker 和 GORM。

## 目录

- [main.go](main.go)
- [go.mod](go.mod)
- [database schema](../database/schema.sql)

## 初始化 MySQL

1. 创建数据库，例如 `project1`
2. 在 PowerShell 中执行：`Get-Content ..\database\schema.sql | mysql -u root -p`
3. 确认种子数据已写入：
   - `admin / 123456`
   - `student01 / 123456`
   - `teacher01 / 123456`
   - `merchant01 / 123456`

说明：`database/schema.sql` 会重建当前阶段需要的 `users`、`merchants`、`stores`、`dish_categories`、`dishes` 表，并写入测试数据。验证码仍然是内存数字验证码，`GET /api/auth/captcha` 会直接返回 `captcha_id` 和 `captcha_code`。

## 进入后端目录

```bash
cd backend-go
```

## 启动 Go 后端

```bash
cd backend-go
go mod tidy
go run .
```

## 运行测试

```bash
go test ./...
```

## 环境变量

- `ADDR` - 服务监听地址，默认 `:8080`
- `MYSQL_DSN` - MySQL DSN，例如 `root:password@tcp(127.0.0.1:3306)/project1?parseTime=true`
- `JWT_SECRET` - JWT 签名密钥，未设置时使用开发默认值

## 返回格式

接口统一使用以下 JSON 结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

失败时返回：

```json
{
  "code": 400,
  "message": "错误原因"
}
```

## 接口

### 公共接口

- `GET /health`
- `GET /api/auth/captcha`
- `POST /api/auth/login`
- `POST /api/auth/register`
- `GET /api/stores`
- `GET /api/stores/:id`
- `GET /api/stores/:id/categories`
- `GET /api/stores/:id/dishes`

### 角色接口

- `GET /api/student/dashboard`
- `GET /api/teacher/dashboard`
- `GET /api/merchant/dashboard`
- `GET /api/admin/dashboard`

### 商家接口

- `GET /api/merchant/store`
- `PUT /api/merchant/store`
- `GET /api/merchant/categories`
- `POST /api/merchant/categories`
- `PUT /api/merchant/categories/:id`
- `DELETE /api/merchant/categories/:id`
- `GET /api/merchant/dishes`
- `POST /api/merchant/dishes`
- `PUT /api/merchant/dishes/:id`
- `PATCH /api/merchant/dishes/:id/status`

### 购物车接口

- `GET /api/cart/items`
- `POST /api/cart/items`
- `PUT /api/cart/items/:id`
- `DELETE /api/cart/items/:id`
- `DELETE /api/cart/items`

### 学生/教师订单接口

- `POST /api/orders`
- `GET /api/orders`
- `GET /api/orders/:id`
- `POST /api/orders/:id/cancel`
- `POST /api/orders/:id/pay`
- `POST /api/orders/:id/review`

### 商家订单接口

- `GET /api/merchant/orders`
- `GET /api/merchant/orders/:id`
- `PATCH /api/merchant/orders/:id/status`

### 管理员订单接口

- `GET /api/admin/orders`
- `GET /api/admin/orders/:id`

## 手工测试流程

1. `student01` 登录获取 token。
2. 查询店铺列表。
3. 查询某个店铺的菜品列表。
4. 加入购物车。
5. 查看购物车。
6. 创建订单。
7. 模拟支付。
8. `merchant01` 登录获取 token。
9. 商家查看订单。
10. 商家依次修改订单状态。
11. `student01` 查看订单详情。
12. `student01` 对已完成订单评价。