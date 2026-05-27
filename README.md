# 校园外卖管理系统

## 1. 项目简介

本项目是数据库课程设计示例，目标是实现一个可演示的校园外卖管理系统，覆盖用户下单、商家履约、管理员监管的完整业务闭环。

## 2. 技术栈

- 后端：Go、Gin、database/sql、MySQL
- 前端：Vue3、Vite、Axios
- 辅助工具：C++17

约束：不使用 Docker，不使用 GORM。

## 3. 目录结构

```text
project1/
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
- AI 功能：菜品推荐排序、菜品文案生成、图片提示词生成
- 食堂体系：旧饭 / 新饭 / 多饭 与 6 个食堂区域结构
- 前端演示页：登录页、用户页、商家页、管理员页
- C++ 工具：测试数据 SQL 生成、CSV 销售统计

## 5. 本地运行步骤

### 5.1 初始化数据库

```powershell
cd F:\study\database\project1
Get-Content .\database\schema.sql | mysql -u root -p
```

### 5.2 启动后端

```powershell
cd F:\study\database\project1\backend-go
$env:MYSQL_DSN="root:200411@tcp(127.0.0.1:3306)/campus_takeout?parseTime=true&charset=utf8mb4&loc=Local"
go run .
```

### 5.3 启动前端

```powershell
cd F:\study\database\project1\frontend
npm install
npm run dev
```

### 5.4 编译 C++ 工具

```powershell
cd F:\study\database\project1\cpp-tools
g++ -std=c++17 data_generator.cpp -o data_generator.exe
g++ -std=c++17 sales_analyzer.cpp -o sales_analyzer.exe
```

## 6. 测试账号

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

- 增加前端订单详情页与更细粒度状态提示
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
- 商家页：店铺与菜品管理、订单流转、AI 文案生成、订单详情弹层
- 管理员页：全局订单、订单详情、统计卡片、用户管理、批量导入

## 14. 常见问题补充

- 订单详情没有明细：点击“查看详情”会自动请求详情接口补全 items
- 订单状态按钮不可点：仅在满足状态流转规则时显示下一步按钮
- 评价失败提示已评价：后端已限制一单一评，属于预期行为
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
