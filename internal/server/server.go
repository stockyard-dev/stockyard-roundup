package server

import (
	"context"
	"encoding/json"
	"fmt"
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

	s.subscribeBus()
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

// subscribeBus wires cross-tool events to auto-drafted tasks.
// No-op when s.bus is nil (standalone mode).
//
// Allowlist-only (not SubscribeAll) so future topics don't silently
// start creating follow-up tasks. Expanding this list is a
// PR-reviewed change — every addition is "a new way tasks appear in
// the user's queue without them clicking New Task."
//
// Idempotency: the [invoice:<id>] marker is embedded in the task
// Description and scanned on each fire. Bus cursor initializes at the
// current high-water mark on Open (bus.go:199), so process restart
// does NOT replay old events — we only dedup in-process duplicate
// fires during a single bundle lifetime.
//
// Handlers return nil on decode/data errors. The bus has no automatic
// retry (bus.go Handler docstring).
func (s *Server) subscribeBus() {
	if s.bus == nil {
		return
	}
	s.bus.Subscribe("invoice.overdue", func(_ context.Context, e bus.Event) error {
		return s.handleInvoiceOverdue(e)
	})
	log.Printf("roundup: subscribed to invoice.overdue")
}

// handleInvoiceOverdue creates a follow-up task when billfold flags
// an invoice overdue.
//
// Shape decisions (see BUS-TOPICS.md):
//   - Title = "Follow up on overdue invoice: <client_name>". Short
//     enough for the task list UI, specific enough to scan.
//   - Description = human-readable summary (invoice id, client,
//     amount, due date) with a [invoice:<id>] idempotency marker
//     appended. The marker is plain text in the body because
//     roundup's store doesn't have a dedicated source field and
//     Description is the natural place for provenance — a human
//     reading the task sees why it exists.
//   - Status = "todo". Enters the standard kanban intake column.
//   - Priority = "high". Overdue money is always the top of the
//     stack; user can demote in the UI if they disagree.
//   - Assignee = "" (user assigns manually — we have no mapping from
//     invoice → responsible person).
//   - DueDate = "" (no policy to infer; user sets when they triage).
//   - Tags = ["overdue", "billing"]. Conventional enough for filter.
//   - ProjectID = "" (lives in the default bucket).
func (s *Server) handleInvoiceOverdue(e bus.Event) error {
	var p map[string]any
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		log.Printf("roundup: decode invoice.overdue: %v", err)
		return nil
	}
	invoiceID := stringField(p, "invoice_id")
	if invoiceID == "" {
		log.Printf("roundup: invoice.overdue missing invoice_id, skipping")
		return nil
	}
	marker := fmt.Sprintf("[invoice:%s]", invoiceID)
	// Idempotency: scan existing tasks for this marker in the
	// description. ListTasks with all-empty filters returns every task.
	for _, existing := range s.db.ListTasks("", "", "") {
		if strings.Contains(existing.Description, marker) {
			log.Printf("roundup: invoice %s already has follow-up task %s, skipping",
				invoiceID, existing.ID)
			return nil
		}
	}
	clientName := stringField(p, "client_name")
	if clientName == "" {
		clientName = "(unknown client)"
	}
	amount := intField(p, "amount")
	dueDate := stringField(p, "due_date")
	title := fmt.Sprintf("Follow up on overdue invoice: %s", clientName)
	desc := fmt.Sprintf(
		"Invoice %s for %s (amount %d) is overdue",
		invoiceID, clientName, amount,
	)
	if dueDate != "" {
		desc += " since " + dueDate
	}
	desc += ". " + marker
	task := store.Task{
		Title:       title,
		Description: desc,
		Status:      "todo",
		Priority:    "high",
		Tags:        []string{"overdue", "billing"},
	}
	if err := s.db.CreateTask(&task); err != nil {
		log.Printf("roundup: create task for invoice %s: %v", invoiceID, err)
		return nil
	}
	log.Printf("roundup: auto-drafted task %s for overdue invoice %s (client=%q)",
		task.ID, invoiceID, clientName)
	return nil
}

// stringField returns m[k] as a string, or "" if absent / wrong type.
func stringField(m map[string]any, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}

// intField returns m[k] as an int, truncating float64 JSON numbers.
func intField(m map[string]any, k string) int {
	switch v := m[k].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	}
	return 0
}
