package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web GUI server",
	Long:  `Start the web-based GUI for browsing and managing TODOs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")

		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("Starting TODO Tracker web server on http://%s\n", addr)

		// Create server
		server := &Server{
			Port: port,
			Host: host,
		}

		// Start server
		return server.Start(addr)
	},
}

type Server struct {
	Port int
	Host string
	DB   *database.DB
}

func (s *Server) Start(addr string) error {
	// Get project path
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Open database
	s.DB, err = database.New(projectPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Determine the path to serve static files from (React app)
	webPath := filepath.Join(projectPath, "web", "dist")

	// Register routes
	// API routes with CORS middleware
	apiHandler := corsMiddleware(http.HandlerFunc(s.handleAPITodos))
	http.Handle("/api/todos", apiHandler)

	apiDetailHandler := corsMiddleware(http.HandlerFunc(s.handleAPITodoDetail))
	http.Handle("/api/todo/", apiDetailHandler)

	statsHandler := corsMiddleware(http.HandlerFunc(s.handleAPIStats))
	http.Handle("/api/stats", statsHandler)

	searchHandler := corsMiddleware(http.HandlerFunc(s.handleAPISearch))
	http.Handle("/api/search", searchHandler)

	// Serve React static files for all other routes (SPA support)
	staticHandler := corsMiddleware(http.HandlerFunc(s.handleStaticFiles(webPath)))
	http.Handle("/", staticHandler)

	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleStaticFiles serves React static files or falls back to index.html for SPA
func (s *Server) handleStaticFiles(webPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// API routes are already handled, skip them
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Get the file path
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		filePath := filepath.Join(webPath, path)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// For SPA, serve index.html for non-file routes
			indexPath := filepath.Join(webPath, "index.html")
			http.ServeFile(w, r, indexPath)
			return
		}

		// Serve the file
		http.ServeFile(w, r, filePath)
	}
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// API Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type TodoListResponse struct {
	Todos []database.TODO `json:"todos"`
	Total int             `json:"total"`
}

type StatsResponse struct {
	Total      int64             `json:"total"`
	ByStatus   map[string]int64  `json:"by_status"`
	ByType     map[string]int64  `json:"by_type"`
	ByPriority map[string]int64  `json:"by_priority"`
}

// handleAPITodos handles GET /api/todos
func (s *Server) handleAPITodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	// Parse query parameters for filtering
	filters := make(map[string]interface{})
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filters["priority"] = priority
	}
	if assignee := r.URL.Query().Get("assignee"); assignee != "" {
		filters["assignee"] = assignee
	}
	if todoType := r.URL.Query().Get("type"); todoType != "" {
		filters["type"] = todoType
	}

	// Get TODOs
	todos, err := s.DB.GetTODOs(filters)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(TodoListResponse{
		Todos: todos,
		Total: len(todos),
	})
}

// handleAPITodoDetail handles GET, PUT, DELETE /api/todo/:id
func (s *Server) handleAPITodoDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/api/todo/")
	if id == "" {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "TODO ID required"})
		return
	}

	switch r.Method {
	case "GET":
		s.handleGetTodo(w, id)
	case "PUT":
		s.handleUpdateTodo(w, r, id)
	case "DELETE":
		s.handleDeleteTodo(w, id)
	default:
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "Method not allowed"})
	}
}

func (s *Server) handleGetTodo(w http.ResponseWriter, id string) {
	todo, err := s.DB.GetTODOByID(id)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "TODO not found"})
		return
	}

	// Return unwrapped TODO object (frontend expects direct object, not wrapped)
	json.NewEncoder(w).Encode(todo)
}

func (s *Server) handleUpdateTodo(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "Invalid request body"})
		return
	}

	// Get existing TODO
	todo, err := s.DB.GetTODOByID(id)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "TODO not found"})
		return
	}

	// Apply updates
	if status, ok := updates["status"].(string); ok {
		todo.Status = status
	}
	if priority, ok := updates["priority"].(string); ok {
		todo.Priority = priority
	}
	if assignee, ok := updates["assignee"].(string); ok {
		todo.Assignee = assignee
	}
	if content, ok := updates["content"].(string); ok {
		todo.Content = content
	}
	if category, ok := updates["category"].(string); ok {
		todo.Category = category
	}
	if dueDate, ok := updates["due_date"].(string); ok {
		if dueDate != "" {
			todo.DueDate = nil // Handle date parsing if needed
		}
	}

	// Save
	if err := s.DB.UpdateTODO(todo); err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: todo})
}

func (s *Server) handleDeleteTodo(w http.ResponseWriter, id string) {
	if err := s.DB.DeleteTODO(id); err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Success: true})
}

// handleAPIStats handles GET /api/stats
func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	stats, err := s.DB.GetStats()
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	// Handle nil maps when database is empty
	byStatus := stats["by_status"]
	byType := stats["by_type"]
	byPriority := stats["by_priority"]

	response := StatsResponse{
		Total:      0,
		ByStatus:   make(map[string]int64),
		ByType:     make(map[string]int64),
		ByPriority: make(map[string]int64),
	}

	if stats["total"] != nil {
		response.Total = stats["total"].(int64)
	}
	if byStatus != nil {
		response.ByStatus = byStatus.(map[string]int64)
	}
	if byType != nil {
		response.ByType = byType.(map[string]int64)
	}
	if byPriority != nil {
		response.ByPriority = byPriority.(map[string]int64)
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPISearch handles GET /api/search
func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode(TodoListResponse{Todos: []database.TODO{}, Total: 0})
		return
	}

	// Search in content, assignee, and file path
	filters := map[string]interface{}{
		"file_path": query,
	}
	todos, err := s.DB.GetTODOs(filters)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Error: err.Error()})
		return
	}

	// Also filter by content in memory for more comprehensive search
	var filtered []database.TODO
	queryLower := strings.ToLower(query)
	for _, todo := range todos {
		if strings.Contains(strings.ToLower(todo.Content), queryLower) ||
			strings.Contains(strings.ToLower(todo.Assignee), queryLower) {
			filtered = append(filtered, todo)
		}
	}

	// If no results from file_path filter, search all
	if len(filtered) == 0 {
		allTodos, _ := s.DB.GetTODOs(nil)
		for _, todo := range allTodos {
			if strings.Contains(strings.ToLower(todo.Content), queryLower) ||
				strings.Contains(strings.ToLower(todo.Assignee), queryLower) ||
				strings.Contains(strings.ToLower(todo.FilePath), queryLower) {
				filtered = append(filtered, todo)
			}
		}
	}

	json.NewEncoder(w).Encode(TodoListResponse{
		Todos: filtered,
		Total: len(filtered),
	})
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to bind the server to")
	rootCmd.AddCommand(serveCmd)
}
