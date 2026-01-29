package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"openai-router-go/internal/database"

	log "github.com/sirupsen/logrus"
)

type RouteService struct {
	db *sql.DB
}

func NewRouteService(db *sql.DB) *RouteService {
	return &RouteService{db: db}
}

// GetAllRoutes 获取所有路由
func (s *RouteService) GetAllRoutes() ([]database.ModelRoute, error) {
	query := `SELECT id, name, model, api_url, api_key, "group", COALESCE(format, 'openai'), enabled, created_at, updated_at
	          FROM model_routes ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []database.ModelRoute
	for rows.Next() {
		var route database.ModelRoute
		err := rows.Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl, &route.APIKey,
			&route.Group, &route.Format, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// GetRouteByModel 根据模型名获取路由(支持负载均衡和后缀匹配)
// 匹配规则: 精确匹配 + 后缀匹配 一起参与负载均衡
// 例如: 请求 "gemini-3-flash" 可匹配 "gemini-3-flash" 和 "流式抗截断/gemini-3-flash"
func (s *RouteService) GetRouteByModel(model string) (*database.ModelRoute, error) {
	// 精确匹配 + 后缀匹配 一起参与负载均衡
	query := `SELECT id, name, model, api_url, api_key, "group", COALESCE(format, 'openai'), enabled, created_at, updated_at
	          FROM model_routes 
	          WHERE (model = ? OR model LIKE '%/' || ?) AND enabled = 1 
	          ORDER BY RANDOM() LIMIT 1`

	var route database.ModelRoute
	err := s.db.QueryRow(query, model, model).Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl,
		&route.APIKey, &route.Group, &route.Format, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("model not found: %s", model)
	}
	if err != nil {
		return nil, err
	}

	// 如果是后缀匹配，记录日志
	if route.Model != model {
		log.Infof("[Suffix Match] '%s' matched to '%s'", model, route.Model)
	}
	return &route, nil
}

// GetAllRoutesByModel 根据模型名获取所有匹配的路由(用于 Fallback 故障转移)
// 返回所有匹配的路由，随机排序用于负载均衡
// 匹配规则: 精确匹配 + 后缀匹配
func (s *RouteService) GetAllRoutesByModel(model string) ([]database.ModelRoute, error) {
	query := `SELECT id, name, model, api_url, api_key, "group", COALESCE(format, 'openai'), enabled, created_at, updated_at
	          FROM model_routes 
	          WHERE (model = ? OR model LIKE '%/' || ?) AND enabled = 1 
	          ORDER BY RANDOM()`

	rows, err := s.db.Query(query, model, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []database.ModelRoute
	for rows.Next() {
		var route database.ModelRoute
		err := rows.Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl, &route.APIKey,
			&route.Group, &route.Format, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("model not found: %s", model)
	}

	return routes, nil
}

// GetRouteByID 根据路由ID获取路由
func (s *RouteService) GetRouteByID(id int64) (*database.ModelRoute, error) {
	query := `SELECT id, name, model, api_url, api_key, "group", COALESCE(format, 'openai'), enabled, created_at, updated_at
	          FROM model_routes WHERE id = ? AND enabled = 1`

	var route database.ModelRoute
	err := s.db.QueryRow(query, id).Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl,
		&route.APIKey, &route.Group, &route.Format, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("route not found: %d", id)
	}
	if err != nil {
		return nil, err
	}

	return &route, nil
}

// AddRoute 添加路由
func (s *RouteService) AddRoute(name, model, apiUrl, apiKey, group, format string) error {
	query := `INSERT INTO model_routes (name, model, api_url, api_key, "group", format, enabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`

	now := time.Now()
	_, err := s.db.Exec(query, name, model, apiUrl, apiKey, group, format, now, now)
	if err != nil {
		log.Errorf("Failed to add route: %v", err)
		return err
	}

	log.Infof("Route added: %s -> %s (%s) [%s]", model, apiUrl, name, format)
	return nil
}

// UpdateRoute 更新路由
func (s *RouteService) UpdateRoute(id int64, name, model, apiUrl, apiKey, group, format string) error {
	query := `UPDATE model_routes SET name = ?, model = ?, api_url = ?, api_key = ?, "group" = ?, format = ?, updated_at = ?
	          WHERE id = ?`

	result, err := s.db.Exec(query, name, model, apiUrl, apiKey, group, format, time.Now(), id)
	if err != nil {
		log.Errorf("Failed to update route: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("route not found: id=%d", id)
	}

	log.Infof("Route updated: id=%d", id)
	return nil
}

// DeleteRoute 删除路由及其相关的请求日志
func (s *RouteService) DeleteRoute(id int64) error {
	// 先删除该路由相关的请求日志
	_, err := s.db.Exec(`DELETE FROM request_logs WHERE route_id = ?`, id)
	if err != nil {
		log.Errorf("Failed to delete route logs: %v", err)
		return err
	}

	// 再删除路由
	query := `DELETE FROM model_routes WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Errorf("Failed to delete route: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("route not found: id=%d", id)
	}

	log.Infof("Route deleted: id=%d (with related logs)", id)
	return nil
}

// ToggleRoute 启用/禁用路由
func (s *RouteService) ToggleRoute(id int64, enabled bool) error {
	query := `UPDATE model_routes SET enabled = ?, updated_at = ? WHERE id = ?`

	_, err := s.db.Exec(query, enabled, time.Now(), id)
	if err != nil {
		log.Errorf("Failed to toggle route: %v", err)
		return err
	}

	log.Infof("Route toggled: id=%d, enabled=%v", id, enabled)
	return nil
}

// GetStats 获取统计信息
// 合并 hourly_stats（历史压缩数据）和 request_logs（实时数据）
func (s *RouteService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 路由总数
	var routeCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM model_routes WHERE enabled = 1").Scan(&routeCount)
	if err != nil {
		return nil, err
	}
	stats["route_count"] = routeCount

	// 模型总数（去重）
	var modelCount int
	err = s.db.QueryRow("SELECT COUNT(DISTINCT model) FROM model_routes WHERE enabled = 1").Scan(&modelCount)
	if err != nil {
		return nil, err
	}
	stats["model_count"] = modelCount

	// 总请求数 = hourly_stats 中的历史数据 + request_logs 中的实时数据
	var historyRequests, realtimeRequests int
	s.db.QueryRow("SELECT COALESCE(SUM(request_count), 0) FROM hourly_stats").Scan(&historyRequests)
	s.db.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&realtimeRequests)
	stats["total_requests"] = historyRequests + realtimeRequests

	// 总Token使用量 = hourly_stats + request_logs
	var historyTokens, realtimeTokens int
	s.db.QueryRow("SELECT COALESCE(SUM(total_tokens), 0) FROM hourly_stats").Scan(&historyTokens)
	s.db.QueryRow("SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs").Scan(&realtimeTokens)
	stats["total_tokens"] = historyTokens + realtimeTokens

	// 今日请求数 - 直接比较日期字符串
	var todayRequests int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM request_logs 
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
	`).Scan(&todayRequests)
	if err != nil {
		return nil, err
	}
	stats["today_requests"] = todayRequests

	// 今日Token消耗 - 直接比较日期字符串
	var todayTokens int
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs 
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
	`).Scan(&todayTokens)
	if err != nil {
		return nil, err
	}
	stats["today_tokens"] = todayTokens

	// 成功率 = (历史成功 + 实时成功) / (历史总数 + 实时总数)
	var historySuccess, realtimeSuccess int
	s.db.QueryRow("SELECT COALESCE(SUM(success_count), 0) FROM hourly_stats").Scan(&historySuccess)
	s.db.QueryRow("SELECT COUNT(*) FROM request_logs WHERE success = 1").Scan(&realtimeSuccess)

	totalRequests := historyRequests + realtimeRequests
	totalSuccess := historySuccess + realtimeSuccess
	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(totalSuccess) / float64(totalRequests) * 100
	}
	stats["success_rate"] = successRate

	log.Infof("Stats loaded: today_requests=%d, today_tokens=%d, total_requests=%d, total_tokens=%d",
		todayRequests, todayTokens, totalRequests, historyTokens+realtimeTokens)

	return stats, nil
}

