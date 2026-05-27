#include <fstream>
#include <iomanip>
#include <iostream>
#include <map>
#include <sstream>
#include <string>

struct Record {
    long long orderId = 0;
    std::string storeName;
    std::string dishName;
    int quantity = 0;
    double subtotal = 0.0;
};

bool parseLine(const std::string& line, Record& rec) {
    std::stringstream ss(line);
    std::string token;

    if (!std::getline(ss, token, ',')) return false;
    rec.orderId = std::stoll(token);

    if (!std::getline(ss, rec.storeName, ',')) return false;
    if (!std::getline(ss, rec.dishName, ',')) return false;

    if (!std::getline(ss, token, ',')) return false;
    rec.quantity = std::stoi(token);

    if (!std::getline(ss, token, ',')) return false;
    rec.subtotal = std::stod(token);

    return true;
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "用法: sales_analyzer <orders_export.csv>" << std::endl;
        return 1;
    }

    std::ifstream in(argv[1]);
    if (!in.is_open()) {
        std::cerr << "无法打开文件: " << argv[1] << std::endl;
        return 1;
    }

    std::map<std::string, double> storeRevenue;
    std::map<std::string, long long> dishSales;
    double totalRevenue = 0.0;

    std::string line;
    bool firstLine = true;
    while (std::getline(in, line)) {
        if (line.empty()) continue;
        if (firstLine) {
            firstLine = false;
            if (line.find("order_id") == 0) {
                continue;
            }
        }

        Record rec;
        try {
            if (!parseLine(line, rec)) {
                std::cerr << "跳过无法解析的行: " << line << std::endl;
                continue;
            }
        } catch (const std::exception&) {
            std::cerr << "跳过格式错误的行: " << line << std::endl;
            continue;
        }

        storeRevenue[rec.storeName] += rec.subtotal;
        dishSales[rec.dishName] += rec.quantity;
        totalRevenue += rec.subtotal;
    }

    std::cout << std::fixed << std::setprecision(2);

    std::cout << "===== 每个店铺销售额 =====" << std::endl;
    for (const auto& item : storeRevenue) {
        std::cout << item.first << ": " << item.second << std::endl;
    }

    std::cout << "\n===== 每个菜品销量 =====" << std::endl;
    for (const auto& item : dishSales) {
        std::cout << item.first << ": " << item.second << std::endl;
    }

    std::cout << "\n===== 总销售额 =====" << std::endl;
    std::cout << totalRevenue << std::endl;

    return 0;
}
