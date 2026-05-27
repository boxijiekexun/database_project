# 课程设计答辩演示脚本

## 1. 项目简介

本项目实现了一个校园外卖管理系统，覆盖用户下单、商家履约、管理员监管的完整业务闭环。系统目标是提供可运行、可演示、可扩展的课程设计样例。

## 2. 技术栈说明

1. 后端：Go + Gin + database/sql + MySQL
2. 前端：Vue3 + Vite + Axios
3. 辅助工具：C++17

约束遵循：不引入 Docker，不使用 GORM。

## 3. 数据库设计亮点

1. 核心表结构完整：users、dining_areas、canteens、stores、dishes、cart_items、orders、order_items、reviews
2. 外键与约束明确，保证数据一致性
3. 订单明细保存菜品快照（dish_name、dish_price），避免历史订单受菜品变更影响
4. 订单金额由后端基于数据库价格计算，防止前端篡改
5. AI 推荐基于订单统计数据，兼顾销量、数量、金额和评分，成本低且可解释

## 4. 后端设计亮点

1. 角色权限分离：STUDENT、TEACHER、MERCHANT、ADMIN
2. JWT 鉴权和中间件控制接口访问
3. 下单与状态流转采用事务控制，避免并发一致性问题
4. 订单状态按业务规则流转：
   PENDING_PAYMENT -> PAID -> ACCEPTED -> PREPARING -> DELIVERING -> COMPLETED

## 5. C++ 工具说明

1. data_generator.cpp
   - 生成 generated_data.sql 测试数据脚本
   - 用于快速扩充用户、商家、店铺、分类、菜品数据

2. sales_analyzer.cpp
   - 读取 orders_export.csv
   - 统计每店销售额、每菜品销量、总销售额

## 6. 演示流程

1. 启动 MySQL
2. 导入数据库脚本
   - Get-Content .\database\schema.sql | mysql -u root -p
3. 启动 Go 后端
   - 进入 backend-go
   - 设置 MYSQL_DSN
   - go run .
4. 启动 Vue 前端
   - 进入 frontend
   - npm run dev
5. student01 登录
6. 浏览店铺和菜品
7. 加入购物车
8. 创建订单
9. 支付订单
10. merchant01 登录
11. 商家执行接单、制作、配送、完成
12. student01 登录查看订单并评价
13. admin 登录查看全局订单和统计卡片
14. admin 在用户管理区修改用户状态、调整角色、重置密码
15. student01 登录查看旧饭 / 新饭 / 多饭，并进入 6 个食堂体系浏览
16. 查看 AI 推荐菜品并加入购物车
17. merchant01 使用 AI 生成菜品文案和图片提示词
18. 展示 C++ data_generator 和 sales_analyzer

## 7. 讲解建议

1. 先展示业务闭环，再解释分层设计
2. 对照订单详情弹层和状态时间线，说明流程可视化
3. 用管理员统计卡片说明运营视角
4. 重点说明 AI 接口并非依赖昂贵外部调用，而是先用数据库统计实现低成本推荐
5. 最后补充 C++ 工具体现工程化与数据处理能力
