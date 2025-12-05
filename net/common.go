package net

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
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
