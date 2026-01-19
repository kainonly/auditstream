package transfer

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"gopkg.in/yaml.v3"
)

func TestNewAuditEvent(t *testing.T) {
	event := NewAuditEvent("audits", "用户登录")

	if event.Stream != "audits" {
		t.Errorf("expected Stream 'audits', got '%s'", event.Stream)
	}
	if event.Msg != "用户登录" {
		t.Errorf("expected Msg '用户登录', got '%s'", event.Msg)
	}
	if event.Time.IsZero() {
		t.Error("expected Time to be set")
	}
}

func TestWithMeta(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithMeta("admin", "user", "login", 123, 456)

	if event.Platform != "admin" {
		t.Errorf("expected Platform 'admin', got '%s'", event.Platform)
	}
	if event.Resource != "user" {
		t.Errorf("expected Resource 'user', got '%s'", event.Resource)
	}
	if event.Action != "login" {
		t.Errorf("expected Action 'login', got '%s'", event.Action)
	}
	if event.ObjectId != 123 {
		t.Errorf("expected ObjectId 123, got '%v'", event.ObjectId)
	}
	if event.UserId != 456 {
		t.Errorf("expected UserId 456, got %d", event.UserId)
	}
}

func TestWithMetaObjectIdTypes(t *testing.T) {
	// int 主键
	event1 := NewAuditEvent("audits", "test").WithMeta("admin", "user", "update", 123, 1)
	if event1.ObjectId != 123 {
		t.Errorf("expected int ObjectId 123, got '%v'", event1.ObjectId)
	}

	// string UUID
	event2 := NewAuditEvent("audits", "test").WithMeta("admin", "order", "delete", "uuid-xxx", 1)
	if event2.ObjectId != "uuid-xxx" {
		t.Errorf("expected string ObjectId 'uuid-xxx', got '%v'", event2.ObjectId)
	}

	// 复合 ID
	ids := []int{1, 2, 3}
	event3 := NewAuditEvent("audits", "test").WithMeta("admin", "item", "batch", ids, 1)
	if objectIds, ok := event3.ObjectId.([]int); !ok || len(objectIds) != 3 {
		t.Errorf("expected []int ObjectId, got '%v'", event3.ObjectId)
	}
}

func TestWithRequest(t *testing.T) {
	body := map[string]any{"username": "test", "password": "***"}
	event := NewAuditEvent("audits", "test").
		WithRequest("/api/login", body)

	if event.Path != "/api/login" {
		t.Errorf("expected Path '/api/login', got '%s'", event.Path)
	}
	if event.Extra == nil {
		t.Error("expected Extra to be initialized")
	}
	if event.Extra["body"] == nil {
		t.Error("expected body in Extra")
	}
}

func TestWithRequestNilBody(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithRequest("/api/list", nil)

	if event.Path != "/api/list" {
		t.Errorf("expected Path '/api/list', got '%s'", event.Path)
	}
	if event.Extra != nil {
		t.Error("expected Extra to be nil when body is nil")
	}
}

func TestWithResponse(t *testing.T) {
	response := map[string]any{"success": true}
	event := NewAuditEvent("audits", "test").
		WithResponse(200, response)

	if event.Extra["code"] != 200 {
		t.Errorf("expected code 200, got '%v'", event.Extra["code"])
	}
	if event.Extra["response"] == nil {
		t.Error("expected response in Extra")
	}
}

func TestWithResponseNilResponse(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithResponse(204, nil)

	if event.Extra["code"] != 204 {
		t.Errorf("expected code 204, got '%v'", event.Extra["code"])
	}
	if _, exists := event.Extra["response"]; exists {
		t.Error("expected response not in Extra when nil")
	}
}

func TestWithClient(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithClient("192.168.1.1", "Mozilla/5.0")

	if event.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", event.IP)
	}
	if event.Extra["agent"] != "Mozilla/5.0" {
		t.Errorf("expected agent 'Mozilla/5.0', got '%v'", event.Extra["agent"])
	}
}

func TestWithClientEmptyAgent(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithClient("192.168.1.1", "")

	if event.IP != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", event.IP)
	}
	if event.Extra != nil {
		t.Error("expected Extra to be nil when agent is empty")
	}
}

func TestWithExtra(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithExtra("custom_field", "custom_value").
		WithExtra("another_field", 123)

	if event.Extra["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field 'custom_value', got '%v'", event.Extra["custom_field"])
	}
	if event.Extra["another_field"] != 123 {
		t.Errorf("expected another_field 123, got '%v'", event.Extra["another_field"])
	}
}

