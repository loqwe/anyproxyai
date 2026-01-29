package services

import (
	"fmt"
	"os"
	"os/exec"

	"openai-router-go/internal/config"
	"openai-router-go/internal/service"
	"openai-router-go/internal/system"

	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// RouteInfo 路由信息结构体（用于前端）
type RouteInfo struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Model   string `json:"model"`
	APIUrl  string `json:"api_url"`
	APIKey  string `json:"api_key"`
	Group   string `json:"group"`
	Format  string `json:"format"`
	Enabled bool   `json:"enabled"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// StatsInfo 统计信息结构体
type StatsInfo struct {
	RouteCount    int     `json:"route_count"`
	ModelCount    int     `json:"model_count"`
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TodayRequests int64   `json:"today_requests"`
	TodayTokens   int64   `json:"today_tokens"`
	SuccessRate   float64 `json:"success_rate"`
}

// ConfigInfo 配置信息结构体
type ConfigInfo struct {
	LocalApiKey           string `json:"localApiKey"`
	OpenaiEndpoint        string `json:"openaiEndpoint"`
	RedirectEnabled       bool   `json:"redirectEnabled"`
	RedirectKeyword       string `json:"redirectKeyword"`
	RedirectTargetModel   string `json:"redirectTargetModel"`
	RedirectTargetName    string `json:"redirectTargetName"`
	RedirectTargetRouteID int64  `json:"redirectTargetRouteId"`
	MinimizeToTray        bool   `json:"minimizeToTray"`
	AutoStart             bool   `json:"autoStart"`
	EnableFileLog         bool   `json:"enableFileLog"`
	Port                  int    `json:"port"`
}

// AppSettingsInfo 应用设置结构体
type AppSettingsInfo struct {
	MinimizeToTray   bool `json:"minimizeToTray"`
	AutoStart        bool `json:"autoStart"`
	AutoStartEnabled bool `json:"autoStartEnabled"`
}

// DailyStatsInfo 每日统计结构体
type DailyStatsInfo struct {
	Date           string `json:"date"`
	Requests       int64  `json:"requests"`
	RequestTokens  int64  `json:"request_tokens"`
	ResponseTokens int64  `json:"response_tokens"`
	TotalTokens    int64  `json:"total_tokens"`
}

// HourlyStatsInfo 小时统计结构体
type HourlyStatsInfo struct {
	Hour        int   `json:"hour"`
	Requests    int64 `json:"requests"`
	TotalTokens int64 `json:"total_tokens"`
}

// ModelRankingInfo 模型排行结构体
type ModelRankingInfo struct {
	Rank           int     `json:"rank"`
	Model          string  `json:"model"`
	Requests       int64   `json:"requests"`
	RequestTokens  int64   `json:"request_tokens"`
	ResponseTokens int64   `json:"response_tokens"`
	TotalTokens    int64   `json:"total_tokens"`
	SuccessRate    float64 `json:"success_rate"`
}

// AppService 结构体用于 Wails v3 服务绑定
type AppService struct {
	App          *application.App
	RouteService *service.RouteService
	ProxyService *service.ProxyService
	Config       *config.Config
	AutoStart    *system.AutoStart
}

// NewAppService 创建新的 AppService 实例
func NewAppService(routeService *service.RouteService, proxyService *service.ProxyService, cfg *config.Config, autoStart *system.AutoStart) *AppService {
	return &AppService{
		RouteService: routeService,
		ProxyService: proxyService,
		Config:       cfg,
		AutoStart:    autoStart,
	}
}

// SetApp 设置应用引用
func (a *AppService) SetApp(app *application.App) {
	a.App = app
}

// GetLanguage 获取当前语言设置
func (a *AppService) GetLanguage() string {
	return a.Config.Language
}

// SetLanguage 设置语言
func (a *AppService) SetLanguage(lang string) error {
	a.Config.Language = lang
	return a.Config.Save()
}

// GetRoutes 获取所有路由
func (a *AppService) GetRoutes() ([]RouteInfo, error) {
	routes, err := a.RouteService.GetAllRoutes()
	if err != nil {
		return nil, err
	}

	result := make([]RouteInfo, len(routes))
	for i, route := range routes {
		result[i] = RouteInfo{
			ID:      route.ID,
			Name:    route.Name,
			Model:   route.Model,
			APIUrl:  route.APIUrl,
			APIKey:  route.APIKey,
			Group:   route.Group,
			Format:  route.Format,
			Enabled: route.Enabled,
			Created: route.CreatedAt.Format("2006-01-02 15:04:05"),
			Updated: route.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

// AddRoute 添加路由
func (a *AppService) AddRoute(name, model, apiUrl, apiKey, group, format string) error {
	return a.RouteService.AddRoute(name, model, apiUrl, apiKey, group, format)
}

// UpdateRoute 更新路由
func (a *AppService) UpdateRoute(id int64, name, model, apiUrl, apiKey, group, format string) error {
	return a.RouteService.UpdateRoute(id, name, model, apiUrl, apiKey, group, format)
}

// DeleteRoute 删除路由
func (a *AppService) DeleteRoute(id int64) error {
	return a.RouteService.DeleteRoute(id)
}

// ToggleRoute 启用/禁用路由
func (a *AppService) ToggleRoute(id int64, enabled bool) error {
	return a.RouteService.ToggleRoute(id, enabled)
}

// GetStats 获取统计信息
func (a *AppService) GetStats() (StatsInfo, error) {
	stats, err := a.RouteService.GetStats()
	if err != nil {
		return StatsInfo{}, err
	}
	// 从 map 转换为结构体
	result := StatsInfo{}
	if v, ok := stats["route_count"].(int); ok {
		result.RouteCount = v
	}
	if v, ok := stats["model_count"].(int); ok {
		result.ModelCount = v
	}
	// RouteService.GetStats 返回的是 int 类型，不是 int64
	if v, ok := stats["total_requests"].(int); ok {
		result.TotalRequests = int64(v)
	}
	if v, ok := stats["total_tokens"].(int); ok {
		result.TotalTokens = int64(v)
	}
	if v, ok := stats["today_requests"].(int); ok {
		result.TodayRequests = int64(v)
	}
	if v, ok := stats["today_tokens"].(int); ok {
		result.TodayTokens = int64(v)
	}
	if v, ok := stats["success_rate"].(float64); ok {
		result.SuccessRate = v
	}
	return result, nil
}

// GetDailyStats 获取每日统计（用于热力图）
func (a *AppService) GetDailyStats(days int) ([]map[string]interface{}, error) {
	return a.RouteService.GetDailyStats(days)
}

// GetHourlyStats 获取今日按小时统计（用于折线图）
func (a *AppService) GetHourlyStats() ([]map[string]interface{}, error) {
	return a.RouteService.GetHourlyStats()
}

// GetSecondlyStats 获取秒级统计（用于实时折线图）
func (a *AppService) GetSecondlyStats(minutes int) ([]map[string]interface{}, error) {
	return a.RouteService.GetSecondlyStats(minutes)
}

// GetModelRanking 获取模型使用排行
func (a *AppService) GetModelRanking(limit int) ([]map[string]interface{}, error) {
	return a.RouteService.GetModelRanking(limit)
}

// GetConfig 获取配置
func (a *AppService) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"localApiKey":           a.Config.LocalAPIKey,
		"openaiEndpoint":        fmt.Sprintf("http://%s:%d", a.Config.Host, a.Config.Port),
		"redirectEnabled":       a.Config.RedirectEnabled,
		"redirectKeyword":       a.Config.RedirectKeyword,
		"redirectTargetModel":   a.Config.RedirectTargetModel,
		"redirectTargetName":    a.Config.RedirectTargetName,
		"redirectTargetRouteId": a.Config.RedirectTargetRouteID,
		"minimizeToTray":        a.Config.MinimizeToTray,
		"autoStart":             a.Config.AutoStart,
		"enableFileLog":         a.Config.EnableFileLog,
		"fallbackEnabled":       a.Config.FallbackEnabled,
		"tracesEnabled":         a.Config.TracesEnabled,
		"tracesRetentionDays":   a.Config.TracesRetentionDays,
		"port":                  a.Config.Port,
	}
}

// UpdateConfig 更新配置
func (a *AppService) UpdateConfig(redirectEnabled bool, redirectKeyword, redirectTargetModel string, redirectTargetRouteId int64) error {
	a.Config.RedirectEnabled = redirectEnabled
	a.Config.RedirectKeyword = redirectKeyword
	a.Config.RedirectTargetModel = redirectTargetModel
	a.Config.RedirectTargetRouteID = redirectTargetRouteId
	return a.Config.Save()
}

// UpdatePort 更新端口配置
func (a *AppService) UpdatePort(port int) error {
	a.Config.Port = port
	return a.Config.Save()
}

// UpdateLocalApiKey 更新本地 API Key
func (a *AppService) UpdateLocalApiKey(newApiKey string) error {
	a.Config.LocalAPIKey = newApiKey
	return a.Config.Save()
}

// FetchRemoteModels 获取远程模型列表
func (a *AppService) FetchRemoteModels(apiUrl, apiKey string) ([]string, error) {
	return a.ProxyService.FetchRemoteModels(apiUrl, apiKey)
}

// ImportRouteFromFormat 从不同格式导入路由
func (a *AppService) ImportRouteFromFormat(name, model, apiUrl, apiKey, group, targetFormat string) (string, error) {
	return a.RouteService.ImportRouteFromFormat(name, model, apiUrl, apiKey, group, targetFormat)
}

// GetAppSettings 获取应用设置
func (a *AppService) GetAppSettings() map[string]interface{} {
	autoStartEnabled := false
	if a.AutoStart != nil {
		autoStartEnabled = a.AutoStart.IsAutoStartEnabled()
	}

	return map[string]interface{}{
		"minimizeToTray":   a.Config.MinimizeToTray,
		"autoStart":        a.Config.AutoStart,
		"autoStartEnabled": autoStartEnabled,
	}
}

// SetMinimizeToTray 设置关闭时最小化到托盘
func (a *AppService) SetMinimizeToTray(enabled bool) error {
	log.Infof("Setting minimize to tray: %v", enabled)
	a.Config.MinimizeToTray = enabled

	if err := a.Config.Save(); err != nil {
		log.Errorf("Failed to save config: %v", err)
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Info("Minimize to tray setting updated successfully")
	return nil
}

// SetAutoStart 设置开机自启动
func (a *AppService) SetAutoStart(enabled bool) error {
	log.Infof("Setting auto-start: %v", enabled)

	if a.AutoStart == nil {
		log.Error("Auto-start manager not initialized")
		return fmt.Errorf("auto-start manager not initialized")
	}

	if enabled {
		if err := a.AutoStart.EnableAutoStart(); err != nil {
			log.Errorf("Failed to enable auto-start: %v", err)
			return fmt.Errorf("failed to enable auto-start: %v", err)
		}
	} else {
		if err := a.AutoStart.DisableAutoStart(); err != nil {
			log.Errorf("Failed to disable auto-start: %v", err)
			return fmt.Errorf("failed to disable auto-start: %v", err)
		}
	}

	a.Config.AutoStart = enabled
	if err := a.Config.Save(); err != nil {
		log.Errorf("Failed to save config: %v", err)
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Info("Auto-start setting updated successfully")
	return nil
}

// ClearStats 清除统计数据
func (a *AppService) ClearStats() error {
	err := a.RouteService.ClearStats()
	if err != nil {
		return fmt.Errorf("failed to clear statistics: %v", err)
	}

	log.Info("Statistics cleared successfully")
	return nil
}

// LogFile 全局日志文件句柄（由 main 包设置）
var LogFile interface {
	Close() error
}

// SetEnableFileLog 设置是否启用文件日志
func (a *AppService) SetEnableFileLog(enabled bool) error {
	log.Infof("Setting enable file log: %v", enabled)
	a.Config.EnableFileLog = enabled

	if err := a.Config.Save(); err != nil {
		log.Errorf("Failed to save config: %v", err)
		return fmt.Errorf("failed to save config: %v", err)
	}

	// 注意：实际的日志文件启用/禁用逻辑在 main 包中处理
	log.Info("File log setting updated successfully")
	return nil
}

// SetFallbackEnabled 设置是否启用故障转移
func (a *AppService) SetFallbackEnabled(enabled bool) error {
	log.Infof("Setting fallback enabled: %v", enabled)
	a.Config.FallbackEnabled = enabled

	if err := a.Config.Save(); err != nil {
		log.Errorf("Failed to save config: %v", err)
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Info("Fallback setting updated successfully")
	return nil
}

// RestartApp 重启应用
func (a *AppService) RestartApp() error {
	log.Info("Restarting application...")

	// 获取当前可执行文件路径
	executable, err := os.Executable()
	if err != nil {
		log.Errorf("Failed to get executable path: %v", err)
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// 启动新进程
	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to start new process: %v", err)
		return fmt.Errorf("failed to start new process: %v", err)
	}

	log.Info("New process started, quitting current process...")

	// 退出当前应用
	if a.App != nil {
		a.App.Quit()
	}
	return nil
}

// CompressDatabase 压缩数据库
// 将历史请求日志按小时聚合，删除超过366天的数据，更新用量汇总
func (a *AppService) CompressDatabase() (map[string]interface{}, error) {
	log.Info("Starting database compression...")
	result, err := a.RouteService.CompressDatabase()
	if err != nil {
		log.Errorf("Database compression failed: %v", err)
		return nil, fmt.Errorf("database compression failed: %v", err)
	}
	log.Info("Database compression completed successfully")
	return result, nil
}

// GetUsageSummary 获取用量汇总（周/年/总用量）
func (a *AppService) GetUsageSummary() (map[string]interface{}, error) {
	return a.RouteService.GetUsageSummary()
}

// RequestLogsResult 请求日志查询结果
type RequestLogsResult struct {
	Data     []map[string]interface{} `json:"data"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// GetRequestLogs 获取请求日志（支持分页和筛选）
func (a *AppService) GetRequestLogs(page, pageSize int, model, style, success string) (RequestLogsResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建筛选参数
	filters := make(map[string]string)
	if model != "" {
		filters["model"] = model
	}
	if style != "" {
		filters["style"] = style
	}
	if success != "" {
		filters["success"] = success
	}

	logs, total, err := a.RouteService.GetRequestLogs(page, pageSize, filters)
	if err != nil {
		return RequestLogsResult{}, err
	}

	// 转换为 map 切片
	data := make([]map[string]interface{}, len(logs))
	for i, l := range logs {
		data[i] = map[string]interface{}{
			"id":              l.ID,
			"model":           l.Model,
			"provider_model":  l.ProviderModel,
			"provider_name":   l.ProviderName,
			"route_id":        l.RouteID,
			"request_tokens":  l.RequestTokens,
			"response_tokens": l.ResponseTokens,
			"total_tokens":    l.TotalTokens,
			"success":         l.Success,
			"error_message":   l.ErrorMessage,
			"style":           l.Style,
			"user_agent":      l.UserAgent,
			"remote_ip":       l.RemoteIP,
			"proxy_time_ms":   l.ProxyTimeMs,
			"first_chunk_ms":  l.FirstChunkMs,
			"is_stream":       l.IsStream,
			"created_at":      l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return RequestLogsResult{
		Data:     data,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RouteHealthInfo represents health information for a single route (frontend binding)
type RouteHealthInfo struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Model         string  `json:"model"`
	StatusHistory []bool  `json:"status_history"` // Last N requests, true=success, index 0 is oldest
	SuccessRate   float64 `json:"success_rate"`
	TotalRequests int     `json:"total_requests"`
}

// GroupHealthInfo represents health information for a group of routes (frontend binding)
type GroupHealthInfo struct {
	Group       string            `json:"group"`
	Routes      []RouteHealthInfo `json:"routes"`
	SuccessRate float64           `json:"success_rate"`
	RouteCount  int               `json:"route_count"`
}

// GetHealthStatus 获取所有路由的健康状态（按分组）
func (a *AppService) GetHealthStatus() ([]GroupHealthInfo, error) {
	const historyCount = 50 // Display last 50 requests per route
	
	results, err := a.RouteService.GetHealthStatus(historyCount)
	if err != nil {
		return nil, err
	}

	// Convert from service types to AppService types
	var groups []GroupHealthInfo
	for _, g := range results {
		var routes []RouteHealthInfo
		for _, r := range g.Routes {
			routes = append(routes, RouteHealthInfo{
				ID:            r.ID,
				Name:          r.Name,
				Model:         r.Model,
				StatusHistory: r.StatusHistory,
				SuccessRate:   r.SuccessRate,
				TotalRequests: r.TotalRequests,
			})
		}
		groups = append(groups, GroupHealthInfo{
			Group:       g.Group,
			Routes:      routes,
			SuccessRate: g.SuccessRate,
			RouteCount:  g.RouteCount,
		})
	}

	return groups, nil
}

// ================== Traces 相关方法 ==================

// TraceSessionInfo 会话信息结构体（前端）
type TraceSessionInfo struct {
	SessionID     string `json:"session_id"`
	RemoteIP      string `json:"remote_ip"`
	TraceCount    int    `json:"trace_count"`
	FirstTraceAt  string `json:"first_trace_at"`
	LastTraceAt   string `json:"last_trace_at"`
}

// TraceDetailInfo 对话详情结构体（前端）
type TraceDetailInfo struct {
	ID              int64  `json:"id"`
	SessionID       string `json:"session_id"`
	RemoteIP        string `json:"remote_ip"`
	Model           string `json:"model"`
	ProviderModel   string `json:"provider_model"`
	ProviderName    string `json:"provider_name"`
	RequestContent  string `json:"request_content"`
	ResponseContent string `json:"response_content"`
	RequestTokens   int    `json:"request_tokens"`
	ResponseTokens  int    `json:"response_tokens"`
	TotalTokens     int    `json:"total_tokens"`
	Success         bool   `json:"success"`
	ErrorMessage    string `json:"error_message"`
	Style           string `json:"style"`
	IsStream        bool   `json:"is_stream"`
	ProxyTimeMs     int64  `json:"proxy_time_ms"`
	CreatedAt       string `json:"created_at"`
}

// GetTracesEnabled 获取 Traces 功能是否启用
func (a *AppService) GetTracesEnabled() bool {
	return a.Config.TracesEnabled
}

// SetTracesEnabled 设置 Traces 功能启用/禁用
func (a *AppService) SetTracesEnabled(enabled bool) error {
	a.Config.TracesEnabled = enabled
	return a.Config.Save()
}

// GetTracesRetentionDays 获取 Traces 保留天数
func (a *AppService) GetTracesRetentionDays() int {
	return a.Config.TracesRetentionDays
}

// SetTracesRetentionDays 设置 Traces 保留天数
func (a *AppService) SetTracesRetentionDays(days int) error {
	if days < 1 {
		days = 1
	}
	a.Config.TracesRetentionDays = days
	return a.Config.Save()
}

// GetTracesSessionTimeout 获取 Traces 会话超时（分钟）
func (a *AppService) GetTracesSessionTimeout() int {
	return a.Config.TracesSessionTimeout
}

// SetTracesSessionTimeout 设置 Traces 会话超时（分钟）
func (a *AppService) SetTracesSessionTimeout(minutes int) error {
	if minutes < 1 {
		minutes = 1
	}
	a.Config.TracesSessionTimeout = minutes
	return a.Config.Save()
}

// TraceSessionsResult 会话列表结果
type TraceSessionsResult struct {
	Sessions []TraceSessionInfo `json:"sessions"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// GetTraceSessions 获取会话列表（分页）
func (a *AppService) GetTraceSessions(page, pageSize int) (TraceSessionsResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	sessions, total, err := a.RouteService.GetTraceSessions(page, pageSize)
	if err != nil {
		return TraceSessionsResult{}, err
	}

	result := make([]TraceSessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = TraceSessionInfo{
			SessionID:    s.SessionID,
			RemoteIP:     s.RemoteIP,
			TraceCount:   s.MessageCount,
			FirstTraceAt: s.FirstTime.Format("2006-01-02 15:04:05"),
			LastTraceAt:  s.LastTime.Format("2006-01-02 15:04:05"),
		}
	}
	return TraceSessionsResult{
		Sessions: result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetTracesBySession 获取指定会话的对话详情
func (a *AppService) GetTracesBySession(sessionID string) ([]TraceDetailInfo, error) {
	traces, err := a.RouteService.GetTracesBySession(sessionID)
	if err != nil {
		return nil, err
	}

	result := make([]TraceDetailInfo, len(traces))
	for i, t := range traces {
		result[i] = TraceDetailInfo{
			ID:              t.ID,
			SessionID:       t.SessionID,
			RemoteIP:        t.RemoteIP,
			Model:           t.Model,
			ProviderModel:   t.ProviderModel,
			ProviderName:    t.ProviderName,
			RequestContent:  t.RequestContent,
			ResponseContent: t.ResponseContent,
			RequestTokens:   t.RequestTokens,
			ResponseTokens:  t.ResponseTokens,
			TotalTokens:     t.TotalTokens,
			Success:         t.Success,
			ErrorMessage:    t.ErrorMessage,
			Style:           t.Style,
			IsStream:        t.IsStream,
			ProxyTimeMs:     t.ProxyTimeMs,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

// ClearOldTraces 清除过期的 Trace 记录
func (a *AppService) ClearOldTraces(beforeDays int) (int64, error) {
	if beforeDays < 0 {
		beforeDays = a.Config.TracesRetentionDays
	}
	return a.RouteService.ClearTraces(beforeDays)
}

// ClearAllTraces 清除所有 Trace 记录
func (a *AppService) ClearAllTraces() (int64, error) {
	return a.RouteService.ClearAllTraces()
}

// GetTracesCount 获取 Trace 记录总数
func (a *AppService) GetTracesCount() (int64, error) {
	return a.RouteService.GetTracesCount()
}

// AllTracesResult 所有 trace 列表结果
type AllTracesResult struct {
	Traces   []TraceDetailInfo `json:"traces"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// GetAllTraces 获取所有 trace 记录（按时间倒序，分页）
func (a *AppService) GetAllTraces(page, pageSize int) (AllTracesResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	traces, total, err := a.RouteService.GetAllTraces(page, pageSize)
	if err != nil {
		return AllTracesResult{}, err
	}

	result := make([]TraceDetailInfo, len(traces))
	for i, t := range traces {
		result[i] = TraceDetailInfo{
			ID:              t.ID,
			SessionID:       t.SessionID,
			RemoteIP:        t.RemoteIP,
			Model:           t.Model,
			ProviderModel:   t.ProviderModel,
			ProviderName:    t.ProviderName,
			RequestContent:  t.RequestContent,
			ResponseContent: t.ResponseContent,
			RequestTokens:   t.RequestTokens,
			ResponseTokens:  t.ResponseTokens,
			TotalTokens:     t.TotalTokens,
			Success:         t.Success,
			ErrorMessage:    t.ErrorMessage,
			Style:           t.Style,
			IsStream:        t.IsStream,
			ProxyTimeMs:     t.ProxyTimeMs,
			CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return AllTracesResult{
		Traces:   result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
