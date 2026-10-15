# 校园外卖管理系统

## 1. 项目简介

这是一个数据库课程设计项目：从学生选餐下单、商家接单与配送，到管理员查看订单和管理用户，串起一条可本地演示的业务流程。支付为**模拟支付**；项目尚无公开在线演示，按照下方步骤可在本地运行。

### 值得关注的实现

- **下单一致性**：在事务中创建订单，金额按数据库中的菜品价格计算；库存扣减使用条件更新，避免并发超卖。
- **订单状态**：按 `PAID → ACCEPTED → PREPARING → DELIVERING → COMPLETED` 流转，条件更新避免重复操作。
- **后端稳定性**：数据库连接池、SQL 超时、接口限流、短时内存缓存及订单分页。
- **可追溯的设计**：[数据库设计](docs/database_design.md) · [事务设计](docs/transaction_design.md) · [接口联调](docs/api_test_flow.md) · [本地运行指南](docs/run_guide.md)。

> 项目定位是课程设计与工程练习，不是生产环境外卖服务。推荐菜品采用基于订单等数据的规则排序；菜品文案可选接入 DeepSeek，未配置 Key 时使用本地 fallback；图片功能生成提示词，不直接生成图片。

## 2. 技术栈

- 后端：Go、Gin、database/sql、MySQL
- 前端：Vue3、Vite、Axios
- 辅助工具：C++17

约束：不使用 Docker，不使用 GORM。

## 3. 目录结构

```text
database_project/
├── backend-go/     Go + Gin 后端
├── frontend/       Vue3 + Vite 前端
├── database/       MySQL SQL 文件
├── cpp-tools/      C++17 辅助工具
└── docs/           文档
```

## 4. 已完成功能

- 认证：登录、注册、JWT 鉴权、角色权限中间件
- 商家端：店铺管理、菜品分类管理、菜品管理、菜品上下架
- 用户端：店铺与菜品浏览、购物车、创建订单、模拟支付、订单评价
- 订单流转：PAID -> ACCEPTED -> PREPARING -> DELIVERING -> COMPLETED
- 管理员端：全局订单列表与详情、统计卡片、批量导入用户
- 管理员端：用户列表、角色调整、状态启停、重置密码
- AI 相关：规则推荐排序、可选模型生成菜品文案、图片提示词生成（不直接生成图片）
- 食堂体系：旧饭 / 新饭 / 多饭 与 6 个食堂区域结构
- 前端演示页：登录页、用户页、商家页、管理员页
- C++ 工具：测试数据 SQL 生成、CSV 销售统计

## 5. 本地运行步骤

### 5.1 初始化数据库

在克隆后的仓库根目录打开 PowerShell。数据库初始化、后端和前端分别使用终端；后续命令均从仓库根目录开始执行。`schema.sql` 会删除并重建目标库中的表，**只对专用的空白演示库执行；不要对已有数据的数据库运行**。

```powershell
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS campus_takeout DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
Get-Content -Raw .\database\schema.sql | mysql -u root -p campus_takeout
```

### 5.2 启动后端

