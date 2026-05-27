# 概念结构设计（E-R）

## 1. 核心实体说明

1. 用户（users）：系统登录主体，区分 ADMIN、STUDENT、TEACHER、MERCHANT。
2. 商家（merchants）：商家扩展档案，与用户关联。
3. 店铺（stores）：商家经营实体。
4. 菜品分类（dish_categories）：店铺内分类。
5. 菜品（dishes）：可售卖商品，包含价格、库存、状态。
6. 购物车项（cart_items）：用户临时选购记录。
7. 订单（orders）：交易主单。
8. 订单明细（order_items）：订单内的菜品快照项。
9. 评价（reviews）：用户对订单的反馈记录。

## 2. 实体关系说明

1. 一个用户可以有多个订单（users 1:N orders）。
2. 一个商家对应一个店铺（merchants 1:1 stores，当前设计通过唯一约束保证）。
3. 一个店铺有多个菜品分类（stores 1:N dish_categories）。
4. 一个店铺有多个菜品（stores 1:N dishes）。
5. 一个订单属于一个用户和一个店铺（orders N:1 users，orders N:1 stores）。
6. 一个订单包含多个订单明细（orders 1:N order_items）。
7. 一个订单最多一个评价（orders 1:0..1 reviews，通过 reviews.order_id 唯一约束实现）。

## 3. E-R 图（Mermaid）

```mermaid
erDiagram
    users ||--o| merchants : has_profile
    merchants ||--|| stores : owns
    stores ||--o{ dish_categories : contains
    stores ||--o{ dishes : sells
    users ||--o{ cart_items : adds
    stores ||--o{ cart_items : belongs_to
    dishes ||--o{ cart_items : references

    users ||--o{ orders : places
    stores ||--o{ orders : receives
    orders ||--|{ order_items : includes
    dishes ||--o{ order_items : snapshots
    orders ||--o| reviews : has
    users ||--o{ reviews : writes
    stores ||--o{ reviews : receives

    users {
      bigint id PK
      varchar username UK
      varchar email UK
      varchar phone UK
      varchar role
      varchar status
    }

    merchants {
      bigint id PK
      bigint user_id FK_UK
      varchar merchant_name
      varchar audit_status
    }

    stores {
      bigint id PK
      bigint merchant_id FK_UK
      varchar name UK
      varchar status
      decimal rating
    }

    dish_categories {
      bigint id PK
      bigint store_id FK
      varchar name
      int sort_order
    }

    dishes {
      bigint id PK
      bigint store_id FK
      bigint category_id FK
      varchar name
      decimal price
      int stock
      varchar status
    }

    cart_items {
      bigint id PK
      bigint user_id FK
      bigint store_id FK
      bigint dish_id FK
      int quantity
    }

    orders {
      bigint id PK
      varchar order_no UK
      bigint user_id FK
      bigint store_id FK
      decimal total_amount
      varchar status
    }

    order_items {
      bigint id PK
      bigint order_id FK
      bigint dish_id FK
      varchar dish_name
      decimal dish_price
      int quantity
      decimal subtotal
    }

    reviews {
      bigint id PK
      bigint order_id FK_UK
      bigint user_id FK
      bigint store_id FK
      int rating
      text content
    }
```

## 4. 设计说明补充

1. order_items 的 dish_name、dish_price 为快照字段，保证历史订单可追溯。
2. cart_items 使用 user_id + dish_id 唯一约束，防止同一用户重复插入同一菜品行。
3. orders.status、dishes.status、reviews.rating 等关键字段配合 CHECK 约束限制非法值。
