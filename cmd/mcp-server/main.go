package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// MCPTool представляет инструмент MCP
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPRequest представляет запрос к MCP серверу
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// MCPResponse представляет ответ от MCP сервера
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError представляет ошибку MCP
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// FileInfo представляет информацию о файле
type FileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// DatabaseQuery представляет запрос к базе данных
type DatabaseQuery struct {
	Query  string                 `json:"query"`
	Params map[string]interface{} `json:"params"`
}

// APICall представляет вызов API
type APICall struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origins для разработки
	},
}

// MCP сервер
type MCPServer struct {
	tools map[string]func(context.Context, interface{}) (interface{}, error)
}

// NewMCPServer создает новый MCP сервер
func NewMCPServer() *MCPServer {
	server := &MCPServer{
		tools: make(map[string]func(context.Context, interface{}) (interface{}, error)),
	}

	// Регистрируем инструменты
	server.registerTools()

	return server
}

// registerTools регистрирует все доступные инструменты
func (s *MCPServer) registerTools() {
	// Инструменты для работы с файлами
	s.tools["list_dir"] = s.listDirectory
	s.tools["read_file"] = s.readFile
	s.tools["write_file"] = s.writeFile
	s.tools["delete_file"] = s.deleteFile
	s.tools["search_files"] = s.searchFiles

	// Инструменты для работы с базой данных
	s.tools["db_query"] = s.databaseQuery
	s.tools["db_execute"] = s.databaseExecute

	// Инструменты для работы с API
	s.tools["api_call"] = s.apiCall
	s.tools["list_endpoints"] = s.listEndpoints

	// Инструменты для работы с проектом
	s.tools["run_command"] = s.runCommand
	s.tools["get_project_info"] = s.getProjectInfo
	s.tools["deploy_service"] = s.deployService
}

// добавляем helper
func (s *MCPServer) getTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "list_dir",
			Description: "List directory contents",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path to list",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read file contents",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path to read",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write",
					},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "delete_file",
			Description: "Delete file or directory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File or directory path to delete",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "search_files",
			Description: "Search files by pattern",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query or pattern",
					},
					"include_pattern": map[string]interface{}{
						"type":        "string",
						"description": "File pattern to include",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "db_query",
			Description: "Execute database query",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query to execute",
					},
					"params": map[string]interface{}{
						"type":        "object",
						"description": "Query parameters",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "api_call",
			Description: "Make API call to service",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method",
					},
					"url": map[string]interface{}{
						"type":        "string",
						"description": "API endpoint URL",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "HTTP headers",
					},
					"body": map[string]interface{}{
						"type":        "object",
						"description": "Request body",
					},
				},
				"required": []string{"method", "url"},
			},
		},
		{
			Name:        "run_command",
			Description: "Run terminal command",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Command to execute",
					},
					"cwd": map[string]interface{}{
						"type":        "string",
						"description": "Working directory",
					},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "get_project_info",
			Description: "Get project information and status",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// обновляем handleToolsList
func (s *MCPServer) handleToolsList(request MCPRequest) MCPResponse {
	tools := s.getTools()
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleWebSocket обрабатывает WebSocket соединения
func (s *MCPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("MCP WebSocket connection established")

	for {
		// Читаем сообщение
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Парсим запрос
		var request MCPRequest
		if err := json.Unmarshal(message, &request); err != nil {
			log.Printf("Failed to parse request: %v", err)
			continue
		}

		// Обрабатываем запрос
		response := s.handleRequest(request)

		// Отправляем ответ
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}

		// После успешной инициализации отправляем список инструментов
		if request.Method == "initialize" {
			notif := map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "tools/list",
				"params": map[string]interface{}{
					"tools": s.getTools(),
				},
			}
			if nb, err := json.Marshal(notif); err == nil {
				if err := conn.WriteMessage(websocket.TextMessage, nb); err != nil {
					log.Printf("WebSocket write error (notification): %v", err)
				}
			} else {
				log.Printf("Failed to marshal notification: %v", err)
			}
		}
	}
}

// handleHTTP обрабатывает HTTP POST запросы (streamableHttp transport)
func (s *MCPServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(MCPResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &MCPError{Code: -32700, Message: "Invalid JSON"},
		})
		return
	}

	resp := s.handleRequest(req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("HTTP encode error: %v", err)
	}
}

// handleRequest обрабатывает MCP запрос
func (s *MCPServer) handleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolCall(request)
	case "initialize":
		return s.handleInitialize(request)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize обрабатывает инициализацию
func (s *MCPServer) handleInitialize(request MCPRequest) MCPResponse {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "FunPay MCP Server",
				"version": "1.0.0",
			},
		},
	}

	return response
}

// handleToolCall обрабатывает вызов инструмента
func (s *MCPServer) handleToolCall(request MCPRequest) MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	name, ok := params["name"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Tool name is required",
			},
		}
	}

	arguments, _ := params["arguments"].(map[string]interface{})

	toolFunc, exists := s.tools[name]
	if !exists {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}

	result, err := toolFunc(context.Background(), arguments)
	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		},
	}
}

// Реализация инструментов

