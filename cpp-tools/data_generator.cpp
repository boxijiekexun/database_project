#include <fstream>
#include <iostream>
#include <sstream>
#include <string>

int main() {
    const std::string bcryptHash = "$2a$10$DmiEBphOd8Zgij0ek6haWOg.Wl5PdRMCMJSOo4wUW1z/.NtThWcm6";
    const int studentCount = 20;
    const int teacherCount = 10;
    const int merchantCount = 5;

    std::ofstream out("generated_data.sql", std::ios::out | std::ios::trunc);
    if (!out.is_open()) {
        std::cerr << "无法创建 generated_data.sql" << std::endl;
        return 1;
    }

    out << "-- 批量测试数据（由 data_generator.cpp 生成）\n";
    out << "-- 密码哈希可复用 schema.sql 中的测试哈希\n\n";

    out << "START TRANSACTION;\n\n";

    out << "-- 学生用户\n";
    for (int i = 1; i <= studentCount; ++i) {
        out << "INSERT INTO users (username, email, phone, password_hash, role, status) VALUES ('gen_student"
            << i << "', 'gen_student" << i << "@example.com', '1391000" << (1000 + i)
            << "', '" << bcryptHash << "', 'STUDENT', 'ACTIVE');\n";
    }

    out << "\n-- 教师用户\n";
    for (int i = 1; i <= teacherCount; ++i) {
        out << "INSERT INTO users (username, email, phone, password_hash, role, status) VALUES ('gen_teacher"
            << i << "', 'gen_teacher" << i << "@example.com', '1392000" << (1000 + i)
            << "', '" << bcryptHash << "', 'TEACHER', 'ACTIVE');\n";
    }

    out << "\n-- 商家用户\n";
    for (int i = 1; i <= merchantCount; ++i) {
        out << "INSERT INTO users (username, email, phone, password_hash, role, status) VALUES ('gen_merchant"
            << i << "', 'gen_merchant" << i << "@example.com', '1393000" << (1000 + i)
            << "', '" << bcryptHash << "', 'MERCHANT', 'ACTIVE');\n";
    }

    out << "\n-- 商家档案\n";
    out << "INSERT INTO merchants (user_id, merchant_name, contact_name, contact_phone, audit_status)\n";
    out << "SELECT id, CONCAT('生成商户', id), username, phone, 'APPROVED'\n";
    out << "FROM users WHERE username LIKE 'gen_merchant%';\n";

    out << "\n-- 店铺\n";
    out << "INSERT INTO stores (merchant_id, name, description, phone, address, status, rating, sales_count)\n";
    out << "SELECT m.id, CONCAT('生成店铺', m.id), '课程设计测试店铺', u.phone, CONCAT('校园测试地址', m.id), 'OPEN', 4.50, 0\n";
    out << "FROM merchants m JOIN users u ON u.id = m.user_id WHERE u.username LIKE 'gen_merchant%';\n";

    out << "\n-- 菜品分类\n";
    out << "INSERT INTO dish_categories (store_id, name, sort_order)\n";
    out << "SELECT id, '主食', 1 FROM stores WHERE name LIKE '生成店铺%';\n";
    out << "INSERT INTO dish_categories (store_id, name, sort_order)\n";
    out << "SELECT id, '饮品', 2 FROM stores WHERE name LIKE '生成店铺%';\n";

    out << "\n-- 菜品\n";
    out << "INSERT INTO dishes (store_id, category_id, name, description, price, stock, status, image_url, sales_count)\n";
    out << "SELECT c.store_id, c.id, CONCAT('测试主食-', c.store_id), '自动生成主食', 16.80, 80, 'ON_SALE', NULL, 0\n";
    out << "FROM dish_categories c JOIN stores s ON s.id = c.store_id\n";
    out << "WHERE c.name = '主食' AND s.name LIKE '生成店铺%';\n";

    out << "INSERT INTO dishes (store_id, category_id, name, description, price, stock, status, image_url, sales_count)\n";
    out << "SELECT c.store_id, c.id, CONCAT('测试饮品-', c.store_id), '自动生成饮品', 5.50, 120, 'ON_SALE', NULL, 0\n";
    out << "FROM dish_categories c JOIN stores s ON s.id = c.store_id\n";
    out << "WHERE c.name = '饮品' AND s.name LIKE '生成店铺%';\n";

    out << "\nCOMMIT;\n";

    out.close();
    std::cout << "已生成 generated_data.sql" << std::endl;
    std::cout << "提示：如果你的密码哈希策略不同，请替换 SQL 中的 password_hash。" << std::endl;
    return 0;
}
