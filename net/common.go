package net

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leoheung/go-patterns/utils"
)

type UniversalResponse struct {
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
	IsSuccess bool        `json:"isSuccess"`
}

func buildUniversalResponse(data interface{}, err string, isSuccess bool) *UniversalResponse {
	return &UniversalResponse{
		Data:      data,
		Error:     err,
		IsSuccess: isSuccess,
	}
}

// PPrint 格式化打印任意对象
func PPrint(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println("PPrint error:", err)
		return
	}
	fmt.Println(string(b))
}

// ReturnJsonResponse is a helper function to send JSON responses
func ReturnJsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(buildUniversalResponse(payload, "", true)); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func ReturnErrorResponse(w http.ResponseWriter, code int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(buildUniversalResponse(nil, errorMsg, false)); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func IsDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func PreprocessInput(input string) string {
	return strings.TrimSpace(input)
}

// ReturnCSVResponse 返回 CSV 格式的响应
func ReturnCSVResponse(w http.ResponseWriter, filename string, headers []string, rows [][]string) {
	if filename == "" {
		filename = "report.csv"
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)

	cw := csv.NewWriter(w)

	// 写表头
	if len(headers) > 0 {
		if err := cw.Write(headers); err != nil {
			http.Error(w, fmt.Sprintf("failed to write csv header: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 写数据行
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			http.Error(w, fmt.Sprintf("failed to write csv row: %v", err), http.StatusInternalServerError)
			return
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		http.Error(w, fmt.Sprintf("csv flush error: %v", err), http.StatusInternalServerError)
		return
	}
}

// Ptr helpers for common DB types
func PtrString(s string) *string { return &s }
func PtrBool(b bool) *bool       { return &b }

func PtrInt(i int) *int       { return &i }
func PtrInt8(i int8) *int8    { return &i }
func PtrInt16(i int16) *int16 { return &i }
func PtrInt32(i int32) *int32 { return &i }
func PtrInt64(i int64) *int64 { return &i }

func PtrUint(u uint) *uint       { return &u }
func PtrUint8(u uint8) *uint8    { return &u }
func PtrUint16(u uint16) *uint16 { return &u }
func PtrUint32(u uint32) *uint32 { return &u }
func PtrUint64(u uint64) *uint64 { return &u }

func PtrFloat32(f float32) *float32 { return &f }
func PtrFloat64(f float64) *float64 { return &f }

func PtrTime(t time.Time) *time.Time  { return &t }
func PtrUUID(id uuid.UUID) *uuid.UUID { return &id }

func PtrBytes(b []byte) *[]byte { return &b }

func DelayDo(d time.Duration, fn func()) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	<-timer.C
	fn()
}

// RetryWork 执行工作函数，捕获panic或error并最多重试retryTimes次
// work: 需要执行的工作函数
// retryTimes: 最大重试次数（不包括首次执行）
func RetryWork(work func() error, retryTimes int) {
	totalAttempts := retryTimes + 1
	for attempt := 0; attempt < totalAttempts; attempt++ {
		var err error
		func(attempt int) {
			defer func() {
				if r := recover(); r != nil {
					// 捕获panic，转换为错误
					err = fmt.Errorf("panic: %v", r)
				}
			}()
			// 执行工作函数并获取返回的错误
			err = work()
		}(attempt)

		// 判断是否需要重试
		if err == nil {
			utils.LogMessage(fmt.Sprintf("尝试 %d 成功", attempt+1))
			return // 成功，退出重试
		}

		utils.LogMessage(fmt.Sprintf("业务逻辑出现error/panic: %s", err.Error()))

		// 失败处理
		if attempt < totalAttempts-1 {
			utils.LogMessage(fmt.Sprintf("尝试 %d 失败: %v，将重试...", attempt+1, err))
			time.Sleep(500 * time.Millisecond)
		} else {
			utils.LogMessage(fmt.Sprintf("最后一次尝试 %d 失败: %v，已耗尽重试次数", attempt+1, err))
		}
	}
}
