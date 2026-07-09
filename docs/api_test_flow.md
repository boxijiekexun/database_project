# API 联调流程

本文档用于课程设计演示，基于当前项目的真实 MySQL + Go 后端。

## 1. 启动 MySQL

确保本机 MySQL 服务已启动，并且可以使用 root 账号连接。

## 2. 导入数据库脚本

在 PowerShell 执行：

```powershell
cd F:\study\database\project1
Get-Content .\database\schema.sql | mysql -u root -p
```

## 3. 设置后端环境变量并启动

```powershell
cd F:\study\database\project1\backend-go
$env:MYSQL_DSN="root:your_mysql_password@tcp(127.0.0.1:3306)/campus_takeout?parseTime=true&charset=utf8mb4&loc=Local"
go run .
```

## 4. student01 登录

```powershell
$loginBody = @{ username = "student01"; password = "123456" } | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body $loginBody
$token = $loginResp.data.token
$headers = @{ Authorization = "Bearer $token" }
```

## 5. 查询店铺

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/stores" -Method GET -Headers $headers
```

## 6. 查询店铺菜品

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/stores/1/dishes" -Method GET -Headers $headers
```

## 7. 加入购物车

```powershell
$cartBody = @{ dish_id = 1; quantity = 2 } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/cart/items" -Method POST -Headers $headers -ContentType "application/json" -Body $cartBody
```

## 8. 创建订单

```powershell
$orderBody = @{ remark = "课程设计联调订单" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/orders" -Method POST -Headers $headers -ContentType "application/json" -Body $orderBody
```

## 9. 支付订单

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/orders/1/pay" -Method POST -Headers $headers
```

## 10. merchant01 登录并处理订单

```powershell
$merchantLogin = @{ username = "merchant01"; password = "123456" } | ConvertTo-Json
$merchantResp = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -ContentType "application/json" -Body $merchantLogin
$merchantToken = $merchantResp.data.token
$merchantHeaders = @{ Authorization = "Bearer $merchantToken" }

Invoke-RestMethod -Uri "http://localhost:8080/api/merchant/orders" -Method GET -Headers $merchantHeaders

$next1 = @{ status = "ACCEPTED" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/merchant/orders/1/status" -Method PATCH -Headers $merchantHeaders -ContentType "application/json" -Body $next1

$next2 = @{ status = "PREPARING" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/merchant/orders/1/status" -Method PATCH -Headers $merchantHeaders -ContentType "application/json" -Body $next2

$next3 = @{ status = "DELIVERING" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/merchant/orders/1/status" -Method PATCH -Headers $merchantHeaders -ContentType "application/json" -Body $next3

$next4 = @{ status = "COMPLETED" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/merchant/orders/1/status" -Method PATCH -Headers $merchantHeaders -ContentType "application/json" -Body $next4
```

## 11. student01 评价订单

```powershell
$reviewBody = @{ rating = 5; content = "口味很好，配送很快" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/orders/1/review" -Method POST -Headers $headers -ContentType "application/json" -Body $reviewBody
```

完成后即可演示完整链路：下单 -> 支付 -> 商家接单流转 -> 完成后评价。
