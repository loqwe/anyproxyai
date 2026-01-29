package database

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// ModelRoute 模型路由表结构
type ModelRoute struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Model     string    `json:"model"`
	APIUrl    string    `json:"api_url"`
	APIKey    string    `json:"api_key"`
	Group     string    `json:"group"`
	Format    string    `json:"format"` // 格式类型 (openai, claude, gemini)
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RequestLog 请求日志表结构
type RequestLog struct {
	ID             int64     `json:"id"`
	Model          string    `json:"model"`           // 请求的模型名
	ProviderModel  string    `json:"provider_model"` // 实际使用的提供商模型
	ProviderName   string    `json:"provider_name"`  // 提供商/路由名称
	RouteID        int64     `json:"route_id"`
	RequestTokens  int       `json:"request_tokens"`
	ResponseTokens int       `json:"response_tokens"`
	TotalTokens    int       `json:"total_tokens"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message"`
	Style          string    `json:"style"`           // 请求类型: openai, claude, gemini
	UserAgent      string    `json:"user_agent"`      // 用户代理
	RemoteIP       string    `json:"remote_ip"`       // 客户端IP
	ProxyTimeMs    int64     `json:"proxy_time_ms"`   // 代理总耗时(毫秒)
	FirstChunkMs   int64     `json:"first_chunk_ms"` // 首字节时间(毫秒)
	IsStream       bool      `json:"is_stream"`       // 是否流式请求
	CreatedAt      time.Time `json:"created_at"`
}

// HourlyStats 每小时统计表结构（压缩后的数据）
type HourlyStats struct {
	ID             int64  `json:"id"`
	Date           string `json:"date"`            // 日期 YYYY-MM-DD
	Hour           int    `json:"hour"`            // 小时 0-23
	Model          string `json:"model"`           // 模型名
	RequestCount   int    `json:"request_count"`   // 请求次数
	RequestTokens  int    `json:"request_tokens"`  // 输入 token
	ResponseTokens int    `json:"response_tokens"` // 输出 token
	TotalTokens    int    `json:"total_tokens"`    // 总 token
	SuccessCount   int    `json:"success_count"`   // 成功次数
	FailCount      int    `json:"fail_count"`      // 失败次数
}

// UsageSummary 用量汇总表结构
type UsageSummary struct {
	ID             int64  `json:"id"`
	PeriodType     string `json:"period_type"`     // week, year, total
	PeriodKey      string `json:"period_key"`      // 2026-W02, 2026, total
	RequestCount   int64  `json:"request_count"`   // 请求次数
	RequestTokens  int64  `json:"request_tokens"`  // 输入 token
	ResponseTokens int64  `json:"response_tokens"` // 输出 token
	TotalTokens    int64  `json:"total_tokens"`    // 总 token
	SuccessCount   int64  `json:"success_count"`   // 成功次数
	FailCount      int64  `json:"fail_count"`      // 失败次数
	UpdatedAt      string `json:"updated_at"`      // 更新时间
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	// 执行数据库迁移
	if err := migrateDB(db); err != nil {
		log.Warnf("Database migration warning: %v", err)
	}

	log.Info("Database initialized successfully")
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS model_routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		model TEXT NOT NULL,
		api_url TEXT NOT NULL,
		api_key TEXT,
		"group" TEXT,
		format TEXT DEFAULT 'openai',
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_model_routes_model ON model_routes(model);
	CREATE INDEX IF NOT EXISTS idx_model_routes_enabled ON model_routes(enabled);
	CREATE INDEX IF NOT EXISTS idx_model_routes_group ON model_routes("group");

	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model TEXT NOT NULL,
		provider_model TEXT,
		provider_name TEXT,
		route_id INTEGER,
		request_tokens INTEGER DEFAULT 0,
		response_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		success INTEGER DEFAULT 1,
		error_message TEXT,
		style TEXT,
		user_agent TEXT,
		remote_ip TEXT,
		proxy_time_ms INTEGER DEFAULT 0,
		first_chunk_ms INTEGER DEFAULT 0,
		is_stream INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (route_id) REFERENCES model_routes(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model);
	CREATE INDEX IF NOT EXISTS idx_request_logs_route_id ON request_logs(route_id);
	CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_request_logs_success ON request_logs(success);
	CREATE INDEX IF NOT EXISTS idx_request_logs_provider_name ON request_logs(provider_name);
	CREATE INDEX IF NOT EXISTS idx_request_logs_style ON request_logs(style);
	CREATE INDEX IF NOT EXISTS idx_request_logs_user_agent ON request_logs(user_agent);

	-- 每小时统计表（压缩后的数据）
	CREATE TABLE IF NOT EXISTS hourly_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		hour INTEGER NOT NULL,
		model TEXT NOT NULL,
		request_count INTEGER DEFAULT 0,
		request_tokens INTEGER DEFAULT 0,
		response_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		success_count INTEGER DEFAULT 0,
		fail_count INTEGER DEFAULT 0,
		UNIQUE(date, hour, model)
	);

	CREATE INDEX IF NOT EXISTS idx_hourly_stats_date ON hourly_stats(date);
	CREATE INDEX IF NOT EXISTS idx_hourly_stats_model ON hourly_stats(model);

	-- 用量汇总表（周/年/总用量）
	CREATE TABLE IF NOT EXISTS usage_summary (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		period_type TEXT NOT NULL,
		period_key TEXT NOT NULL,
		request_count INTEGER DEFAULT 0,
		request_tokens INTEGER DEFAULT 0,
		response_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		success_count INTEGER DEFAULT 0,
		fail_count INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(period_type, period_key)
	);

	CREATE INDEX IF NOT EXISTS idx_usage_summary_type ON usage_summary(period_type);
	CREATE INDEX IF NOT EXISTS idx_usage_summary_key ON usage_summary(period_key);
	`

	_, err := db.Exec(schema)
	return err
}

// migrateDB 执行数据库迁移，确保表结构是最新的
func migrateDB(db *sql.DB) error {
	// 添加 format 列（如果不存在）
	db.Exec(`ALTER TABLE model_routes ADD COLUMN format TEXT DEFAULT 'openai'`)

	// 检查并迁移 request_logs 表的 id 字段为 BIGINT 兼容
	// SQLite 的 INTEGER PRIMARY KEY 已经是 64 位，无需额外迁移

	// 检查 hourly_stats 表是否存在所有必要的列
	db.Exec(`ALTER TABLE hourly_stats ADD COLUMN success_count INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE hourly_stats ADD COLUMN fail_count INTEGER DEFAULT 0`)

	// 检查 usage_summary 表是否存在所有必要的列
	db.Exec(`ALTER TABLE usage_summary ADD COLUMN success_count INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE usage_summary ADD COLUMN fail_count INTEGER DEFAULT 0`)

	// 添加 request_logs 的新字段（如果不存在）
	db.Exec(`ALTER TABLE request_logs ADD COLUMN provider_model TEXT`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN provider_name TEXT`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN style TEXT`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN user_agent TEXT`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN remote_ip TEXT`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN proxy_time_ms INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN first_chunk_ms INTEGER DEFAULT 0`)
	db.Exec(`ALTER TABLE request_logs ADD COLUMN is_stream INTEGER DEFAULT 0`)

	log.Info("Database migration completed")
	return nil
}
