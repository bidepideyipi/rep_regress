package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rtp-processor/config"
	"github.com/rtp-processor/models"
)

// ClickHouseWriter ClickHouse写入器（使用HTTP接口）
type ClickHouseWriter struct {
	baseURL    string
	httpClient *http.Client
	cfg        *config.Config
	tableName  string
	flushCount int64
	errorCount int64
}

// NewClickHouseWriter 创建ClickHouse写入器（使用HTTP）
func NewClickHouseWriter(cfg *config.Config) (*ClickHouseWriter, error) {
	chConfig := cfg.ClickHouse
	if chConfig.Host == "" || chConfig.Port == 0 || chConfig.Username == "" || chConfig.Database == "" {
		return nil, fmt.Errorf("ClickHouse配置不完整")
	}

	log.Printf("[ClickHouse] 初始化写入器(HTTP): %s:%d/%s",
		chConfig.Host, 8123, chConfig.Database)

	baseURL := fmt.Sprintf("http://%s:8123", chConfig.Host)

	writer := &ClickHouseWriter{
		baseURL: baseURL,
		cfg:     cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		tableName: "game_log_detail_local",
	}

	// 测试连接
	if err := writer.ping(); err != nil {
		return nil, fmt.Errorf("Ping ClickHouse失败: %v", err)
	}

	log.Printf("[ClickHouse] 连接成功")
	return writer, nil
}

// ping 测试连接
func (w *ClickHouseWriter) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", w.baseURL+"/ping", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(w.cfg.ClickHouse.Username, w.cfg.ClickHouse.Password)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed: status %d", resp.StatusCode)
	}

	return nil
}

// BatchWrite 批量写入日志到 ClickHouse（使用HTTP接口）
func (w *ClickHouseWriter) BatchWrite(entries []models.GameLogDetail) error {
	if len(entries) == 0 {
		return nil
	}

	log.Printf("[ClickHouse] 批量写入 %d 条记录", len(entries))

	// 构建插入数据（JSONEachRow格式）
	var rows []string
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			log.Printf("[ClickHouse] JSON序列化失败: %v", err)
			continue
		}
		rows = append(rows, string(data))
	}

	if len(rows) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询URL
	params := url.Values{}
	params.Set("query", fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow", w.tableName))
	params.Set("database", w.cfg.ClickHouse.Database)

	reqURL := w.baseURL + "?" + params.Encode()

	// 发送数据
	body := strings.Join(rows, "\n")
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBufferString(body))
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.SetBasicAuth(w.cfg.ClickHouse.Username, w.cfg.ClickHouse.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		atomic.AddInt64(&w.errorCount, 1)
		return fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	atomic.AddInt64(&w.flushCount, 1)
	return nil
}

// Exec 执行SQL查询
func (w *ClickHouseWriter) Exec(ctx context.Context, query string) error {
	params := url.Values{}
	params.Set("query", query)
	params.Set("database", w.cfg.ClickHouse.Database)

	reqURL := w.baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(w.cfg.ClickHouse.Username, w.cfg.ClickHouse.Password)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close 关闭写入器
func (w *ClickHouseWriter) Close() error {
	log.Printf("[ClickHouse] 关闭写入器")
	w.httpClient.CloseIdleConnections()
	return nil
}

// GetStats 获取统计信息
func (w *ClickHouseWriter) GetStats() (flushCount, errorCount int64) {
	return atomic.LoadInt64(&w.flushCount),
		atomic.LoadInt64(&w.errorCount)
}

// GetConn 返回nil（HTTP接口不需要原生连接）
func (w *ClickHouseWriter) GetConn() interface{} {
	return w
}
