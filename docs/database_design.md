# 数据库设计说明

## 1. 系统角色

系统包含四类角色：

1. ADMIN：系统管理员，查看全局订单、维护基础数据
2. STUDENT：学生用户，可下单、支付、评价
3. TEACHER：教师用户，可下单、支付、评价
4. MERCHANT：商家用户，管理店铺与菜品，处理订单

## 2. 核心实体

核心业务实体包括：用户、学生档案、商家、店铺、分类、菜品、购物车、订单、订单明细、评价。

## 3. 主要数据表

### users

存储账号、密码哈希、角色和状态。

### student_profiles

存储学生扩展信息，通过 user_id 关联 users。批量导入用户时，如果 CSV 中提供了 real_name、student_no、college、class_name，这些信息会同步写入该表。

### dining_areas

存储食堂区域信息，示例分为旧饭、新饭、多饭三大区域。

### canteens

存储具体食堂信息，通过 area_id 关联 dining_areas，并保存简介与特色美食。

### merchants

商家扩展信息，通过 user_id 关联 users。

### stores

店铺主体信息，通过 merchant_id 关联 merchants，通过 canteen_id 关联 canteens。

### dish_categories

店铺下的菜品分类，通过 store_id 关联 stores。

### dishes

菜品信息，包括价格、库存、上下架状态，通过 store_id 和 category_id 关联。

### cart_items

购物车条目，按 user_id + dish_id 唯一，保证同一用户同一菜品只有一条记录。

### orders

订单主表，保存订单号、用户、店铺、总金额、状态与备注。

### order_items

订单明细表，保存每个菜品的购买数量与小计。

### reviews

评价表，一单一评，通过 order_id 唯一约束保证重复评价不可写入。

## 4. 主要外键关系

1. student_profiles.user_id -> users.id
2. merchants.user_id -> users.id
3. dining_areas.id <- canteens.area_id
4. canteens.id <- stores.canteen_id
5. stores.merchant_id -> merchants.id
6. dish_categories.store_id -> stores.id
7. dishes.store_id -> stores.id
8. dishes.category_id -> dish_categories.id
9. cart_items.user_id -> users.id
10. cart_items.store_id -> stores.id
11. cart_items.dish_id -> dishes.id
12. orders.user_id -> users.id
13. orders.store_id -> stores.id
14. order_items.order_id -> orders.id
15. order_items.dish_id -> dishes.id
16. reviews.order_id -> orders.id
17. reviews.user_id -> users.id
18. reviews.store_id -> stores.id

## 5. AI 推荐使用的数据统计

AI 推荐菜品并不依赖外部模型计算排序，而是优先使用数据库聚合统计：

1. order_items 中的订单出现次数
2. order_items 中的购买数量总和
3. order_items 中的销售额总和
4. dishes.sales_count 历史销量字段
5. stores.rating 食堂/店铺评分

排序得分由后端规则组合生成，推荐理由也是基于这些统计项本地拼接而成，因此成本低、可解释、适合课程设计演示。

## 6. 下单事务流程

后端在创建订单时使用事务和行锁，核心步骤：

1. 锁定当前用户购物车记录
2. 锁定对应菜品记录（FOR UPDATE）
3. 校验状态与库存
4. 后端计算总金额与每条明细小计
5. 写入 orders 与 order_items
6. 扣减库存
7. 清空购物车
8. 提交事务

这样可以避免并发下单导致的超卖和金额篡改。

## 7. 为何 order_items 要保存 dish_name 和 dish_price 快照

因为菜品名称和价格可能在后续运营中发生变化。订单作为历史交易记录，需要保留“下单时刻”的事实。

如果不做快照，查询历史订单时会展示当前价格，导致账单与实际支付不一致。

## 8. 为何订单金额不能由前端传入

订单金额属于安全敏感字段，必须由后端基于数据库实时价格计算。原因：

1. 前端可被篡改，请求体不可信
2. 防止恶意用户把价格改为 0.01
3. 保证金额与库存校验在同一事务内完成
4. 便于审计与对账

因此本系统采用：前端仅传菜品和数量，金额全部由后端计算并落库。
