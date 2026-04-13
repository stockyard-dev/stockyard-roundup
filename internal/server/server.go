package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/stockyard-dev/stockyard-roundup/internal/store"
	"github.com/stockyard-dev/stockyard/bus"
)

// Resource keys for the extras table.
// Roundup has two resources but only tasks get custom fields in practice
// (projects are simple containers). We declare both for cleanliness.
const (
	resourceTasks    = "tasks"
	resourceProjects = "projects"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
	bus     *bus.Bus // optional cross-tool event bus; nil if not configured
}

func New(db *store.DB, limits Limits, dataDir string, b *bus.Bus) *Server {
	s := &Server{
		db:      db,
		mux:     http.NewServeMux(),
		limits:  limits,
		dataDir: dataDir,
		bus:     b,
	}
	s.loadPersonalConfig()

	// Project routes
	s.mux.HandleFunc("GET /api/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/projects", s.createProject)
	s.mux.HandleFunc("GET /api/projects/{id}", s.getProject)
	s.mux.HandleFunc("PUT /api/projects/{id}", s.updateProject)
	s.mux.HandleFunc("DELETE /api/projects/{id}", s.deleteProject)

	// Task routes
	s.mux.HandleFunc("GET /api/tasks", s.listTasks)
	s.mux.HandleFunc("POST /api/tasks", s.createTask)
	s.mux.HandleFunc("GET /api/tasks/{id}", s.getTask)
	s.mux.HandleFunc("PUT /api/tasks/{id}", s.updateTask)
	s.mux.HandleFunc("DELETE /api/tasks/{id}", s.deleteTask)
	s.mux.HandleFunc("PATCH /api/tasks/{id}/status", s.setStatus)

	// Stats / health
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// Personalization
	s.mux.HandleFunc("GET /api/config", s.configHandler)

	// Extras (custom fields)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)

	// Tier
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{
			"tier":        s.limits.Tier,
			"upgrade_url": "https://stockyard.dev/roundup/",
		})
	})

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ─── helpers ──────────────────────────────────────────────────────

func wj(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func we(w http.ResponseWriter, code int, msg string) {
	wj(w, code, map[string]string{"error": msg})
}

func ot(t []store.Task) []store.Task {
	if t == nil {
		return []store.Task{}
	}
	return t
}

func op(p []store.Project) []store.Project {
	if p == nil {
		return []store.Project{}
	}
	return p
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}

// ─── personalization ──────────────────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("roundup: warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("roundup: loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// ─── extras ───────────────────────────────────────────────────────

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	wj(w, 200, out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		we(w, 400, "read body")
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		we(w, 500, "save failed")
		return
	}
	wj(w, 200, map[string]string{"ok": "saved"})
}

// ─── projects ─────────────────────────────────────────────────────

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"projects": op(s.db.ListProjects())})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var p store.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if p.Name == "" {
		we(w, 400, "name required")
		return
	}
	if err := s.db.CreateProject(&p); err != nil {
		we(w, 500, "create failed")
		return
	}
	wj(w, 201, s.db.GetProject(p.ID))
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p := s.db.GetProject(r.PathValue("id"))
	if p == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, p)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetProject(r.PathValue("id"))
	if existing == nil {
		we(w, 404, "not found")
		return
	}
	var patch store.Project
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		we(w, 400, "invalid json")
		return
	}
	patch.ID = existing.ID
	if patch.Name == "" {
		patch.Name = existing.Name
	}
	if patch.Color == "" {
		patch.Color = existing.Color
	}
	if err := s.db.UpdateProject(&patch); err != nil {
		we(w, 500, "update failed")
		return
	}
	wj(w, 200, s.db.GetProject(patch.ID))
}

// deleteProject removes the project AND all its tasks. Cascades extras
// for both the project itself and every task in it.
func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	taskIDs := s.db.TasksInProject(id)
	for _, tid := range taskIDs {
		s.db.DeleteExtras(resourceTasks, tid)
	}
	s.db.DeleteExtras(resourceProjects, id)
	if err := s.db.DeleteProject(id); err != nil {
		we(w, 500, "delete failed")
		return
	}
	wj(w, 200, map[string]string{"deleted": "ok"})
}

