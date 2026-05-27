# 事务设计

## 1. 为什么下单必须使用事务

下单同时涉及多张表与多步骤写操作：orders、order_items、dishes、cart_items。若不使用事务，任意中间失败都可能导致脏数据，例如已扣库存但未生成订单，或订单生成但购物车未清空。

## 2. 下单事务步骤

以下为简化流程（伪代码级）：

1. 开启事务。
2. 查询并锁定当前用户购物车项。
3. 校验购物车不为空且同一店铺。
4. 锁定相关菜品记录并校验：
   - 菜品状态为 ON_SALE
   - 库存足够
5. 后端基于数据库价格计算每项小计和订单总金额。
6. 写入 orders 主单。
7. 批量写入 order_items 明细。
8. 扣减 dishes.stock。
9. 清空 cart_items。
10. 提交事务。

## 3. 失败回滚原则

任意一步出现异常均执行回滚，保证：

1. 不会出现半完成订单。
2. 库存与订单状态保持一致。
3. 购物车状态与订单状态一致。

## 4. 为什么不能相信前端金额

1. 前端请求可被篡改，金额不可信。
2. 正确金额应由后端根据数据库价格和数量计算。
3. 金额计算必须和库存校验在同一事务内完成，防止并发价格/库存变化导致不一致。

## 5. 为什么 order_items 保存 dish_name 和 dish_price

1. 订单属于历史交易，需保留下单时的事实。
2. 菜品名称和价格可能后续调整。
3. 快照字段保证订单详情、账单和统计在历史维度可追溯。

## 6. 取消订单恢复库存事务

取消订单也属于多步骤更新，需事务化：

1. 锁定目标订单并校验状态（仅待支付可取消）。
2. 查询订单明细。
3. 更新订单状态为 CANCELLED。
4. 按明细逐项恢复菜品库存。
5. 提交事务。

这样可避免取消过程中出现状态更新成功但库存未恢复的问题。

## 7. 伪 SQL 示例

```sql
BEGIN;
SELECT ... FROM cart_items WHERE user_id = ? FOR UPDATE;
SELECT ... FROM dishes WHERE id IN (...) FOR UPDATE;
INSERT INTO orders (...);
INSERT INTO order_items (...);
UPDATE dishes SET stock = stock - ? WHERE id = ?;
DELETE FROM cart_items WHERE user_id = ?;
COMMIT;
-- on error: ROLLBACK;
```
