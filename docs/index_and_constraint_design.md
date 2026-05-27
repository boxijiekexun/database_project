# 索引与约束设计

## 1. 主键设计

1. 所有核心表均采用自增主键 id（BIGINT UNSIGNED）。
2. 主键用于唯一标识记录并支撑外键引用。

## 2. 外键设计

1. merchants.user_id -> users.id
2. stores.merchant_id -> merchants.id
3. dish_categories.store_id -> stores.id
4. dishes.store_id -> stores.id
5. dishes.category_id -> dish_categories.id
6. cart_items.user_id/store_id/dish_id 分别引用 users/stores/dishes
7. orders.user_id/store_id 分别引用 users/stores
8. order_items.order_id/dish_id 分别引用 orders/dishes
9. reviews.order_id/user_id/store_id 分别引用 orders/users/stores

外键保证引用完整性，避免孤儿记录。

## 3. 唯一约束

1. users.username：登录账号唯一。
2. users.email：邮箱唯一。
3. users.phone：手机号唯一。
4. merchants.user_id：一个商家账号对应一个商家档案。
5. stores.merchant_id：当前业务中一个商家对应一个店铺。
6. stores.name：店铺名称唯一。
7. dish_categories(store_id, name)：店铺内分类名唯一。
8. dishes(store_id, name)：店铺内菜品名唯一。
9. cart_items(user_id, dish_id)：购物车同菜品唯一。
10. orders.order_no：订单号唯一。
11. reviews.order_id：一单一评。

## 4. CHECK 约束

1. stores.status IN (OPEN, CLOSED, DISABLED)
2. stores.rating >= 0 AND <= 5
3. dishes.price >= 0
4. dishes.stock >= 0
5. dishes.status IN (ON_SALE, OFF_SALE)
6. cart_items.quantity > 0
7. orders.total_amount >= 0
8. orders.status IN (PENDING_PAYMENT, PAID, ACCEPTED, PREPARING, DELIVERING, COMPLETED, CANCELLED)
9. order_items.quantity > 0
10. order_items.subtotal >= 0
11. reviews.rating BETWEEN 1 AND 5

CHECK 约束将业务规则下沉到数据库层，减少异常数据写入风险。

## 5. 索引设计

### 5.1 指定索引说明

1. users.username：登录时按用户名查询，提升认证效率。
2. users.email：支持邮箱查找与唯一校验。
3. users.phone：支持手机号查找与唯一校验。
4. stores.merchant_id：商家端查询自有店铺高频使用。
5. dishes.store_id：按店铺查询菜品列表高频使用。
6. dishes.category_id：按分类过滤菜品时使用。
7. orders.user_id：用户端查询我的订单高频使用。
8. orders.store_id：商家端与管理员按店铺查询订单使用。
9. order_items.order_id：订单详情查询明细高频使用。

### 5.2 索引对性能的作用

1. 登录效率：users.username 索引可快速定位账号，减少全表扫描。
2. 菜品查询效率：dishes.store_id 和 dishes.category_id 索引支撑店铺页与分类筛选。
3. 订单查询效率：orders.user_id、orders.store_id、order_items.order_id 索引分别服务用户、商家、管理员和详情查询场景。

## 6. 约束与索引协同

1. 唯一索引同时承担唯一约束和查询加速作用。
2. 外键保证数据正确，普通索引保证关联查询速度。
3. CHECK 约束保证值域正确，减少应用层兜底压力。
