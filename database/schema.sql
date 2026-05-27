DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS dishes;
DROP TABLE IF EXISTS dish_categories;
DROP TABLE IF EXISTS stores;
DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS canteens;
DROP TABLE IF EXISTS dining_areas;
DROP TABLE IF EXISTS student_profiles;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NULL,
    phone VARCHAR(30) NULL,
    password_hash VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_phone (phone),
    KEY idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE student_profiles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    real_name VARCHAR(100) NULL,
    student_no VARCHAR(50) NULL,
    college VARCHAR(100) NULL,
    class_name VARCHAR(100) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_student_profiles_user_id (user_id),
    UNIQUE KEY uk_student_profiles_student_no (student_no),
    CONSTRAINT fk_student_profiles_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE dining_areas (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dining_areas_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE canteens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    area_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    specialties VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_canteens_name (name),
    KEY idx_canteens_area_id (area_id),
    CONSTRAINT fk_canteens_area FOREIGN KEY (area_id) REFERENCES dining_areas(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE merchants (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    merchant_name VARCHAR(100) NOT NULL,
    contact_name VARCHAR(100) NULL,
    contact_phone VARCHAR(30) NULL,
    audit_status VARCHAR(20) NOT NULL DEFAULT 'APPROVED',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_merchants_user_id (user_id),
    CONSTRAINT fk_merchants_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stores (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    merchant_id BIGINT UNSIGNED NOT NULL,
    canteen_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    phone VARCHAR(30) NULL,
    address VARCHAR(255) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    rating DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    sales_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_stores_merchant_id (merchant_id),
    UNIQUE KEY uk_stores_name (name),
    KEY idx_stores_merchant_id (merchant_id),
    KEY idx_stores_canteen_id (canteen_id),
    CONSTRAINT fk_stores_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id),
    CONSTRAINT fk_stores_canteen FOREIGN KEY (canteen_id) REFERENCES canteens(id),
    CONSTRAINT chk_stores_status CHECK (status IN ('OPEN', 'CLOSED', 'DISABLED')),
    CONSTRAINT chk_stores_rating CHECK (rating >= 0 AND rating <= 5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE dish_categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    store_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dish_categories_store_name (store_id, name),
    KEY idx_dish_categories_store_id (store_id),
    CONSTRAINT fk_dish_categories_store FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE dishes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    store_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    stock INT UNSIGNED NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'ON_SALE',
    image_url VARCHAR(255) NULL,
    sales_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dishes_store_name (store_id, name),
    KEY idx_dishes_store_id (store_id),
    KEY idx_dishes_category_id (category_id),
    CONSTRAINT fk_dishes_store FOREIGN KEY (store_id) REFERENCES stores(id),
    CONSTRAINT fk_dishes_category FOREIGN KEY (category_id) REFERENCES dish_categories(id),
    CONSTRAINT chk_dishes_price CHECK (price >= 0),
    CONSTRAINT chk_dishes_stock CHECK (stock >= 0),
    CONSTRAINT chk_dishes_status CHECK (status IN ('ON_SALE', 'OFF_SALE'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE cart_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    store_id BIGINT UNSIGNED NOT NULL,
    dish_id BIGINT UNSIGNED NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_cart_items_user_dish (user_id, dish_id),
    KEY idx_cart_items_user_id (user_id),
    KEY idx_cart_items_store_id (store_id),
    KEY idx_cart_items_dish_id (dish_id),
    CONSTRAINT fk_cart_items_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_cart_items_store FOREIGN KEY (store_id) REFERENCES stores(id),
    CONSTRAINT fk_cart_items_dish FOREIGN KEY (dish_id) REFERENCES dishes(id),
    CONSTRAINT chk_cart_items_quantity CHECK (quantity > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_no VARCHAR(50) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    store_id BIGINT UNSIGNED NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL,
    remark TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_orders_order_no (order_no),
    KEY idx_orders_user_id (user_id),
    KEY idx_orders_store_id (store_id),
    KEY idx_orders_status (status),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_orders_store FOREIGN KEY (store_id) REFERENCES stores(id),
    CONSTRAINT chk_orders_total_amount CHECK (total_amount >= 0),
    CONSTRAINT chk_orders_status CHECK (status IN ('PENDING_PAYMENT', 'PAID', 'ACCEPTED', 'PREPARING', 'DELIVERING', 'COMPLETED', 'CANCELLED'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE order_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id BIGINT UNSIGNED NOT NULL,
    dish_id BIGINT UNSIGNED NOT NULL,
    dish_name VARCHAR(100) NOT NULL,
    dish_price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    quantity INT UNSIGNED NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_order_items_order_id (order_id),
    KEY idx_order_items_dish_id (dish_id),
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_order_items_dish FOREIGN KEY (dish_id) REFERENCES dishes(id),
    CONSTRAINT chk_order_items_quantity CHECK (quantity > 0),
    CONSTRAINT chk_order_items_subtotal CHECK (subtotal >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE reviews (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    store_id BIGINT UNSIGNED NOT NULL,
    rating INT NOT NULL,
    content TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_reviews_order_id (order_id),
    KEY idx_reviews_user_id (user_id),
    KEY idx_reviews_store_id (store_id),
    CONSTRAINT fk_reviews_order FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_reviews_store FOREIGN KEY (store_id) REFERENCES stores(id),
    CONSTRAINT chk_reviews_rating CHECK (rating BETWEEN 1 AND 5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO users (id, username, email, phone, password_hash, role, status) VALUES
    (1, 'admin', 'admin@example.com', '13000000000', '$2a$10$DmiEBphOd8Zgij0ek6haWOg.Wl5PdRMCMJSOo4wUW1z/.NtThWcm6', 'ADMIN', 'ACTIVE'),
    (2, 'student01', 'student01@example.com', '13000000001', '$2a$10$DmiEBphOd8Zgij0ek6haWOg.Wl5PdRMCMJSOo4wUW1z/.NtThWcm6', 'STUDENT', 'ACTIVE'),
    (3, 'teacher01', 'teacher01@example.com', '13000000002', '$2a$10$DmiEBphOd8Zgij0ek6haWOg.Wl5PdRMCMJSOo4wUW1z/.NtThWcm6', 'TEACHER', 'ACTIVE'),
    (4, 'merchant01', 'merchant01@example.com', '13000000003', '$2a$10$DmiEBphOd8Zgij0ek6haWOg.Wl5PdRMCMJSOo4wUW1z/.NtThWcm6', 'MERCHANT', 'ACTIVE');

INSERT IGNORE INTO dining_areas (id, name, description) VALUES
    (1, '旧饭', '学校老校区生活区，食堂种类丰富，价格实惠'),
    (2, '新饭', '新校区主要就餐区域，环境较新，窗口选择多'),
    (3, '多饭', '多功能生活区食堂，兼顾快餐、小吃和饮品');

INSERT IGNORE INTO canteens (id, area_id, name, description, specialties) VALUES
    (1, 1, '第一食堂', '老校区人气最高的综合食堂，提供盖饭、炒饭、面食和小吃。', '宫保鸡丁饭、牛肉炒饭、番茄鸡蛋面'),
    (2, 1, '第二食堂', '以家常菜和经济套餐为主，适合日常就餐。', '鱼香肉丝饭、土豆牛肉饭、青椒肉丝面'),
    (3, 1, '第三食堂', '靠近教学楼，出餐快，适合课间快速用餐。', '鸡腿饭、卤肉饭、热干面'),
    (4, 2, '第四食堂', '新校区现代化食堂，环境整洁，窗口丰富。', '麻辣香锅、黄焖鸡米饭、烤肉拌饭'),
    (5, 2, '第五食堂', '以轻食、饮品和特色小吃为主，适合晚餐和加餐。', '沙拉、三明治、柠檬茶、鸡米花'),
    (6, 3, '第六食堂', '多功能生活区食堂，提供快餐、小吃和饮品。', '汉堡、炸鸡、薯条、可乐');

INSERT IGNORE INTO merchants (id, user_id, merchant_name, contact_name, contact_phone, audit_status) VALUES
    (1, 4, 'merchant01商户', 'merchant01', '13000000003', 'APPROVED');

INSERT IGNORE INTO stores (id, merchant_id, canteen_id, name, description, phone, address, status, rating, sales_count) VALUES
    (1, 1, 1, '第一食堂窗口', '校园热门外卖窗口', '13000000003', '校园东区食堂1楼', 'OPEN', 4.80, 1280);

INSERT IGNORE INTO dish_categories (id, store_id, name, sort_order) VALUES
    (1, 1, '主食', 1),
    (2, 1, '饮品', 2),
    (3, 1, '小吃', 3);

INSERT IGNORE INTO dishes (id, store_id, category_id, name, description, price, stock, status, image_url, sales_count) VALUES
    (1, 1, 1, '宫保鸡丁饭', '招牌热销盖饭', 18.80, 50, 'ON_SALE', NULL, 320),
    (2, 1, 1, '番茄鸡蛋面', '经典家常面食', 15.00, 40, 'ON_SALE', NULL, 210),
    (3, 1, 1, '香辣鸡腿堡', '香辣口味汉堡', 16.50, 35, 'ON_SALE', NULL, 180),
    (4, 1, 1, '牛肉炒饭', '粒粒分明，牛肉香浓', 19.00, 30, 'ON_SALE', NULL, 260),
    (5, 1, 2, '可乐', '冰镇碳酸饮料', 5.00, 100, 'ON_SALE', NULL, 540),
    (6, 1, 2, '柠檬茶', '清爽解腻', 6.50, 80, 'ON_SALE', NULL, 430),
    (7, 1, 3, '薯条', '现炸薯条', 7.00, 60, 'ON_SALE', NULL, 390),
    (8, 1, 3, '鸡米花', '酥脆鸡肉小食', 9.50, 55, 'ON_SALE', NULL, 360);
