package server

import (
	"encoding/json"
	"net/http"

	"github.com/stockyard-dev/stockyard-roundup/internal/store"
)

type Server struct{ db *store.DB; mux *http.ServeMux; limits Limits }

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}

	s.mux.HandleFunc("GET /api/projects", s.listProjects)
	s.mux.HandleFunc("POST /api/projects", s.createProject)
	s.mux.HandleFunc("DELETE /api/projects/{id}", s.deleteProject)

	s.mux.HandleFunc("GET /api/tasks", s.listTasks)
	s.mux.HandleFunc("POST /api/tasks", s.createTask)
	s.mux.HandleFunc("GET /api/tasks/{id}", s.getTask)
	s.mux.HandleFunc("PUT /api/tasks/{id}", s.updateTask)
	s.mux.HandleFunc("DELETE /api/tasks/{id}", s.deleteTask)
	s.mux.HandleFunc("PATCH /api/tasks/{id}/status", s.setStatus)

	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/roundup/"})
	})
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func ot(t []store.Task) []store.Task { if t == nil { return []store.Task{} }; return t }
func op(p []store.Project) []store.Project { if p == nil { return []store.Project{} }; return p }

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) { wj(w, 200, map[string]any{"projects": op(s.db.ListProjects())}) }
func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var p store.Project; json.NewDecoder(r.Body).Decode(&p); if p.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateProject(&p); wj(w, 201, map[string]string{"id": p.ID, "status": "created"})
}
func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) { s.db.DeleteProject(r.PathValue("id")); wj(w, 200, map[string]string{"status": "deleted"}) }

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"tasks": ot(s.db.ListTasks(r.URL.Query().Get("project"), r.URL.Query().Get("status"), r.URL.Query().Get("priority")))})
}
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	if s.limits.MaxItems > 0 && len(s.db.ListTasks("", "", "")) >= s.limits.MaxItems { we(w, 402, "Free tier limit reached"); return }
	var t store.Task; json.NewDecoder(r.Body).Decode(&t); if t.Title == "" { we(w, 400, "title required"); return }
	s.db.CreateTask(&t); wj(w, 201, s.db.GetTask(t.ID))
}
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTask(r.PathValue("id")); if t == nil { we(w, 404, "not found"); return }; wj(w, 200, t)
}
func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetTask(r.PathValue("id")); if existing == nil { we(w, 404, "not found"); return }
	var t store.Task; json.NewDecoder(r.Body).Decode(&t); if t.Title == "" { t.Title = existing.Title }
	if t.Status == "" { t.Status = existing.Status }; if t.Priority == "" { t.Priority = existing.Priority }
	if t.Tags == nil { t.Tags = existing.Tags }; t.CompletedAt = existing.CompletedAt
	s.db.UpdateTask(existing.ID, &t); wj(w, 200, s.db.GetTask(existing.ID))
}
func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) { s.db.DeleteTask(r.PathValue("id")); wj(w, 200, map[string]string{"status": "deleted"}) }
func (s *Server) setStatus(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTask(r.PathValue("id")); if t == nil { we(w, 404, "not found"); return }
	var body struct { Status string `json:"status"` }; json.NewDecoder(r.Body).Decode(&body)
	t.Status = body.Status; s.db.UpdateTask(t.ID, t); wj(w, 200, s.db.GetTask(t.ID))
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) { st := s.db.Stats(); wj(w, 200, map[string]any{"service": "roundup", "status": "ok", "tasks": st.Tasks, "projects": st.Projects}) }