```powershell
Set-Location .\backend-go
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

### 5.3 启动前端

```powershell
Set-Location .\frontend
npm install
npm run dev
```

### 5.4 编译 C++ 工具

```powershell
Set-Location .\cpp-tools
g++ -std=c++17 data_generator.cpp -o data_generator.exe
g++ -std=c++17 sales_analyzer.cpp -o sales_analyzer.exe
```

## 6. 测试账号

以下仅供本地演示，公开部署前必须删除或重设这些默认账号与密码。

- admin / 123456
- student01 / 123456
- teacher01 / 123456
- merchant01 / 123456

## 7. 后端启动方式

后端地址默认：http://localhost:8080

详细接口和后端说明见 [backend-go/README.md](backend-go/README.md)。

## 8. 前端启动方式

前端地址默认：http://localhost:5173

登录后按角色自动跳转：

- STUDENT/TEACHER -> 用户主页
- MERCHANT -> 商家主页
- ADMIN -> 管理员主页

## 9. C++ 工具编译方式

详见 [cpp-tools/README.md](cpp-tools/README.md)。

## 10. 后续计划

- 增加导出订单 CSV 的后端接口
- 增加基础自动化测试用例
- 增加销售统计可视化页面

## 11. 第五阶段新增功能

- 用户端、商家端、管理员端均支持订单详情查看
- 订单详情新增状态时间线（含取消态文案）
- 商家端订单流转按钮优化为“仅展示下一步操作”
- 用户端订单状态中文化、支付/评价按钮按状态显示
- 管理员端新增前端统计卡片：订单总数、已完成、总销售额、待处理
- 管理员端新增 CSV 批量导入用户功能，支持导入成功/失败明细展示
- 管理员用户管理增强：支持关键词、角色、状态筛选，支持重置密码与详情查看
- 用户端新增食堂区域体系与 AI 推荐菜品区
- 商家端新增 AI 生成菜品文案与图片提示词

## 11.1 本次改动说明

- 文件注释：后端主入口、认证/用户/店铺/订单/AI 模块，前端入口与核心页面，数据库脚本和 C++ 工具都补充了中文文件作用说明。
- 评价修复：订单查询接口增加 `reviewed` 字段，前端只在 `status === "COMPLETED" && !reviewed` 时显示“评价”，评价成功后刷新订单和评论区。
- 店铺详情页：重排为更像外卖平台的三栏布局，包含分类导航、菜品列表、固定购物车和评论区。
- 图片上传：商家端新增菜品图片上传入口，后端开放 `/uploads` 静态目录，学生端可直接看到菜品图片。
- 评论分析：店铺详情页新增点评关键词云和评论摘要，基于 `reviews.rating` 与 `reviews.content` 实时统计生成。
- AI 图片提示词：商家端支持生成中文/英文图片提示词与负面提示词，该功能只生成 prompt，不直接生成真实图片。
- 稳定性优化：新增数据库连接池配置、SQL 超时、登录/下单/上传限流、食堂与推荐接口缓存、订单分页与关键失败日志。

## 11.2 稳定性与并发优化

- 连接池：支持 `DB_MAX_OPEN_CONNS`、`DB_MAX_IDLE_CONNS`、`DB_CONN_MAX_LIFETIME_MINUTES`、`DB_CONN_MAX_IDLE_TIME_MINUTES`。
- SQL 超时：仓储层主要数据库操作统一使用 `DB_QUERY_TIMEOUT_SECONDS` 控制超时，避免慢 SQL 长时间占用连接。
- 下单事务：下单仍在一个事务内完成，金额由后端按数据库价格计算，库存扣减使用 `UPDATE ... WHERE stock >= ?` 防止超卖。
- 状态流转：商家更新订单状态时使用条件更新，避免并发下重复流转。
- 限流：登录、下单、图片上传接口按 IP 做分钟级限流，超出返回 `429`。
- 缓存：`/api/dining-areas`、`/api/dining-areas/:id/canteens`、`/api/canteens/:id/stores`、`/api/ai/recommend/dishes` 使用内存短缓存。
- 分页：订单列表接口支持可选 `page` / `page_size`；不传分页参数时仍保持原返回结构，兼容现有前端。

## 11.3 压测建议

- 推荐先压测只读接口：`/api/dining-areas`、`/api/canteens/:id/stores`、`/api/ai/recommend/dishes`。
- 再压测写接口：登录、下单、商家流转、图片上传。
- 可使用 `ab`、`wrk`、`hey` 或 JMeter，例如：

```powershell
hey -n 500 -c 50 http://localhost:8080/health
hey -n 200 -c 20 -m GET -H "Authorization: Bearer <token>" http://localhost:8080/api/dining-areas
```

- 重点观察：
  - 后端日志中是否出现登录失败、库存不足、状态流转冲突、AI fallback、上传失败等记录
  - MySQL 连接数是否稳定在连接池限制之内
  - 慢查询或锁等待是否明显增加

## 11.4 服务器部署建议

- 生产环境应修改 `JWT_SECRET`、`MYSQL_DSN`，不要使用开发默认值。
- 建议使用 Nginx 反向代理 Go 服务，并把 `/uploads/` 暴露为静态目录。
- 建议把 MySQL 与后端部署在同一内网，降低连接延迟。
- 建议根据机器配置调整连接池，例如 2C/4G 机器可从 `30~50` 的最大连接数开始试。
- 若访问量继续增大，可把当前内存缓存和限流迁移到 Redis / 网关层。

## 12. 演示顺序

1. 启动 MySQL 并导入 schema.sql
2. 启动后端服务（http://localhost:8080）
3. 启动前端服务（http://localhost:5173）
4. student01 演示浏览、下单、支付
5. merchant01 演示接单到完成
6. student01 演示查看详情与评价
7. admin 演示全局订单与统计
8. 演示 C++ 数据生成与销售统计工具

## 13. 前端页面说明

- 登录页：账号登录并按角色跳转
- 用户页：店铺/菜品浏览、购物车、下单、支付、评价、订单详情弹层、食堂区域与 AI 推荐
- 商家页：店铺与菜品管理、菜品图片上传、订单流转、AI 文案生成、订单详情弹层
- 管理员页：全局订单、订单详情、统计卡片、用户管理、批量导入

## 14. 常见问题补充

- 订单详情没有明细：点击“查看详情”会自动请求详情接口补全 items
- 订单状态按钮不可点：仅在满足状态流转规则时显示下一步按钮
- 评价失败提示已评价：后端已限制一单一评，属于预期行为
- 菜品图片不显示：检查后端是否启动在 `http://localhost:8080`，并确认 `backend-go/uploads/dishes` 下已有图片文件
- 上传图片失败：仅支持 `jpg/jpeg/png/webp`，并且文件大小不能超过后端限制
- 前端请求 401：检查 localStorage 中 token 是否过期，重新登录
- 管理员批量导入用户：上传符合表头要求的 CSV 文件，系统会逐行校验并返回成功/失败明细
- AI 接口没有配置 Key：系统会自动使用本地 fallback，不会返回 500

## 15. 课程设计交付材料

以下文档已整理为课程设计报告与答辩可直接使用的材料：

1. 需求分析：[docs/requirements.md](docs/requirements.md)
2. E-R 设计与关系说明：[docs/er_design.md](docs/er_design.md)
3. 关系模式设计：[docs/relation_schema.md](docs/relation_schema.md)
4. 事务设计（下单与取消）：[docs/transaction_design.md](docs/transaction_design.md)
5. 索引与约束设计：[docs/index_and_constraint_design.md](docs/index_and_constraint_design.md)
6. 课程设计报告大纲：[docs/report_outline.md](docs/report_outline.md)
7. 答辩 PPT 大纲：[docs/ppt_outline.md](docs/ppt_outline.md)
8. 演示脚本：[docs/demo_script.md](docs/demo_script.md)

## 参考文档

- 接口联调流程：[docs/api_test_flow.md](docs/api_test_flow.md)
- 数据库设计说明：[docs/database_design.md](docs/database_design.md)
- 运行指南：[docs/run_guide.md](docs/run_guide.md)
- 答辩演示脚本：[docs/demo_script.md](docs/demo_script.md)
