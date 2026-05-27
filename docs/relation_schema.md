# 关系模式设计

## 1. 从 E-R 到关系模式

系统 E-R 图转换为 9 张核心关系表：users、merchants、stores、dish_categories、dishes、cart_items、orders、order_items、reviews。

下文按表说明字段、主键、外键、唯一约束和用途。

## 2. 核心表清单

### 2.1 users

1. 字段：id, username, email, phone, password_hash, role, status, created_at, updated_at
2. 主键：id
3. 外键：无
4. 唯一约束：username, email, phone
5. 说明：账号中心表，承载认证与角色控制。

### 2.2 merchants

1. 字段：id, user_id, merchant_name, contact_name, contact_phone, audit_status, created_at, updated_at
2. 主键：id
3. 外键：user_id -> users.id
4. 唯一约束：user_id
5. 说明：商家扩展资料，每个商家账号对应一个商家档案。

### 2.3 stores

1. 字段：id, merchant_id, name, description, phone, address, status, rating, sales_count, created_at, updated_at
2. 主键：id
3. 外键：merchant_id -> merchants.id
4. 唯一约束：merchant_id, name
5. 说明：店铺主数据，支持状态、评分和销量统计。

### 2.4 dish_categories

1. 字段：id, store_id, name, sort_order, created_at, updated_at
2. 主键：id
3. 外键：store_id -> stores.id
4. 唯一约束：(store_id, name)
5. 说明：店铺内分类管理。

### 2.5 dishes

1. 字段：id, store_id, category_id, name, description, price, stock, status, image_url, sales_count, created_at, updated_at
2. 主键：id
3. 外键：store_id -> stores.id, category_id -> dish_categories.id
4. 唯一约束：(store_id, name)
5. 说明：菜品主数据，支持库存、价格和上下架状态。

### 2.6 cart_items

1. 字段：id, user_id, store_id, dish_id, quantity, created_at, updated_at
2. 主键：id
3. 外键：user_id -> users.id, store_id -> stores.id, dish_id -> dishes.id
4. 唯一约束：(user_id, dish_id)
5. 说明：用户购物车临时数据，限制同用户同菜品只有一条记录。

### 2.7 orders

1. 字段：id, order_no, user_id, store_id, total_amount, status, remark, created_at, updated_at
2. 主键：id
3. 外键：user_id -> users.id, store_id -> stores.id
4. 唯一约束：order_no
5. 说明：订单主表，记录交易状态和金额。

### 2.8 order_items

1. 字段：id, order_id, dish_id, dish_name, dish_price, quantity, subtotal, created_at
2. 主键：id
3. 外键：order_id -> orders.id, dish_id -> dishes.id
4. 唯一约束：无
5. 说明：订单明细表，保存下单时快照字段，保障历史数据一致性。

### 2.9 reviews

1. 字段：id, order_id, user_id, store_id, rating, content, created_at
2. 主键：id
3. 外键：order_id -> orders.id, user_id -> users.id, store_id -> stores.id
4. 唯一约束：order_id
5. 说明：一单一评，rating 取值受约束。

## 3. 重点说明

### 3.1 users

通过 username、email、phone 三重唯一约束确保账号标识不冲突，支撑登录和注册校验。

### 3.2 stores

merchant_id 唯一约束保证当前系统一个商家对应一个店铺，便于商家端简化管理。

### 3.3 dishes

通过 (store_id, name) 唯一约束限制店铺内重名菜品，避免运营混乱。

### 3.4 cart_items

通过 (user_id, dish_id) 唯一约束实现购物车合并逻辑，减少重复行。

### 3.5 orders

order_no 唯一约束保证交易单号可追溯；status 和 total_amount 受 CHECK 约束保护。

### 3.6 order_items

快照字段 dish_name、dish_price 避免后续菜品改名或改价影响历史订单展示与对账。

### 3.7 reviews

order_id 唯一约束直接实现一单一评规则，减少应用层重复评价分支复杂度。