// RequestLogParams 请求日志参数
type RequestLogParams struct {
	Model          string // 请求的模型名
	ProviderModel  string // 实际使用的提供商模型
	ProviderName   string // 提供商/路由名称
	RouteID        int64
	RequestTokens  int
	ResponseTokens int
	TotalTokens    int
	Success        bool
	ErrorMessage   string
	Style          string // 请求类型: openai, claude, gemini
	UserAgent      string
	RemoteIP       string
	ProxyTimeMs    int64 // 代理总耗时(毫秒)
	FirstChunkMs   int64 // 首字节时间(毫秒)
	IsStream       bool  // 是否流式请求
}

// LogRequest 记录请求日志（兼容旧版本 - 自动从 routeID 查询补全信息）
func (s *RouteService) LogRequest(model string, routeID int64, requestTokens, responseTokens, totalTokens int, success bool, errorMsg string) error {
	params := RequestLogParams{
		Model:          model,
		RouteID:        routeID,
		RequestTokens:  requestTokens,
		ResponseTokens: responseTokens,
		TotalTokens:    totalTokens,
		Success:        success,
		ErrorMessage:   errorMsg,
		IsStream:       true, // 旧版 LogRequest 主要被流式请求使用
	}

	// 自动根据 routeID 查询补全 ProviderName 和 ProviderModel
	if routeID > 0 {
		route, err := s.GetRouteByID(routeID)
		if err == nil && route != nil {
			params.ProviderName = route.Name
			params.ProviderModel = route.Model
			// 根据路由 format 推断 Style
			if route.Format != "" {
				params.Style = strings.ToLower(route.Format)
			}
		}
	}

	return s.LogRequestFull(params)
}

