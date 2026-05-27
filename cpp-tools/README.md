# C++ 辅助工具

本目录包含两个 C++17 工具：

1. data_generator.cpp：生成批量测试 SQL 到 generated_data.sql
2. sales_analyzer.cpp：读取订单导出 CSV 并统计销售数据

## 编译

Linux/macOS:

```bash
g++ -std=c++17 data_generator.cpp -o data_generator
g++ -std=c++17 sales_analyzer.cpp -o sales_analyzer
```

Windows PowerShell:

```powershell
g++ -std=c++17 data_generator.cpp -o data_generator.exe
g++ -std=c++17 sales_analyzer.cpp -o sales_analyzer.exe
```

## 运行

Linux/macOS:

```bash
./data_generator
./sales_analyzer orders_export.csv
```

Windows PowerShell:

```powershell
.\data_generator.exe
.\sales_analyzer.exe orders_export.csv
```

## 输入格式

sales_analyzer 读取 CSV，字段顺序固定为：

order_id,store_name,dish_name,quantity,subtotal

示例：

```csv
order_id,store_name,dish_name,quantity,subtotal
1,第一食堂窗口,宫保鸡丁饭,2,37.60
1,第一食堂窗口,可乐,1,5.00
```
