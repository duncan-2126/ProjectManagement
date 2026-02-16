package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

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

	// Register routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/todos", s.handleTodos)
	http.HandleFunc("/todo/", s.handleTodoDetail)
	http.HandleFunc("/api/todos", s.handleAPITodos)
	http.HandleFunc("/api/todo/", s.handleAPITodoUpdate)
	http.HandleFunc("/static/", s.handleStatic)

	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// PageData holds common data for templates
type PageData struct {
	Title string
	Todos []database.TODO
	Stats map[string]interface{}
}

// handleDashboard serves the main dashboard page
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get all TODOs
	todos, err := s.DB.GetTODOs(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get stats
	stats, err := s.DB.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "TODO Tracker Dashboard",
		Todos: todos,
		Stats: stats,
	}

	// Render template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="/static/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header>
            <h1>TODO Tracker</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/todos">All TODOs</a>
            </nav>
        </header>

        <main>
            <section class="dashboard-stats">
                <h2>Statistics</h2>
                <div class="stats-grid">
                    {{range $key, $value := .Stats}}
                        <div class="stat-card">
                            <h3>{{$key}}</h3>
                            <p>{{$value}}</p>
                        </div>
                    {{end}}
                </div>
            </section>

            <section class="recent-todos">
                <h2>Recent TODOs</h2>
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>File</th>
                            <th>Type</th>
                            <th>Status</th>
                            <th>Priority</th>
                            <th>Content</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Todos}}
                        <tr>
                            <td>{{.ID}}</td>
                            <td>{{.FilePath}}:{{.LineNumber}}</td>
                            <td>{{.Type}}</td>
                            <td>{{.Status}}</td>
                            <td>{{.Priority}}</td>
                            <td>{{.Content}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </section>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleTodos serves the list of all TODOs
func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	// Get all TODOs
	todos, err := s.DB.GetTODOs(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: "All TODOs",
		Todos: todos,
	}

	// Render template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="/static/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header>
            <h1>TODO Tracker</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/todos">All TODOs</a>
            </nav>
        </header>

        <main>
            <h2>{{.Title}}</h2>
            <div class="filters">
                <form method="GET">
                    <input type="text" name="search" placeholder="Search...">
                    <select name="status">
                        <option value="">All Statuses</option>
                        <option value="open">Open</option>
                        <option value="in_progress">In Progress</option>
                        <option value="resolved">Resolved</option>
                        <option value="closed">Closed</option>
                    </select>
                    <button type="submit">Filter</button>
                </form>
            </div>

            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>File</th>
                        <th>Type</th>
                        <th>Status</th>
                        <th>Priority</th>
                        <th>Assignee</th>
                        <th>Content</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Todos}}
                    <tr>
                        <td>{{.ID}}</td>
                        <td>{{.FilePath}}:{{.LineNumber}}</td>
                        <td>{{.Type}}</td>
                        <td>{{.Status}}</td>
                        <td>{{.Priority}}</td>
                        <td>{{.Assignee}}</td>
                        <td>{{.Content}}</td>
                        <td>
                            <a href="/todo/{{.ID}}">View</a>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("todos").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleTodoDetail serves the detail page for a specific TODO