func (s *MCPServer) listDirectory(ctx context.Context, args interface{}) (interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	path, ok := argsMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	// Получаем абсолютный путь
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Проверяем, что путь находится в пределах проекта
	projectRoot, _ := filepath.Abs(".")
	if !strings.HasPrefix(absPath, projectRoot) {
		return nil, fmt.Errorf("access denied: path outside project")
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:     entry.Name(),
			Path:     filepath.Join(absPath, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return files, nil
}

func (s *MCPServer) readFile(ctx context.Context, args interface{}) (interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	path, ok := argsMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	// Проверяем безопасность пути
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	projectRoot, _ := filepath.Abs(".")
	if !strings.HasPrefix(absPath, projectRoot) {
		return nil, fmt.Errorf("access denied: path outside project")
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

func (s *MCPServer) writeFile(ctx context.Context, args interface{}) (interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	path, ok := argsMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	content, ok := argsMap["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	// Проверяем безопасность пути
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	projectRoot, _ := filepath.Abs(".")
	if !strings.HasPrefix(absPath, projectRoot) {
		return nil, fmt.Errorf("access denied: path outside project")
	}

	// Создаем директории если нужно
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	return "File written successfully", nil
}

func (s *MCPServer) deleteFile(ctx context.Context, args interface{}) (interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	path, ok := argsMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}

	// Проверяем безопасность пути
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	projectRoot, _ := filepath.Abs(".")
	if !strings.HasPrefix(absPath, projectRoot) {
		return nil, fmt.Errorf("access denied: path outside project")
	}

	if err := os.RemoveAll(absPath); err != nil {
		return nil, fmt.Errorf("failed to delete: %v", err)
	}

	return "Deleted successfully", nil
}

func (s *MCPServer) searchFiles(ctx context.Context, args interface{}) (interface{}, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments")
	}

	query, ok := argsMap["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	includePattern, _ := argsMap["include_pattern"].(string)

	// Простая реализация поиска файлов
	var results []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Проверяем паттерн включения
		if includePattern != "" {
			matched, _ := filepath.Match(includePattern, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		// Читаем содержимое файла и ищем совпадения
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), query) || strings.Contains(path, query) {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	return results, nil
}

func (s *MCPServer) databaseQuery(ctx context.Context, args interface{}) (interface{}, error) {
	// Заглушка для работы с базой данных
	// В реальной реализации здесь будет подключение к PostgreSQL/MongoDB
	return "Database query functionality not implemented yet", nil
}

func (s *MCPServer) databaseExecute(ctx context.Context, args interface{}) (interface{}, error) {
	// Заглушка для выполнения команд БД
	return "Database execute functionality not implemented yet", nil
}

func (s *MCPServer) apiCall(ctx context.Context, args interface{}) (interface{}, error) {
	// Заглушка для API вызовов
	return "API call functionality not implemented yet", nil
}

func (s *MCPServer) listEndpoints(ctx context.Context, args interface{}) (interface{}, error) {
	// Возвращаем список доступных эндпоинтов
	endpoints := map[string][]string{
		"user-service": {
			"POST /api/users/register",
			"POST /api/users/login",
			"GET /api/users/profile",
			"PUT /api/users/profile",
		},
		"product-service": {
			"GET /api/products",
			"GET /api/products/{id}",
			"POST /api/products",
			"PUT /api/products/{id}",
			"DELETE /api/products/{id}",
		},
		"order-service": {
			"GET /api/orders",
			"GET /api/orders/{id}",
			"POST /api/orders",
			"PUT /api/orders/{id}/status",
		},
		"payment-service": {
			"POST /api/payments/create",
			"GET /api/payments/{id}",
			"POST /api/payments/{id}/confirm",
		},
		"chat-service": {
			"GET /api/chat/conversations",
			"GET /api/chat/conversations/{id}/messages",
			"POST /api/chat/conversations/{id}/messages",
		},
	}

	return endpoints, nil
}

func (s *MCPServer) runCommand(ctx context.Context, args interface{}) (interface{}, error) {
	// Заглушка для выполнения команд
	return "Command execution functionality not implemented yet", nil
}

func (s *MCPServer) getProjectInfo(ctx context.Context, args interface{}) (interface{}, error) {
	info := map[string]interface{}{
		"name":        "FunPay - Gaming Marketplace",
		"description": "Микросервисная платформа для торговли игровыми аккаунтами",
		"architecture": map[string]interface{}{
			"backend":    "Go (microservices)",
			"frontend":   "Next.js",
			"database":   "PostgreSQL + MongoDB + Redis",
			"deployment": "Docker + Kubernetes",
		},
		"services": []string{
			"api-gateway",
			"user-service",
			"product-service",
			"order-service",
			"payment-service",
			"chat-service",
			"notification-service",
		},
		"status": "Development",
	}

	return info, nil
}

func (s *MCPServer) deployService(ctx context.Context, args interface{}) (interface{}, error) {
	// Заглушка для деплоя сервисов
	return "Deploy functionality not implemented yet", nil
}

func main() {
	server := NewMCPServer()

	router := mux.NewRouter()

	// WebSocket endpoint для MCP
	router.HandleFunc("/mcp", server.handleWebSocket)

	// HTTP endpoint для проверки статуса
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "mcp-server",
		})
	})

	// HTTP endpoint для streamableHttp POST
	router.HandleFunc("/sse", server.handleHTTP).Methods(http.MethodPost)

	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("MCP Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
