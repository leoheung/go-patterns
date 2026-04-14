package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// UserResponse 定义测试用的响应结构体
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// TestRequest 测试 Request 函数
func TestRequest(t *testing.T) {
	// 启动 dummy HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 构造 JSON 响应
		response := UserResponse{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// 初始化 shared HTTP client
	InitDefaultSharedHTTPClient()

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// 发送请求
	body, headers, httpCode, err := Request(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	// 验证 HTTP 状态码
	if httpCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, httpCode)
	}

	// 验证 headers 不为 nil
	if headers == nil {
		t.Fatal("expected headers to be not nil")
	}

	// 验证 Content-Type header
	if headers.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", headers.Get("Content-Type"))
	}

	// 解析 JSON 响应
	var user UserResponse
	if err := ParseResponse(body, &user); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// 验证解析结果
	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got '%s'", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", user.Email)
	}

	// 打印结果
	fmt.Printf("✓ Test passed!\n")
	fmt.Printf("  HTTP Code: %d\n", httpCode)
	fmt.Printf("  Response Body: %s\n", string(body))
	fmt.Printf("  Parsed User: %+v\n", user)
}

// TestRequestWithNilClient 测试 client 未初始化的情况
func TestRequestWithNilClient(t *testing.T) {
	// 重置 client（如果之前已初始化）
	sharedHTTPClient = nil
	once = sync.Once{}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, _, httpCode, err := Request(req)

	if err == nil {
		t.Fatal("expected error when client is nil")
	}

	if httpCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, httpCode)
	}

	fmt.Printf("✓ Test passed! Error: %v\n", err)
}

// TestRequestWithNilRequest 测试 req 为 nil 的情况
func TestRequestWithNilRequest(t *testing.T) {
	// 确保 client 已初始化
	InitDefaultSharedHTTPClient()

	_, _, httpCode, err := Request(nil)

	if err == nil {
		t.Fatal("expected error when request is nil")
	}

	if httpCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, httpCode)
	}

	fmt.Printf("✓ Test passed! Error: %v\n", err)
}
