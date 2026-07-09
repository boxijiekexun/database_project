// 文件作用：
//   - 统一读取后端运行配置，避免把端口、数据库和限流参数写死在代码里。
//
// 业务说明：
//   - 当前课程设计主要通过环境变量控制运行行为。
//   - 没有配置环境变量时，会回退到适合本地演示的默认值。
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Address 是 HTTP 服务监听地址，例如 :8080。
	Address string
	// MySQLDSN 是连接 MySQL 的完整 DSN。
	MySQLDSN string
	// JWTSecret 是签发登录 token 时使用的密钥。
	JWTSecret string
	// TokenTTLHours 表示 token 默认有效小时数。
	TokenTTLHours int
	// DBMaxOpenConns 控制最大数据库连接数。
	DBMaxOpenConns int
	// DBMaxIdleConns 控制最大空闲连接数。
	DBMaxIdleConns int
	// DBConnMaxLifetime 控制单个数据库连接的最长生命周期。
	DBConnMaxLifetime time.Duration
	// DBConnMaxIdleTime 控制空闲连接最长保留时间。
	DBConnMaxIdleTime time.Duration
	// DBQueryTimeout 控制单次数据库操作超时时间。
	DBQueryTimeout time.Duration
	// LoginRateLimit 控制登录接口每分钟限流次数。
	LoginRateLimit int
	// OrderRateLimit 控制下单接口每分钟限流次数。
	OrderRateLimit int
	// UploadRateLimit 控制上传接口每分钟限流次数。
	UploadRateLimit int
	// RouteCacheTTL 控制食堂和店铺浏览接口缓存时间。
	RouteCacheTTL time.Duration
	// RecommendCacheTTL 控制 AI 推荐接口缓存时间。
	RecommendCacheTTL time.Duration
}

// Load 从环境变量读取配置，并为缺失项填入默认值。
func Load() Config {
	// 读取 HTTP 服务监听地址，如果没有设置环境变量，则使用默认值。
	address := os.Getenv("ADDR")
	if address == "" {
		address = ":8080"
	}

	// 读取 JWT_SECRET，如果没有设置环境变量，则使用默认值。
	// JWT 配置
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-key"
	}

	// 读取 MySQL DSN，如果没有设置环境变量，则使用默认值。
	// 数据库连接配置
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:your_mysql_password@tcp(127.0.0.1:3306)/campus_takeout?parseTime=true&charset=utf8mb4&loc=Local"
	}

	// return Config 结构体，包含所有配置项。
	return Config{
		Address:       address,
		MySQLDSN:      dsn,
		JWTSecret:     secret,
		TokenTTLHours: getEnvInt("TOKEN_TTL_HOURS", 24),
		// 数据库连接池配置
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 50),                                          // 最多开多少数据库连接
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),                                          // 空闲时保留多少连接
		DBConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,  // 连接多久后重建
		DBConnMaxIdleTime: time.Duration(getEnvInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 10)) * time.Minute, // 空闲连接多久后释放
		DBQueryTimeout:    time.Duration(getEnvInt("DB_QUERY_TIMEOUT_SECONDS", 3)) * time.Second,
		// 限流配置
		LoginRateLimit:  getEnvInt("LOGIN_RATE_LIMIT_PER_MIN", 30),
		OrderRateLimit:  getEnvInt("ORDER_RATE_LIMIT_PER_MIN", 20),
		UploadRateLimit: getEnvInt("UPLOAD_RATE_LIMIT_PER_MIN", 20),
		// 缓存配置
		RouteCacheTTL:     time.Duration(getEnvInt("ROUTE_CACHE_TTL_SECONDS", 30)) * time.Second,
		RecommendCacheTTL: time.Duration(getEnvInt("RECOMMEND_CACHE_TTL_SECONDS", 60)) * time.Second,
	}
}

// getEnvInt 读取正整数环境变量；读取失败时返回默认值。
func getEnvInt(key string, fallback int) int {
	// 读取环境变量
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