// LogRequestFull 记录完整的请求日志
func (s *RouteService) LogRequestFull(params RequestLogParams) error {
	// 如果 ProviderName/ProviderModel 为空且有 RouteID，尝试查询补全
	if (params.ProviderName == "" || params.ProviderModel == "") && params.RouteID > 0 {
		route, err := s.GetRouteByID(params.RouteID)
		if err == nil && route != nil {
			if params.ProviderName == "" {
				params.ProviderName = route.Name
			}
			if params.ProviderModel == "" {
				params.ProviderModel = route.Model
			}
			// 如果 Style 为空，根据路由 format 推断
			if params.Style == "" && route.Format != "" {
				params.Style = strings.ToLower(route.Format)
			}
		}
	}

	query := `INSERT INTO request_logs (
		model, provider_model, provider_name, route_id, 
		request_tokens, response_tokens, total_tokens, 
		success, error_message, style, user_agent, remote_ip,
		proxy_time_ms, first_chunk_ms, is_stream, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now', 'localtime'))`

	_, err := s.db.Exec(query,
		params.Model, params.ProviderModel, params.ProviderName, params.RouteID,
		params.RequestTokens, params.ResponseTokens, params.TotalTokens,
		params.Success, params.ErrorMessage, params.Style, params.UserAgent, params.RemoteIP,
		params.ProxyTimeMs, params.FirstChunkMs, params.IsStream,
	)
	if err != nil {
		log.Errorf("LogRequestFull error: %v", err)
	} else {
		log.Infof("LogRequest: model=%s, provider=%s, tokens=%d, success=%v, time=%dms, stream=%v",
			params.Model, params.ProviderName, params.TotalTokens, params.Success, params.ProxyTimeMs, params.IsStream)
	}
	return err
}