func (s *Server) handleTodoDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/todo/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	// Get TODO by ID
	todo, err := s.DB.GetTODOByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Render template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>TODO Detail</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="/static/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <header>
            <h1>TODO Tracker</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/todos">All TODOs</a>
            </nav>
        </header>

        <main>
            <h2>TODO Detail</h2>
            <div class="todo-detail">
                <form method="POST" action="/api/todo/{{.ID}}">
                    <div class="form-group">
                        <label>ID:</label>
                        <input type="text" value="{{.ID}}" readonly>
                    </div>

                    <div class="form-group">
                        <label>File Path:</label>
                        <input type="text" value="{{.FilePath}}" readonly>
                    </div>

                    <div class="form-group">
                        <label>Line Number:</label>
                        <input type="text" value="{{.LineNumber}}" readonly>
                    </div>

                    <div class="form-group">
                        <label>Type:</label>
                        <input type="text" value="{{.Type}}" readonly>
                    </div>

                    <div class="form-group">
                        <label>Status:</label>
                        <select name="status">
                            <option value="open" {{if eq .Status "open"}}selected{{end}}>Open</option>
                            <option value="in_progress" {{if eq .Status "in_progress"}}selected{{end}}>In Progress</option>
                            <option value="blocked" {{if eq .Status "blocked"}}selected{{end}}>Blocked</option>
                            <option value="resolved" {{if eq .Status "resolved"}}selected{{end}}>Resolved</option>
                            <option value="wontfix" {{if eq .Status "wontfix"}}selected{{end}}>Won't Fix</option>
                            <option value="closed" {{if eq .Status "closed"}}selected{{end}}>Closed</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>Priority:</label>
                        <select name="priority">
                            <option value="P0" {{if eq .Priority "P0"}}selected{{end}}>P0 - Critical</option>
                            <option value="P1" {{if eq .Priority "P1"}}selected{{end}}>P1 - High</option>
                            <option value="P2" {{if eq .Priority "P2"}}selected{{end}}>P2 - Medium</option>
                            <option value="P3" {{if eq .Priority "P3"}}selected{{end}}>P3 - Low</option>
                            <option value="P4" {{if eq .Priority "P4"}}selected{{end}}>P4 - Trivial</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>Assignee:</label>
                        <input type="text" name="assignee" value="{{.Assignee}}">
                    </div>

                    <div class="form-group">
                        <label>Content:</label>
                        <textarea name="content">{{.Content}}</textarea>
                    </div>

                    <button type="submit">Update TODO</button>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("todo-detail").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAPITodos handles API requests for TODOs
func (s *Server) handleAPITodos(w http.ResponseWriter, r *http.Request) {
	// Get all TODOs
	todos, err := s.DB.GetTODOs(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON and send
	w.Header().Set("Content-Type", "application/json")
	// In a real implementation, we'd marshal the todos to JSON here
	fmt.Fprintf(w, `{"todos": [%d items]}`, len(todos))
}

// handleAPITodoUpdate handles API requests to update a TODO
func (s *Server) handleAPITodoUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/todo/")
	if id == "" {
		http.Error(w, "TODO ID required", http.StatusBadRequest)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get TODO by ID
	todo, err := s.DB.GetTODOByID(id)
	if err != nil {
		http.Error(w, "TODO not found", http.StatusNotFound)
		return
	}

	// Update fields from form
	if status := r.FormValue("status"); status != "" {
		todo.Status = status
	}

	if priority := r.FormValue("priority"); priority != "" {
		todo.Priority = priority
	}

	if assignee := r.FormValue("assignee"); assignee != "" {
		todo.Assignee = assignee
	}

	if content := r.FormValue("content"); content != "" {
		todo.Content = content
	}

	// Save updated TODO
	err = s.DB.UpdateTODO(todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to detail page
	http.Redirect(w, r, "/todo/"+id, http.StatusSeeOther)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, we'd serve static files from a directory
	// For now, we'll just return a simple CSS file
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
		fmt.Fprintf(w, `
.container {
	max-width: 1200px;
	margin: 0 auto;
	padding: 20px;
}

header {
	background: #333;
	color: white;
	padding: 1rem;
	margin-bottom: 2rem;
}

nav a {
	color: white;
	text-decoration: none;
	margin-right: 1rem;
}

.stats-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	gap: 1rem;
	margin: 2rem 0;
}

.stat-card {
	background: #f5f5f5;
	padding: 1rem;
	border-radius: 4px;
	text-align: center;
}

table {
	width: 100%;
	border-collapse: collapse;
	margin: 1rem 0;
}

th, td {
	padding: 0.5rem;
	text-align: left;
	border-bottom: 1px solid #ddd;
}

th {
	background: #f5f5f5;
}

.form-group {
	margin-bottom: 1rem;
}

.form-group label {
	display: block;
	margin-bottom: 0.5rem;
	font-weight: bold;
}

.form-group input, .form-group select, .form-group textarea {
	width: 100%;
	padding: 0.5rem;
	border: 1px solid #ddd;
	border-radius: 4px;
}

button {
	background: #007cba;
	color: white;
	padding: 0.5rem 1rem;
	border: none;
	border-radius: 4px;
	cursor: pointer;
}

button:hover {
	background: #005a87;
}

.filters {
	background: #f5f5f5;
	padding: 1rem;
	margin-bottom: 1rem;
	border-radius: 4px;
}
`)
	} else {
		http.NotFound(w, r)
	}
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to bind the server to")
	rootCmd.AddCommand(serveCmd)
}