// ─── tasks ────────────────────────────────────────────────────────

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{
		"tasks": ot(s.db.ListTasks(
			r.URL.Query().Get("project"),
			r.URL.Query().Get("status"),
			r.URL.Query().Get("priority"),
		)),
	})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	if s.limits.MaxItems > 0 && len(s.db.ListTasks("", "", "")) >= s.limits.MaxItems {
		we(w, 402, "Free tier limit reached. Upgrade at https://stockyard.dev/roundup/")
		return
	}
	var t store.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if t.Title == "" {
		we(w, 400, "title required")
		return
	}
	if err := s.db.CreateTask(&t); err != nil {
		we(w, 500, "create failed")
		return
	}
	created := s.db.GetTask(t.ID)
	s.publishTask("task.created", created)
	wj(w, 201, created)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTask(r.PathValue("id"))
	if t == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, t)
}

// updateTask accepts a full or partial task. Empty string fields are
// preserved from existing. Tags=nil is treated as "not sent" (preserved).
// The tags field is special: clients that want to clear tags should send
// an empty array, not omit the field.
func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetTask(r.PathValue("id"))
	if existing == nil {
		we(w, 404, "not found")
		return
	}
	var t store.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if t.Title == "" {
		t.Title = existing.Title
	}
	if t.Description == "" {
		t.Description = existing.Description
	}
	if t.Status == "" {
		t.Status = existing.Status
	}
	if t.Priority == "" {
		t.Priority = existing.Priority
	}
	if t.Assignee == "" {
		t.Assignee = existing.Assignee
	}
	if t.DueDate == "" {
		t.DueDate = existing.DueDate
	}
	// project_id of "" preserves the existing project link. To explicitly
	// remove a task from its project, the API client should pass a sentinel
	// value (none currently supported) or use a separate endpoint. Empty
	// string is treated as "not sent" rather than "remove project".
	if t.ProjectID == "" {
		t.ProjectID = existing.ProjectID
	}
	if t.Tags == nil {
		t.Tags = existing.Tags
	}
	t.CompletedAt = existing.CompletedAt
	if err := s.db.UpdateTask(existing.ID, &t); err != nil {
		we(w, 500, "update failed")
		return
	}
	updated := s.db.GetTask(existing.ID)
	// Fire task.completed only on status transition into done.
	// Subscribers don't want to see the same completed event every
	// time an admin edits notes on an already-done task.
	if updated != nil && existing.Status != updated.Status &&
		strings.ToLower(updated.Status) == "done" {
		s.publishTask("task.completed", updated)
	}
	wj(w, 200, updated)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.db.DeleteTask(id)
	s.db.DeleteExtras(resourceTasks, id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

// setStatus is the kanban-style fast path for moving a task between columns.
func (s *Server) setStatus(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTask(r.PathValue("id"))
	if t == nil {
		we(w, 404, "not found")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if body.Status == "" {
		we(w, 400, "status required")
		return
	}
	prevStatus := t.Status
	t.Status = body.Status
	if err := s.db.UpdateTask(t.ID, t); err != nil {
		we(w, 500, "update failed")
		return
	}
	updated := s.db.GetTask(t.ID)
	// Kanban-style status flip: mirror the same transition firing
	// the full updateTask path uses. Both endpoints now emit
	// task.completed consistently when a task moves into "done".
	if updated != nil && prevStatus != updated.Status &&
		strings.ToLower(updated.Status) == "done" {
		s.publishTask("task.completed", updated)
	}
	wj(w, 200, updated)
}

// ─── stats / health ───────────────────────────────────────────────

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, s.db.Stats())
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	wj(w, 200, map[string]any{
		"service":  "roundup",
		"status":   "ok",
		"tasks":    st.Tasks,
		"projects": st.Projects,
	})
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// publishTask fires a task.* event on the bus. No-op when s.bus is
// nil. Runs in a goroutine; errors logged not surfaced. Payload
// shape locked by docs/BUS-TOPICS.md v1 in stockyard-desktop.
//
// Reality notes:
// - Roundup has no contact_id FK on tasks. Assignee is a free-text
//   field; subscribers wanting contact linkage must fuzzy-match.
// - Tags is a []string; payload forwards it as-is.
// - CompletedAt is store-populated when status flips to done.
func (s *Server) publishTask(topic string, t *store.Task) {
	if s.bus == nil || t == nil {
		return
	}
	payload := map[string]any{
		"task_id":      t.ID,
		"project_id":   t.ProjectID,
		"title":        t.Title,
		"description":  t.Description,
		"status":       t.Status,
		"priority":     t.Priority,
		"assignee":     t.Assignee,
		"due_date":     t.DueDate,
		"tags":         t.Tags,
		"completed_at": t.CompletedAt,
	}
	go func() {
		if _, err := s.bus.Publish(topic, payload); err != nil {
			log.Printf("roundup: bus publish %s failed: %v", topic, err)
		}
	}()
}
