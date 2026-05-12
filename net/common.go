package net

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
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

// SafelyReadBody reads the body (io.ReadCloser) safely, limiting the size to maxBodyMB MB.
// It returns the body content as a byte slice and any error that occurred during the read operation.
// It will close the body after reading.
func SafelyReadBody(body io.ReadCloser, maxBodyMB *int) (data []byte, err error) {
	if maxBodyMB != nil && *maxBodyMB < 0 {
		return nil, fmt.Errorf("maxBodyMB < 0")
	}

	defer body.Close()

	if maxBodyMB == nil {
		return io.ReadAll(body)
	}

	actualLimit := int64(*maxBodyMB) * 1024 * 1024
	reader := io.LimitReader(body, actualLimit+1)

	data, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > actualLimit {
		return nil, fmt.Errorf("body exceeds %dMB limit", *maxBodyMB)
	}

	return data, nil
}

// DeepCopyRequest creates a deep copy of the HTTP request r, including the body content.
// It limits the body size to maxBodyMB MB.
// If the body size exceeds maxBodyMB MB, it returns an error.
func DeepCopyRequest(r *http.Request, maxBodyMB int) (*http.Request, error) {
	if r == nil {
		return nil, fmt.Errorf("r is nil")
	}

	if maxBodyMB <= 0 {
		return nil, fmt.Errorf("maxBodyMB <= 0")
	}

	limit := int64(maxBodyMB) * 1024 * 1024
	if r.ContentLength > limit {
		return nil, fmt.Errorf("request body too large: %d MB > %d MB",
			r.ContentLength/(1024*1024), maxBodyMB)
	}

	// 1. safely read body
	bodyBytes, err := SafelyReadBody(r.Body, &maxBodyMB)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// 2. 为原始请求创建新的 Body（以便后续使用）
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// 3. 创建副本请求，使用新的 Body
	reqCopy := r.Clone(context.Background())
	reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return reqCopy, nil
}

// DeepCopyResponse creates a deep copy of the HTTP response r, including the body content.
// It limits the body size to maxBodyMB MB.
// If the body size exceeds maxBodyMB MB, it returns an error.
// Note: Request and TLS fields are shared (not deep copied) as they are typically read-only.
func DeepCopyResponse(r *http.Response, maxBodyMB int) (*http.Response, error) {
	if r == nil {
		return nil, fmt.Errorf("r is nil")
	}

	if maxBodyMB <= 0 {
		return nil, fmt.Errorf("maxBodyMB <= 0")
	}

	// 1. Create basic copy with value types and cloned headers
	respCopy := &http.Response{
		Status:           r.Status,
		StatusCode:       r.StatusCode,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header.Clone(),
		ContentLength:    r.ContentLength,
		TransferEncoding: slices.Clone(r.TransferEncoding),
		Close:            r.Close,
		Uncompressed:     r.Uncompressed,
		Trailer:          r.Trailer.Clone(),
		Request:          r.Request, // shared pointer
		TLS:              r.TLS,     // shared pointer (read-only)
	}

	// 2. Handle body (similar to DeepCopyRequest)
	if r.Body != nil && r.Body != http.NoBody {
		bodyBytes, err := SafelyReadBody(r.Body, &maxBodyMB)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		respCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return respCopy, nil
}