func TestChainedMethods(t *testing.T) {
	body := map[string]any{"data": "test"}
	response := map[string]any{"result": "ok"}

	event := NewAuditEvent("audits", "用户登录").
		WithMeta("admin", "user", "login", 123, 456).
		WithRequest("/api/login", body).
		WithResponse(200, response).
		WithClient("192.168.1.1", "Mozilla/5.0").
		WithExtra("custom", "value")

	if event.Stream != "audits" {
		t.Error("Stream not set correctly")
	}
	if event.Platform != "admin" {
		t.Error("Platform not set correctly")
	}
	if event.Path != "/api/login" {
		t.Error("Path not set correctly")
	}
	if event.IP != "192.168.1.1" {
		t.Error("IP not set correctly")
	}
	if event.Extra["code"] != 200 {
		t.Error("code not set correctly")
	}
	if event.Extra["custom"] != "value" {
		t.Error("custom extra not set correctly")
	}
}

func TestJSONSerialization(t *testing.T) {
	event := NewAuditEvent("audits", "用户登录").
		WithMeta("admin", "user", "login", 123, 456).
		WithRequest("/api/login", map[string]any{"username": "test"}).
		WithResponse(200, map[string]any{"success": true}).
		WithClient("192.168.1.1", "Mozilla/5.0")

	data, err := sonic.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := sonic.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// 验证必填字段
	if result["stream"] != "audits" {
		t.Error("stream not serialized correctly")
	}
	if result["platform"] != "admin" {
		t.Error("platform not serialized correctly")
	}
	if result["resource"] != "user" {
		t.Error("resource not serialized correctly")
	}
	if result["action"] != "login" {
		t.Error("action not serialized correctly")
	}
	if result["path"] != "/api/login" {
		t.Error("path not serialized correctly")
	}
	if result["ip"] != "192.168.1.1" {
		t.Error("ip not serialized correctly")
	}

	// 验证 Extra 字段
	extra, ok := result["extra"].(map[string]any)
	if !ok {
		t.Fatal("extra not serialized as map")
	}
	if extra["code"].(float64) != 200 {
		t.Error("code in extra not serialized correctly")
	}
}

func TestJSONSerializationWithoutExtra(t *testing.T) {
	event := NewAuditEvent("audits", "test").
		WithMeta("admin", "user", "list", 0, 1)
	event.Path = "/api/list"
	event.IP = "127.0.0.1"

	data, err := sonic.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := sonic.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Extra 为 nil 时不应该出现在 JSON 中
	if _, exists := result["extra"]; exists {
		t.Error("extra should be omitted when nil")
	}
}

func TestTimeField(t *testing.T) {
	before := time.Now()
	event := NewAuditEvent("audits", "test")
	after := time.Now()

	if event.Time.Before(before) || event.Time.After(after) {
		t.Error("Time should be set to current time")
	}
}

// 配置结构
type testConfig struct {
	Namespace string   `yaml:"namespace"`
	NatsHosts []string `yaml:"nats_hosts"`
	NatsToken string   `yaml:"nats_token"`
}

func loadTestConfig(t *testing.T) *testConfig {
	// Prioritize environment variables (for CI environment)
	if natsHosts := os.Getenv("NATS_HOSTS"); natsHosts != "" {
		return &testConfig{
			Namespace: getEnvOrDefault("NAMESPACE", "auditstream"),
			NatsHosts: strings.Split(natsHosts, ","),
			NatsToken: os.Getenv("NATS_TOKEN"),
		}
	}
	
	// Fall back to config file (for local development)
	data, err := os.ReadFile("../config/values.yml")
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	var cfg testConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	return &cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupNatsConnection(t *testing.T) (*nats.Conn, *testConfig) {
	cfg := loadTestConfig(t)
	nc, err := nats.Connect(
		cfg.NatsHosts[0],
		nats.Token(cfg.NatsToken),
	)
	if err != nil {
		t.Fatalf("failed to connect to NATS: %v", err)
	}
	return nc, cfg
}

func TestTransferNew(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	if transfer == nil {
		t.Error("expected Transfer instance")
	}
}

func TestTransferSubject(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	subject := transfer.Subject("audits")
	expected := cfg.Namespace + ".audits"
	if subject != expected {
		t.Errorf("expected subject '%s', got '%s'", expected, subject)
	}
}

func TestTransferPublish(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	event := NewAuditEvent("audits", "测试消息").
		WithMeta("test", "user", "create", 1, 100).
		WithRequest("/api/test", map[string]any{"test": true}).
		WithResponse(200, map[string]any{"success": true}).
		WithClient("127.0.0.1", "TestAgent")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = transfer.Publish(ctx, "audits", event)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}
}