// GetRequestLogs 获取请求日志（支持分页和筛选）
func (s *RouteService) GetRequestLogs(page, pageSize int, filters map[string]string) ([]database.RequestLog, int, error) {
	// 构建 WHERE 子句
	var conditions []string
	var args []interface{}

	if model, ok := filters["model"]; ok && model != "" {
		conditions = append(conditions, "model = ?")
		args = append(args, model)
	}
	if providerName, ok := filters["provider_name"]; ok && providerName != "" {
		conditions = append(conditions, "provider_name = ?")
		args = append(args, providerName)
	}
	if style, ok := filters["style"]; ok && style != "" {
		conditions = append(conditions, "style = ?")
		args = append(args, style)
	}
	if success, ok := filters["success"]; ok && success != "" {
		if success == "true" {
			conditions = append(conditions, "success = 1")
		} else if success == "false" {
			conditions = append(conditions, "success = 0")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM request_logs %s", whereClause)
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, model, COALESCE(provider_model, ''), COALESCE(provider_name, ''), 
		       COALESCE(route_id, 0), request_tokens, response_tokens, total_tokens,
		       success, COALESCE(error_message, ''), COALESCE(style, ''), 
		       COALESCE(user_agent, ''), COALESCE(remote_ip, ''),
		       COALESCE(proxy_time_ms, 0), COALESCE(first_chunk_ms, 0), 
		       COALESCE(is_stream, 0), created_at
		FROM request_logs %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []database.RequestLog
	for rows.Next() {
		var l database.RequestLog
		var isStream int
		err := rows.Scan(
			&l.ID, &l.Model, &l.ProviderModel, &l.ProviderName,
			&l.RouteID, &l.RequestTokens, &l.ResponseTokens, &l.TotalTokens,
			&l.Success, &l.ErrorMessage, &l.Style,
			&l.UserAgent, &l.RemoteIP,
			&l.ProxyTimeMs, &l.FirstChunkMs, &isStream, &l.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		l.IsStream = isStream == 1
		logs = append(logs, l)
	}

	return logs, total, nil
}

// GetAvailableModels 获取所有可用的模型列表（包含重定向关键字）
func (s *RouteService) GetAvailableModels() ([]string, error) {
	query := `SELECT DISTINCT model FROM model_routes WHERE enabled = 1 ORDER BY model`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// GetAvailableModelsWithRedirect 获取所有可用的模型列表（包含重定向关键字）
func (s *RouteService) GetAvailableModelsWithRedirect(redirectKeyword string) ([]string, error) {
	query := `SELECT DISTINCT model FROM model_routes WHERE enabled = 1 ORDER BY model`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []string

	// 首先添加重定向关键字（如果配置了）
	if redirectKeyword != "" {
		models = append(models, redirectKeyword)
	}

	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// GetTodayStats 获取今日统计
func (s *RouteService) GetTodayStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 今日请求数
	var todayRequests int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM request_logs 
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
	`).Scan(&todayRequests)
	if err != nil {
		return nil, err
	}
	stats["today_requests"] = todayRequests

	// 今日Token消耗
	var todayTokens int
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs 
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
	`).Scan(&todayTokens)
	if err != nil {
		return nil, err
	}
	stats["today_tokens"] = todayTokens

	return stats, nil
}

// GetDailyStats 获取每日统计（用于热力图）
// 从 hourly_stats（历史压缩数据）和 request_logs（今天的实时数据）合并读取
func (s *RouteService) GetDailyStats(days int) ([]map[string]interface{}, error) {
	// 使用 UNION ALL 合并历史压缩数据和今天的实时数据
	query := `
		SELECT date, SUM(requests) as requests, SUM(request_tokens) as request_tokens, 
		       SUM(response_tokens) as response_tokens, SUM(total_tokens) as total_tokens
		FROM (
			-- 从 hourly_stats 获取历史压缩数据
			SELECT 
				date,
				SUM(request_count) as requests,
				SUM(request_tokens) as request_tokens,
				SUM(response_tokens) as response_tokens,
				SUM(total_tokens) as total_tokens
			FROM hourly_stats
			WHERE date >= date('now', 'localtime', ?)
			GROUP BY date
			
			UNION ALL
			
			-- 从 request_logs 获取今天的实时数据
			SELECT
				substr(created_at, 1, 10) as date,
				COUNT(*) as requests,
				COALESCE(SUM(request_tokens), 0) as request_tokens,
				COALESCE(SUM(response_tokens), 0) as response_tokens,
				COALESCE(SUM(total_tokens), 0) as total_tokens
			FROM request_logs
			WHERE substr(created_at, 1, 10) >= date('now', 'localtime', ?)
			GROUP BY substr(created_at, 1, 10)
		)
		GROUP BY date
		ORDER BY date
	`

	daysParam := fmt.Sprintf("-%d days", days)
	rows, err := s.db.Query(query, daysParam, daysParam)
	if err != nil {
		log.Errorf("GetDailyStats query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var date string
		var requests, requestTokens, responseTokens, totalTokens int
		err := rows.Scan(&date, &requests, &requestTokens, &responseTokens, &totalTokens)
		if err != nil {
			log.Errorf("GetDailyStats scan error: %v", err)
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"date":            date,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
		})
	}

	log.Infof("GetDailyStats: loaded %d days of data", len(stats))
	return stats, nil
}

// GetHourlyStats 获取今日按小时统计
func (s *RouteService) GetHourlyStats() ([]map[string]interface{}, error) {
	query := `
		SELECT
			CAST(substr(created_at, 12, 2) AS INTEGER) as hour,
			COUNT(*) as requests,
			COALESCE(SUM(request_tokens), 0) as request_tokens,
			COALESCE(SUM(response_tokens), 0) as response_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_logs
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Errorf("GetHourlyStats query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var hour, requests, requestTokens, responseTokens, totalTokens int
		err := rows.Scan(&hour, &requests, &requestTokens, &responseTokens, &totalTokens)
		if err != nil {
			log.Errorf("GetHourlyStats scan error: %v", err)
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"hour":            hour,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
		})
	}

	log.Infof("GetHourlyStats: loaded %d hours of data", len(stats))
	return stats, nil
}

// GetSecondlyStats 获取今日秒级统计
func (s *RouteService) GetSecondlyStats(minutes int) ([]map[string]interface{}, error) {
	// 获取今日全天的秒级数据
	query := `
		SELECT
			created_at as timestamp,
			COUNT(*) as requests,
			COALESCE(SUM(request_tokens), 0) as request_tokens,
			COALESCE(SUM(response_tokens), 0) as response_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_logs
		WHERE substr(created_at, 1, 10) = date('now', 'localtime')
		GROUP BY substr(created_at, 1, 19)
		ORDER BY timestamp
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Errorf("GetSecondlyStats query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var timestamp string
		var requests, requestTokens, responseTokens, totalTokens int
		err := rows.Scan(&timestamp, &requests, &requestTokens, &responseTokens, &totalTokens)
		if err != nil {
			log.Errorf("GetSecondlyStats scan error: %v", err)
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"timestamp":       timestamp,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
		})
	}

	log.Infof("GetSecondlyStats: loaded %d records for today", len(stats))
	return stats, nil
}

// ImportRouteFromFormat 从不同格式导入路由
func (s *RouteService) ImportRouteFromFormat(name, model, apiUrl, apiKey, group, targetFormat string) (string, error) {
	// 根据目标格式自动转换 API URL 和模型名
	convertedUrl, convertedModel, err := s.convertRouteFormat(apiUrl, model, targetFormat)
	if err != nil {
		return "", fmt.Errorf("格式转换失败: %v", err)
	}

	// 添加转换后的路由
	err = s.AddRoute(name+" ("+targetFormat+")", convertedModel, convertedUrl, apiKey, group, targetFormat)
	if err != nil {
		return "", fmt.Errorf("添加路由失败: %v", err)
	}

	log.Infof("Route imported and converted: %s [%s] -> %s:%s", name, model, convertedUrl, convertedModel)
	return targetFormat, nil
}

// convertRouteFormat 转换路由格式
func (s *RouteService) convertRouteFormat(apiUrl, model, targetFormat string) (string, string, error) {
	switch targetFormat {
	case "openai":
		return s.convertToOpenAI(apiUrl, model)
	case "claude":
		return s.convertToClaude(apiUrl, model)
	case "gemini":
		return s.convertToGemini(apiUrl, model)
	default:
		return apiUrl, model, nil
	}
}

// convertToOpenAI 转换为 OpenAI 格式
func (s *RouteService) convertToOpenAI(apiUrl, model string) (string, string, error) {
	// 如果已经是 OpenAI 格式，直接返回
	if isOpenAIFormat(apiUrl, model) {
		return apiUrl, model, nil
	}

	// Claude 到 OpenAI
	if isClaudeFormat(apiUrl, model) {
		return "https://api.openai.com/v1", convertClaudeModelToOpenAI(model), nil
	}

	// Gemini 到 OpenAI
	if isGeminiFormat(apiUrl, model) {
		return "https://api.openai.com/v1", convertGeminiModelToOpenAI(model), nil
	}

	// 默认转换为 OpenAI 兼容格式
	return "https://api.openai.com/v1", "gpt-3.5-turbo", nil
}

// convertToClaude 转换为 Claude 格式
func (s *RouteService) convertToClaude(apiUrl, model string) (string, string, error) {
	// 如果已经是 Claude 格式，直接返回
	if isClaudeFormat(apiUrl, model) {
		return apiUrl, model, nil
	}

	// OpenAI 到 Claude
	if isOpenAIFormat(apiUrl, model) {
		return "https://api.anthropic.com/v1", convertOpenAIModelToClaude(model), nil
	}

	// Gemini 到 Claude
	if isGeminiFormat(apiUrl, model) {
		return "https://api.anthropic.com/v1", convertGeminiModelToClaude(model), nil
	}

	// 默认转换为 Claude 兼容格式
	return "https://api.anthropic.com/v1", "claude-3-sonnet-20240229", nil
}

// convertToGemini 转换为 Gemini 格式
func (s *RouteService) convertToGemini(apiUrl, model string) (string, string, error) {
	// 如果已经是 Gemini 格式，直接返回
	if isGeminiFormat(apiUrl, model) {
		return apiUrl, model, nil
	}

	// OpenAI 到 Gemini
	if isOpenAIFormat(apiUrl, model) {
		return "https://generativelanguage.googleapis.com/v1", convertOpenAIModelToGemini(model), nil
	}

	// Claude 到 Gemini
	if isClaudeFormat(apiUrl, model) {
		return "https://generativelanguage.googleapis.com/v1", convertClaudeModelToGemini(model), nil
	}

	// 默认转换为 Gemini 兼容格式
	return "https://generativelanguage.googleapis.com/v1", "gemini-pro", nil
}

// 格式检测函数
func isOpenAIFormat(apiUrl, model string) bool {
	return strings.Contains(apiUrl, "api.openai.com") ||
		strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1-")
}

func isClaudeFormat(apiUrl, model string) bool {
	return strings.Contains(apiUrl, "api.anthropic.com") ||
		strings.HasPrefix(model, "claude-")
}

func isGeminiFormat(apiUrl, model string) bool {
	return strings.Contains(apiUrl, "generativelanguage.googleapis.com") ||
		strings.HasPrefix(model, "gemini-")
}

// 模型名转换函数
func convertClaudeModelToOpenAI(model string) string {
	modelMap := map[string]string{
		"claude-3-opus-20240229":     "gpt-4-turbo",
		"claude-3-sonnet-20240229":   "gpt-4",
		"claude-3-haiku-20240307":    "gpt-3.5-turbo",
		"claude-3-5-sonnet-20241022": "gpt-4-turbo",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "gpt-4" // 默认映射
}

func convertOpenAIModelToClaude(model string) string {
	modelMap := map[string]string{
		"gpt-4-turbo":   "claude-3-5-sonnet-20241022",
		"gpt-4":         "claude-3-sonnet-20240229",
		"gpt-3.5-turbo": "claude-3-haiku-20240307",
		"o1-preview":    "claude-3-opus-20240229",
		"o1-mini":       "claude-3-sonnet-20240229",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "claude-3-sonnet-20240229" // 默认映射
}

func convertGeminiModelToOpenAI(model string) string {
	modelMap := map[string]string{
		"gemini-1.5-pro":    "gpt-4-turbo",
		"gemini-1.5-flash":  "gpt-3.5-turbo",
		"gemini-1.0-pro":    "gpt-4",
		"gemini-pro-vision": "gpt-4-vision-preview",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "gpt-4" // 默认映射
}

func convertOpenAIModelToGemini(model string) string {
	modelMap := map[string]string{
		"gpt-4-turbo":          "gemini-1.5-pro",
		"gpt-4":                "gemini-1.0-pro",
		"gpt-3.5-turbo":        "gemini-1.5-flash",
		"gpt-4-vision-preview": "gemini-pro-vision",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "gemini-1.5-pro" // 默认映射
}

func convertGeminiModelToClaude(model string) string {
	modelMap := map[string]string{
		"gemini-1.5-pro":   "claude-3-5-sonnet-20241022",
		"gemini-1.5-flash": "claude-3-haiku-20240307",
		"gemini-1.0-pro":   "claude-3-sonnet-20240229",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "claude-3-sonnet-20240229" // 默认映射
}

func convertClaudeModelToGemini(model string) string {
	modelMap := map[string]string{
		"claude-3-opus-20240229":     "gemini-1.5-pro",
		"claude-3-sonnet-20240229":   "gemini-1.0-pro",
		"claude-3-haiku-20240307":    "gemini-1.5-flash",
		"claude-3-5-sonnet-20241022": "gemini-1.5-pro",
	}
	if mapped, exists := modelMap[model]; exists {
		return mapped
	}
	return "gemini-1.5-pro" // 默认映射
}

// ClearStats 清除统计数据
func (s *RouteService) ClearStats() error {
	query := "DELETE FROM request_logs"
	_, err := s.db.Exec(query)
	if err != nil {
		log.Errorf("Failed to clear stats: %v", err)
		return err
	}
	log.Info("All statistics data cleared")
	return nil
}

// IsRedirectModel 判断是否为重定向模型（排除在排行榜之外）
func (s *RouteService) IsRedirectModel(model string) bool {
	// 常见的重定向/代理模型标识
	redirectPatterns := []string{
		"proxy_auto",
		"proxy_",
		"redirect_",
		"forward_",
	}

	for _, pattern := range redirectPatterns {
		if strings.Contains(strings.ToLower(model), pattern) {
			return true
		}
	}
	return false
}

// GetModelRanking 获取模型使用排行（排除重定向模型）
// 合并 hourly_stats（历史压缩数据）和 request_logs（实时数据）
func (s *RouteService) GetModelRanking(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			model,
			SUM(requests) as requests,
			SUM(request_tokens) as request_tokens,
			SUM(response_tokens) as response_tokens,
			SUM(total_tokens) as total_tokens,
			CASE WHEN SUM(requests) > 0 
				THEN ROUND(SUM(success_count) * 100.0 / SUM(requests), 2) 
				ELSE 0 
			END as success_rate
		FROM (
			-- 从 hourly_stats 获取历史数据
			SELECT 
				model,
				SUM(request_count) as requests,
				SUM(request_tokens) as request_tokens,
				SUM(response_tokens) as response_tokens,
				SUM(total_tokens) as total_tokens,
				SUM(success_count) as success_count
			FROM hourly_stats
			GROUP BY model
			
			UNION ALL
			
			-- 从 request_logs 获取实时数据
			SELECT
				model,
				COUNT(*) as requests,
				COALESCE(SUM(request_tokens), 0) as request_tokens,
				COALESCE(SUM(response_tokens), 0) as response_tokens,
				COALESCE(SUM(total_tokens), 0) as total_tokens,
				SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as success_count
			FROM request_logs
			GROUP BY model
		)
		GROUP BY model
		ORDER BY total_tokens DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	rank := 1
	for rows.Next() {
		var model string
		var requests, requestTokens, responseTokens, totalTokens int
		var successRate float64
		err := rows.Scan(&model, &requests, &requestTokens, &responseTokens, &totalTokens, &successRate)
		if err != nil {
			return nil, err
		}

		// 跳过重定向模型
		if strings.Contains(strings.ToLower(model), "proxy_auto") ||
			strings.Contains(strings.ToLower(model), "proxy_") ||
			strings.Contains(strings.ToLower(model), "redirect_") ||
			strings.Contains(strings.ToLower(model), "forward_") {
			continue
		}

		ranking = append(ranking, map[string]interface{}{
			"rank":            rank,
			"model":           model,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
			"success_rate":    successRate,
		})
		rank++

		// 如果达到限制数量，停止添加
		if rank > limit {
			break
		}
	}

	return ranking, nil
}

// CompressDatabase 压缩数据库
// 1. 将 request_logs 中今天之前的数据按小时聚合到 hourly_stats
// 2. 删除已聚合的 request_logs 数据
// 3. 更新 usage_summary 中的周/年/总用量
// 4. 删除超过 366 天的 hourly_stats 数据
func (s *RouteService) CompressDatabase() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. 统计压缩前的数据量
	var beforeCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&beforeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count request_logs: %v", err)
	}
	result["before_count"] = beforeCount

	// 2. 将今天之前的数据按小时聚合
	// 先创建临时表存储聚合结果
	_, err = tx.Exec(`
		CREATE TEMP TABLE temp_hourly AS
		SELECT 
			substr(created_at, 1, 10) as date,
			CAST(substr(created_at, 12, 2) AS INTEGER) as hour,
			model,
			COUNT(*) as request_count,
			COALESCE(SUM(request_tokens), 0) as request_tokens,
			COALESCE(SUM(response_tokens), 0) as response_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as fail_count
		FROM request_logs
		WHERE substr(created_at, 1, 10) < date('now', 'localtime')
		GROUP BY substr(created_at, 1, 10), CAST(substr(created_at, 12, 2) AS INTEGER), model
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp table: %v", err)
	}

	// 3. 合并到 hourly_stats（累加已存在的记录，插入新记录）
	_, err = tx.Exec(`
		INSERT INTO hourly_stats (date, hour, model, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count)
		SELECT date, hour, model, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count
		FROM temp_hourly
		WHERE NOT EXISTS (
			SELECT 1 FROM hourly_stats h 
			WHERE h.date = temp_hourly.date AND h.hour = temp_hourly.hour AND h.model = temp_hourly.model
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new hourly stats: %v", err)
	}

	_, err = tx.Exec(`
		UPDATE hourly_stats SET
			request_count = hourly_stats.request_count + (SELECT request_count FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model),
			request_tokens = hourly_stats.request_tokens + (SELECT request_tokens FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model),
			response_tokens = hourly_stats.response_tokens + (SELECT response_tokens FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model),
			total_tokens = hourly_stats.total_tokens + (SELECT total_tokens FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model),
			success_count = hourly_stats.success_count + (SELECT success_count FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model),
			fail_count = hourly_stats.fail_count + (SELECT fail_count FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model)
		WHERE EXISTS (SELECT 1 FROM temp_hourly t WHERE t.date = hourly_stats.date AND t.hour = hourly_stats.hour AND t.model = hourly_stats.model)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to update hourly stats: %v", err)
	}

	// 4. 删除临时表
	_, err = tx.Exec("DROP TABLE temp_hourly")
	if err != nil {
		return nil, fmt.Errorf("failed to drop temp table: %v", err)
	}

	// 5. 删除今天之前的原始请求日志
	deleteResult, err := tx.Exec(`
		DELETE FROM request_logs 
		WHERE substr(created_at, 1, 10) < date('now', 'localtime')
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old logs: %v", err)
	}
	deletedLogs, _ := deleteResult.RowsAffected()
	result["deleted_logs"] = deletedLogs

	// 6. 更新 usage_summary
	// 周用量
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO usage_summary (period_type, period_key, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count, updated_at)
		SELECT 
			'week' as period_type,
			strftime('%Y-W%W', date) as period_key,
			SUM(request_count), SUM(request_tokens), SUM(response_tokens), SUM(total_tokens), SUM(success_count), SUM(fail_count),
			datetime('now', 'localtime')
		FROM hourly_stats
		GROUP BY strftime('%Y-W%W', date)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to update week summary: %v", err)
	}

	// 年用量
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO usage_summary (period_type, period_key, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count, updated_at)
		SELECT 
			'year' as period_type,
			strftime('%Y', date) as period_key,
			SUM(request_count), SUM(request_tokens), SUM(response_tokens), SUM(total_tokens), SUM(success_count), SUM(fail_count),
			datetime('now', 'localtime')
		FROM hourly_stats
		GROUP BY strftime('%Y', date)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to update year summary: %v", err)
	}

	// 总用量
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO usage_summary (period_type, period_key, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count, updated_at)
		SELECT 
			'total', 'total',
			SUM(request_count), SUM(request_tokens), SUM(response_tokens), SUM(total_tokens), SUM(success_count), SUM(fail_count),
			datetime('now', 'localtime')
		FROM hourly_stats
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to update total summary: %v", err)
	}

	// 7. 删除超过 366 天的 hourly_stats 数据
	deleteStatsResult, err := tx.Exec(`
		DELETE FROM hourly_stats 
		WHERE date < date('now', 'localtime', '-366 days')
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old hourly stats: %v", err)
	}
	deletedStats, _ := deleteStatsResult.RowsAffected()
	result["deleted_hourly_stats"] = deletedStats

	// 8. 统计压缩后的数据量
	var afterCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&afterCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count request_logs after: %v", err)
	}
	result["after_count"] = afterCount

	var hourlyStatsCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM hourly_stats").Scan(&hourlyStatsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count hourly_stats: %v", err)
	}
	result["hourly_stats_count"] = hourlyStatsCount

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// 9. 执行 VACUUM 压缩数据库文件
	_, err = s.db.Exec("VACUUM")
	if err != nil {
		log.Warnf("VACUUM failed: %v", err)
	}

	log.Infof("Database compressed: before=%d, after=%d, deleted_logs=%d, deleted_stats=%d, hourly_stats=%d",
		beforeCount, afterCount, deletedLogs, deletedStats, hourlyStatsCount)

	return result, nil
}

// GetUsageSummary 获取用量汇总
func (s *RouteService) GetUsageSummary() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取周用量
	weekRows, err := s.db.Query(`
		SELECT period_key, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count
		FROM usage_summary 
		WHERE period_type = 'week' 
		ORDER BY period_key DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer weekRows.Close()

	var weekStats []map[string]interface{}
	for weekRows.Next() {
		var periodKey string
		var requestCount, requestTokens, responseTokens, totalTokens, successCount, failCount int64
		err := weekRows.Scan(&periodKey, &requestCount, &requestTokens, &responseTokens, &totalTokens, &successCount, &failCount)
		if err != nil {
			continue
		}
		weekStats = append(weekStats, map[string]interface{}{
			"period":          periodKey,
			"request_count":   requestCount,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
			"success_count":   successCount,
			"fail_count":      failCount,
		})
	}
	result["week_stats"] = weekStats

	// 获取年用量
	yearRows, err := s.db.Query(`
		SELECT period_key, request_count, request_tokens, response_tokens, total_tokens, success_count, fail_count
		FROM usage_summary 
		WHERE period_type = 'year' 
		ORDER BY period_key DESC
	`)
	if err != nil {
		return nil, err
	}
	defer yearRows.Close()

	var yearStats []map[string]interface{}
	for yearRows.Next() {
		var periodKey string
		var requestCount, requestTokens, responseTokens, totalTokens, successCount, failCount int64
		err := yearRows.Scan(&periodKey, &requestCount, &requestTokens, &responseTokens, &totalTokens, &successCount, &failCount)
		if err != nil {
			continue
		}
		yearStats = append(yearStats, map[string]interface{}{
			"period":          periodKey,
			"request_count":   requestCount,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
			"success_count":   successCount,
			"fail_count":      failCount,
		})
	}
	result["year_stats"] = yearStats

	// 获取总用量
	var totalRequestCount, totalRequestTokens, totalResponseTokens, totalTotalTokens, totalSuccessCount, totalFailCount int64
	err = s.db.QueryRow(`
		SELECT COALESCE(request_count, 0), COALESCE(request_tokens, 0), COALESCE(response_tokens, 0), 
		       COALESCE(total_tokens, 0), COALESCE(success_count, 0), COALESCE(fail_count, 0)
		FROM usage_summary 
		WHERE period_type = 'total' AND period_key = 'total'
	`).Scan(&totalRequestCount, &totalRequestTokens, &totalResponseTokens, &totalTotalTokens, &totalSuccessCount, &totalFailCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result["total_stats"] = map[string]interface{}{
		"request_count":   totalRequestCount,
		"request_tokens":  totalRequestTokens,
		"response_tokens": totalResponseTokens,
		"total_tokens":    totalTotalTokens,
		"success_count":   totalSuccessCount,
		"fail_count":      totalFailCount,
	}

	return result, nil
}

// RouteHealthInfo represents health information for a single route
type RouteHealthInfo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Model         string `json:"model"`
	StatusHistory []bool `json:"status_history"` // Last N requests, true=success, index 0 is oldest
	SuccessRate   float64 `json:"success_rate"`
	TotalRequests int    `json:"total_requests"`
}

// GroupHealthInfo represents health information for a group of routes
type GroupHealthInfo struct {
	Group       string            `json:"group"`
	Routes      []RouteHealthInfo `json:"routes"`
	SuccessRate float64           `json:"success_rate"`
	RouteCount  int               `json:"route_count"`
}

// GetHealthStatus returns health status for all routes, grouped by their group field
// historyCount specifies how many recent requests to include in status_history (e.g., 50)
func (s *RouteService) GetHealthStatus(historyCount int) ([]GroupHealthInfo, error) {
	// Step 1: Get all enabled routes
	query := `SELECT id, name, model, "group" FROM model_routes WHERE enabled = 1 ORDER BY "group", name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type routeBasic struct {
		ID    int64
		Name  string
		Model string
		Group string
	}
	var routes []routeBasic
	for rows.Next() {
		var r routeBasic
		if err := rows.Scan(&r.ID, &r.Name, &r.Model, &r.Group); err != nil {
			return nil, err
		}
		routes = append(routes, r)
	}

	// Step 2: For each route, get last N requests' success status
	// Group routes by group
	groupMap := make(map[string][]RouteHealthInfo)
	
	for _, r := range routes {
		// Query last N requests for this route
		histQuery := `SELECT success FROM request_logs WHERE route_id = ? ORDER BY id DESC LIMIT ?`
		histRows, err := s.db.Query(histQuery, r.ID, historyCount)
		if err != nil {
			log.Warnf("GetHealthStatus: failed to get history for route %d: %v", r.ID, err)
			continue
		}

		var statusHistory []bool
		var successCount int
		for histRows.Next() {
			var success int
			if err := histRows.Scan(&success); err != nil {
				histRows.Close()
				continue
			}
			// Prepend since we query DESC order (newest first), we want oldest first
			statusHistory = append([]bool{success == 1}, statusHistory...)
			if success == 1 {
				successCount++
			}
		}
		histRows.Close()

		totalReqs := len(statusHistory)
		var successRate float64
		if totalReqs > 0 {
			successRate = float64(successCount) / float64(totalReqs) * 100
		}

		routeHealth := RouteHealthInfo{
			ID:            r.ID,
			Name:          r.Name,
			Model:         r.Model,
			StatusHistory: statusHistory,
			SuccessRate:   successRate,
			TotalRequests: totalReqs,
		}

		groupName := r.Group
		if groupName == "" {
			groupName = "default"
		}
		groupMap[groupName] = append(groupMap[groupName], routeHealth)
	}

	// Step 3: Calculate group-level stats
	var result []GroupHealthInfo
	for groupName, routeList := range groupMap {
		var totalSuccess, totalReqs int
		for _, rh := range routeList {
			for _, s := range rh.StatusHistory {
				totalReqs++
				if s {
					totalSuccess++
				}
			}
		}

		var groupSuccessRate float64
		if totalReqs > 0 {
			groupSuccessRate = float64(totalSuccess) / float64(totalReqs) * 100
		}

		result = append(result, GroupHealthInfo{
			Group:       groupName,
			Routes:      routeList,
			SuccessRate: groupSuccessRate,
			RouteCount:  len(routeList),
		})
	}

	// Sort by group name for consistent ordering
	// "default" group should come last
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			swapNeeded := false
			if result[i].Group == "default" {
				swapNeeded = true
			} else if result[j].Group != "default" && result[i].Group > result[j].Group {
				swapNeeded = true
			}
			if swapNeeded {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}
