package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Request represents a minimal JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a minimal JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError is a JSON-RPC error payload.
type ResponseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeResult is returned for the "initialize" method.
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]interface{} `json:"serverInfo"`
}

// Tool describes an MCP tool.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ListToolsResult is returned by "tools/list".
type ListToolsResult struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// ToolCallParams are the parameters for "tools/call".
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ContentItem represents a piece of tool output.
type ContentItem struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	JSON json.RawMessage `json:"json,omitempty"`
}

// ToolCallResult wraps tool output content.
type ToolCallResult struct {
	Content []ContentItem `json:"content"`
}

// MCPServer handles MCP requests over stdio.
type MCPServer struct {
	baseURL string
	client  *http.Client
	in      *bufio.Reader
	out     *bufio.Writer
	outMu   sync.Mutex
	tools   []Tool
}

func main() {
	// 중요: 로그는 반드시 Stderr로 출력해야 합니다. (Stdout은 통신용)
	log.SetOutput(os.Stderr)

	baseURL := strings.TrimRight(getEnv("KOSIS_BASE_URL", "http://localhost:8080/api/v1/mcp"), "/")
	server := &MCPServer{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		in:  bufio.NewReader(os.Stdin),
		out: bufio.NewWriter(os.Stdout),
		tools: []Tool{
			{
				Name:        "reports_by_corp_name",
				Description: "List recent reports matching a partial corp_name.",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"corp_name": map[string]interface{}{
							"type":        "string",
							"description": "Partial or full company name (case-insensitive).",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"minimum":     1,
							"maximum":     100,
							"description": "Number of reports to return (default 10).",
						},
					},
					"required": []string{"corp_name"},
				},
			},
		},
	}

	log.Println("MCP Shim Server starting...")
	if err := server.Serve(); err != nil {
		log.Fatalf("mcp server failed: %v", err)
	}
}

// Serve starts the read/dispatch/write loop.
func (s *MCPServer) Serve() error {
	for {
		req, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			// 파싱 에러나 빈 줄은 로그만 남기고 계속 진행
			// (치명적이지 않은 에러로 간주)
			if err.Error() != "empty line" {
				log.Printf("failed to read/parse message: %v", err)
			}
			continue
		}

		// Notification (ID가 없거나 null)인 경우도 처리 로직은 타야 함 (ex: initialized)
		// 다만 응답은 보내지 않음.

		// Handle request concurrently
		go func(r Request) {
			resp := s.handleRequest(r)

			// 응답이 nil이면(예: Notification) 아무것도 보내지 않음
			if resp == nil {
				return
			}

			if err := s.writeMessage(*resp); err != nil {
				log.Printf("failed to write message: %v", err)
			}
		}(req)
	}
}

// handleRequest routes a single MCP request.
func (s *MCPServer) handleRequest(req Request) *Response {
	switch req.Method {
	case "initialize":
		return s.reply(req, InitializeResult{
			ProtocolVersion: "2024-11-05", // 최신 프로토콜 버전 명시
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{}, // Tool 지원 명시
			},
			ServerInfo: map[string]interface{}{
				"name":    "go-mcp-shim",
				"version": "1.0.0",
			},
		})

	// [추가] 초기화 완료 알림: 응답 불필요 (nil 반환)
	case "notifications/initialized":
		return nil

	case "tools/list":
		return s.reply(req, ListToolsResult{Tools: s.tools})
	case "tools/call":
		return s.handleToolCall(req)
	case "ping":
		return s.reply(req, map[string]interface{}{})
	case "shutdown":
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}()
		return s.reply(req, nil)
	case "notifications/exit":
		os.Exit(0)
		return nil
	}

	// 메서드를 찾을 수 없는 경우
	return s.error(req, -32601, fmt.Sprintf("method not found: %s", req.Method), nil)
}

func (s *MCPServer) handleToolCall(req Request) *Response {
	var params ToolCallParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.error(req, -32602, "invalid params", err.Error())
		}
	}

	switch params.Name {
	case "reports_by_corp_name":
		result, rpcErr := s.callReportsByCorpName(params.Arguments)
		if rpcErr != nil {
			return &Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   rpcErr,
			}
		}
		return s.reply(req, result)
	default:
		return s.error(req, -32601, fmt.Sprintf("tool not found: %s", params.Name), nil)
	}
}

func (s *MCPServer) callReportsByCorpName(args map[string]interface{}) (*ToolCallResult, *ResponseError) {
	rawCorp, ok := args["corp_name"]
	if !ok {
		return nil, &ResponseError{Code: -32602, Message: "corp_name is required"}
	}

	corpName, ok := rawCorp.(string)
	if !ok || strings.TrimSpace(corpName) == "" {
		return nil, &ResponseError{Code: -32602, Message: "corp_name must be a non-empty string"}
	}
	corpName = strings.TrimSpace(corpName)

	limit := 10
	if rawLimit, ok := args["limit"]; ok {
		switch v := rawLimit.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		case json.Number:
			if i, err := strconv.Atoi(string(v)); err == nil {
				limit = i
			}
		default:
			// limit 파싱 실패 시 기본값 사용하거나 에러 처리 (여기선 관대하게 넘어감)
		}
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	urlStr := fmt.Sprintf("%s/reports/by-corp-name?corp_name=%s&limit=%d", s.baseURL, urlEncode(corpName), limit)

	// 로그 추가 (디버깅용)
	log.Printf("Calling upstream: %s", urlStr)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, &ResponseError{Code: -32000, Message: "failed to build request", Data: err.Error()}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &ResponseError{Code: -32000, Message: "request failed", Data: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ResponseError{Code: -32000, Message: "failed to read response", Data: err.Error()}
	}

	if resp.StatusCode >= 300 {
		return nil, &ResponseError{Code: -32000, Message: fmt.Sprintf("upstream error: %s", resp.Status), Data: string(body)}
	}

	// Upstream API가 JSON을 리턴한다고 가정
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		// JSON이 아닐 경우 텍스트로 래핑
		raw = json.RawMessage(fmt.Sprintf("%q", string(body)))
	}

	return &ToolCallResult{
		Content: []ContentItem{
			{
				Type: "text", // Claude에게 JSON 텍스트로 보여주는 것이 일반적임
				Text: string(body),
			},
		},
	}, nil
}

func (s *MCPServer) reply(req Request, result interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *MCPServer) error(req Request, code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &ResponseError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// [수정됨] readMessage: Content-Length 헤더 없이 줄바꿈(NDJSON) 단위로 읽음
func (s *MCPServer) readMessage() (Request, error) {
	// '\n'을 만날 때까지 읽음 (한 줄이 하나의 JSON 메시지)
	line, err := s.in.ReadBytes('\n')
	if err != nil {
		return Request{}, err
	}

	// 앞뒤 공백 제거
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return Request{}, fmt.Errorf("empty line")
	}

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return Request{}, fmt.Errorf("json parse error: %w", err)
	}

	return req, nil
}

// [수정됨] writeMessage: Content-Length 헤더 없이 JSON + 줄바꿈 전송
func (s *MCPServer) writeMessage(resp Response) error {
	s.outMu.Lock()
	defer s.outMu.Unlock()

	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// JSON 데이터 쓰기
	if _, err := s.out.Write(payload); err != nil {
		return err
	}
	// 줄바꿈 추가 (NDJSON 구분자)
	if _, err := s.out.Write([]byte("\n")); err != nil {
		return err
	}

	return s.out.Flush()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func urlEncode(v string) string {
	return url.QueryEscape(v)
}