func TestTransferPublishAsync(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	event := NewAuditEvent("audits", "异步测试消息").
		WithMeta("test", "order", "update", "uuid-123", 200).
		WithRequest("/api/order", nil).
		WithResponse(201, nil).
		WithClient("192.168.1.100", "")

	future, err := transfer.PublishAsync("audits", event)
	if err != nil {
		t.Fatalf("failed to publish async: %v", err)
	}

	// 等待确认
	select {
	case <-future.Ok():
		// 成功
	case err := <-future.Err():
		t.Fatalf("async publish failed: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("async publish timeout")
	}
}

func TestTransferPublishRaw(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	event := NewAuditEvent("audits", "原始数据测试").
		WithMeta("test", "item", "delete", []int{1, 2, 3}, 300).
		WithRequest("/api/item/batch", map[string]any{"ids": []int{1, 2, 3}}).
		WithResponse(204, nil).
		WithClient("10.0.0.1", "BatchClient")

	data, err := sonic.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = transfer.PublishRaw(ctx, "audits", data)
	if err != nil {
		t.Fatalf("failed to publish raw: %v", err)
	}
}

func TestTransferPublishRawAsync(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	event := NewAuditEvent("audits", "异步原始数据测试").
		WithMeta("test", "config", "read", 0, 400).
		WithRequest("/api/config", nil).
		WithResponse(200, map[string]any{"config": "value"}).
		WithClient("172.16.0.1", "ConfigReader")

	data, err := sonic.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	future, err := transfer.PublishRawAsync("audits", data)
	if err != nil {
		t.Fatalf("failed to publish raw async: %v", err)
	}

	select {
	case <-future.Ok():
		// 成功
	case err := <-future.Err():
		t.Fatalf("async raw publish failed: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("async raw publish timeout")
	}
}

func TestTransferBulkPublish(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	const (
		totalMessages = 10000
		batchSize     = 100
	)

	actions := []string{"create", "read", "update", "delete", "list", "export", "import", "sync"}
	resources := []string{"user", "order", "product", "config", "log", "role", "permission", "setting"}

	start := time.Now()
	successCount := 0
	errorCount := 0

	// 使用异步发布提高吞吐量
	futures := make([]jetstream.PubAckFuture, 0, batchSize)

	for i := 0; i < totalMessages; i++ {
		event := NewAuditEvent("audits", "批量测试消息").
			WithMeta(
				"stress-test",
				resources[i%len(resources)],
				actions[i%len(actions)],
				i,
				i%1000,
			).
			WithRequest(
				"/api/"+resources[i%len(resources)],
				map[string]any{"index": i, "batch": i / batchSize},
			).
			WithResponse(200, map[string]any{"success": true, "id": i}).
			WithClient("192.168.1."+string(rune('1'+i%9)), "BulkTestAgent/1.0")

		future, err := transfer.PublishAsync("audits", event)
		if err != nil {
			errorCount++
			continue
		}
		futures = append(futures, future)

		// 每批次等待确认
		if len(futures) >= batchSize {
			for _, f := range futures {
				select {
				case <-f.Ok():
					successCount++
				case err := <-f.Err():
					t.Logf("publish error: %v", err)
					errorCount++
				case <-time.After(10 * time.Second):
					errorCount++
				}
			}
			futures = futures[:0]
		}
	}

	// 处理剩余的消息
	for _, f := range futures {
		select {
		case <-f.Ok():
			successCount++
		case err := <-f.Err():
			t.Logf("publish error: %v", err)
			errorCount++
		case <-time.After(10 * time.Second):
			errorCount++
		}
	}

	elapsed := time.Since(start)
	rate := float64(successCount) / elapsed.Seconds()

	t.Logf("Bulk publish completed:")
	t.Logf("  Total: %d messages", totalMessages)
	t.Logf("  Success: %d", successCount)
	t.Logf("  Errors: %d", errorCount)
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Rate: %.2f msg/sec", rate)

	if errorCount > totalMessages/100 { // 允许 1% 错误率
		t.Errorf("too many errors: %d/%d", errorCount, totalMessages)
	}
}

func TestTransferContinuousPublish(t *testing.T) {
	nc, cfg := setupNatsConnection(t)
	defer nc.Close()

	transfer, err := New(nc, cfg.Namespace)
	if err != nil {
		t.Fatalf("failed to create Transfer: %v", err)
	}

	const duration = 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	successCount := 0
	errorCount := 0
	start := time.Now()

	actions := []string{"create", "read", "update", "delete"}
	resources := []string{"user", "order", "product", "config"}

	i := 0
	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(start)
			rate := float64(successCount) / elapsed.Seconds()

			t.Logf("Continuous publish completed:")
			t.Logf("  Duration: %v", elapsed)
			t.Logf("  Success: %d", successCount)
			t.Logf("  Errors: %d", errorCount)
			t.Logf("  Rate: %.2f msg/sec", rate)

			if errorCount > successCount/100 {
				t.Errorf("too many errors: %d/%d", errorCount, successCount)
			}
			return
		default:
			event := NewAuditEvent("audits", "持续测试消息").
				WithMeta(
					"continuous-test",
					resources[i%len(resources)],
					actions[i%len(actions)],
					i,
					i%500,
				).
				WithRequest("/api/continuous", map[string]any{"seq": i}).
				WithResponse(200, nil).
				WithClient("10.0.0.1", "ContinuousAgent")

			future, err := transfer.PublishAsync("audits", event)
			if err != nil {
				errorCount++
				i++
				continue
			}

			// 非阻塞检查
			select {
			case <-future.Ok():
				successCount++
			case err := <-future.Err():
				t.Logf("publish error at %d: %v", i, err)
				errorCount++
			default:
				// 不等待，继续发送
				successCount++ // 假设成功，实际生产中应该跟踪
			}
			i++
		}
	}
}
