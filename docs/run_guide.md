# 本地运行指南

## 1. 初始化数据库

```powershell
cd F:\study\database\project1
Get-Content .\database\schema.sql | mysql -u root -p
```

## 2. 启动后端

```powershell
cd F:\study\database\project1\backend-go
$env:MYSQL_DSN="root:200411@tcp(127.0.0.1:3306)/campus_takeout?parseTime=true&charset=utf8mb4&loc=Local"
go run .
```

后端默认地址：http://localhost:8080

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
