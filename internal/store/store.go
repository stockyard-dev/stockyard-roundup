package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

// Project groups related tasks. Tasks may also exist with no project.
type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color,omitempty"`
	CreatedAt string `json:"created_at"`
	TaskCount int    `json:"task_count"`
	DoneCount int    `json:"done_count"`
}

// Task is the primary resource. Tags are stored as a JSON array column.
// CompletedAt is set automatically when status transitions to "done"
// and cleared when status transitions away from "done".
type Task struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	CompletedAt string   `json:"completed_at,omitempty"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "roundup.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS projects(
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#c45d2c',
			created_at TEXT DEFAULT(datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS tasks(
			id TEXT PRIMARY KEY,
			project_id TEXT DEFAULT '',
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			status TEXT DEFAULT 'todo',
			priority TEXT DEFAULT 'medium',
			assignee TEXT DEFAULT '',
			due_date TEXT DEFAULT '',
			tags_json TEXT DEFAULT '[]',
			created_at TEXT DEFAULT(datetime('now')),
			completed_at TEXT DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE TABLE IF NOT EXISTS extras(
			resource TEXT NOT NULL,
			record_id TEXT NOT NULL,
			data TEXT NOT NULL DEFAULT '{}',
			PRIMARY KEY(resource, record_id)
		)`,
	}
	for _, q := range migrations {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

// ─── projects ─────────────────────────────────────────────────────

func (d *DB) CreateProject(p *Project) error {
	p.ID = genID()
	p.CreatedAt = now()
	if p.Color == "" {
		p.Color = "#c45d2c"
	}
	_, err := d.db.Exec(
		`INSERT INTO projects(id, name, color, created_at) VALUES(?, ?, ?, ?)`,
		p.ID, p.Name, p.Color, p.CreatedAt,
	)
	return err
}

func (d *DB) GetProject(id string) *Project {
	var p Project
	err := d.db.QueryRow(
		`SELECT id, name, color, created_at FROM projects WHERE id=?`,
		id,
	).Scan(&p.ID, &p.Name, &p.Color, &p.CreatedAt)
	if err != nil {
		return nil
	}
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=?`, p.ID).Scan(&p.TaskCount)
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=? AND status='done'`, p.ID).Scan(&p.DoneCount)
	return &p
}

func (d *DB) ListProjects() []Project {
	rows, _ := d.db.Query(`SELECT id, name, color, created_at FROM projects ORDER BY name`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Name, &p.Color, &p.CreatedAt)
		d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=?`, p.ID).Scan(&p.TaskCount)
		d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=? AND status='done'`, p.ID).Scan(&p.DoneCount)
		o = append(o, p)
	}
	return o
}

func (d *DB) UpdateProject(p *Project) error {
	_, err := d.db.Exec(
		`UPDATE projects SET name=?, color=? WHERE id=?`,
		p.Name, p.Color, p.ID,
	)
	return err
}

// DeleteProject removes the project and all tasks belonging to it.
// Task extras are cleaned up by the server's delete handler before
// the project is deleted, so orphan extras don't accumulate.
func (d *DB) DeleteProject(id string) error {
	d.db.Exec(`DELETE FROM tasks WHERE project_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM projects WHERE id=?`, id)
	return err
}

// TasksInProject returns task IDs for a project. Used by the server's
// delete handler so it can cascade extras cleanup before the project goes.
func (d *DB) TasksInProject(projectID string) []string {
	rows, _ := d.db.Query(`SELECT id FROM tasks WHERE project_id=?`, projectID)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids
}

// ─── tasks ────────────────────────────────────────────────────────

func (d *DB) CreateTask(t *Task) error {
	t.ID = genID()
	t.CreatedAt = now()
	if t.Status == "" {
		t.Status = "todo"
	}
	if t.Priority == "" {
		t.Priority = "medium"
	}
	if t.Tags == nil {
		t.Tags = []string{}
	}
	if t.Status == "done" && t.CompletedAt == "" {
		t.CompletedAt = now()
	}
	tj, _ := json.Marshal(t.Tags)
	_, err := d.db.Exec(
		`INSERT INTO tasks(id, project_id, title, description, status, priority, assignee, due_date, tags_json, created_at, completed_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.Assignee, t.DueDate, string(tj), t.CreatedAt, t.CompletedAt,
	)
	return err
}

// scanTask decodes a row into a Task. Accepts either *sql.Row or *sql.Rows.
func (d *DB) scanTask(sc interface {
	Scan(...any) error
}) *Task {
	var t Task
	var tj string
	if sc.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &t.DueDate, &tj, &t.CreatedAt, &t.CompletedAt) != nil {
		return nil
	}
	json.Unmarshal([]byte(tj), &t.Tags)
	if t.Tags == nil {
		t.Tags = []string{}
	}
	return &t
}

func (d *DB) GetTask(id string) *Task {
	return d.scanTask(d.db.QueryRow(
		`SELECT id, project_id, title, description, status, priority, assignee, due_date, tags_json, created_at, completed_at
		 FROM tasks WHERE id=?`,
		id,
	))
}

// ListTasks returns tasks matching optional project, status, and priority
// filters. Tasks are ordered by priority (urgent → low) then by creation time.
func (d *DB) ListTasks(projectID, status, priority string) []Task {
	where := []string{"1=1"}
	args := []any{}
	if projectID != "" {
		where = append(where, "project_id=?")
		args = append(args, projectID)
	}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if priority != "" {
		where = append(where, "priority=?")
		args = append(args, priority)
	}
	rows, _ := d.db.Query(
		`SELECT id, project_id, title, description, status, priority, assignee, due_date, tags_json, created_at, completed_at
		 FROM tasks WHERE `+strings.Join(where, " AND ")+`
		 ORDER BY CASE priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, created_at DESC`,
		args...,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Task
	for rows.Next() {
		if t := d.scanTask(rows); t != nil {
			o = append(o, *t)
		}
	}
	return o
}

// UpdateTask writes the patched task. Auto-manages completed_at:
//   - status -> done sets completed_at to now (if not already set)
//   - status -> anything else clears completed_at
func (d *DB) UpdateTask(id string, t *Task) error {
	tj, _ := json.Marshal(t.Tags)
	completed := t.CompletedAt
	if t.Status == "done" && completed == "" {
		completed = now()
	}
	if t.Status != "done" {
		completed = ""
	}
	_, err := d.db.Exec(
		`UPDATE tasks SET project_id=?, title=?, description=?, status=?, priority=?, assignee=?, due_date=?, tags_json=?, completed_at=?
		 WHERE id=?`,
		t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.Assignee, t.DueDate, string(tj), completed, id,
	)
	return err
}

func (d *DB) DeleteTask(id string) error {
	_, err := d.db.Exec(`DELETE FROM tasks WHERE id=?`, id)
	return err
}

// ─── stats ────────────────────────────────────────────────────────

type Stats struct {
	Tasks      int `json:"tasks"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
	Projects   int `json:"projects"`
	Overdue    int `json:"overdue"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&s.Tasks)
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='todo'`).Scan(&s.Todo)
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='in_progress'`).Scan(&s.InProgress)
	d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='done'`).Scan(&s.Done)
	d.db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&s.Projects)
	// Overdue = due_date in the past and not yet done
	today := time.Now().UTC().Format("2006-01-02")
	d.db.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE due_date != '' AND due_date < ? AND status != 'done'`,
		today,
	).Scan(&s.Overdue)
	return s
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
