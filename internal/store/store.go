package store
import ("database/sql";"encoding/json";"fmt";"os";"path/filepath";"strings";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Project struct{ID string `json:"id"`;Name string `json:"name"`;Color string `json:"color,omitempty"`;CreatedAt string `json:"created_at"`;TaskCount int `json:"task_count"`;DoneCount int `json:"done_count"`}
type Task struct{ID string `json:"id"`;ProjectID string `json:"project_id,omitempty"`;Title string `json:"title"`;Description string `json:"description,omitempty"`;Status string `json:"status"`;Priority string `json:"priority"`;Assignee string `json:"assignee,omitempty"`;DueDate string `json:"due_date,omitempty"`;Tags []string `json:"tags"`;CreatedAt string `json:"created_at"`;CompletedAt string `json:"completed_at,omitempty"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"roundup.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
for _,q:=range[]string{
`CREATE TABLE IF NOT EXISTS projects(id TEXT PRIMARY KEY,name TEXT NOT NULL,color TEXT DEFAULT '#c45d2c',created_at TEXT DEFAULT(datetime('now')))`,
`CREATE TABLE IF NOT EXISTS tasks(id TEXT PRIMARY KEY,project_id TEXT DEFAULT '',title TEXT NOT NULL,description TEXT DEFAULT '',status TEXT DEFAULT 'todo',priority TEXT DEFAULT 'medium',assignee TEXT DEFAULT '',due_date TEXT DEFAULT '',tags_json TEXT DEFAULT '[]',created_at TEXT DEFAULT(datetime('now')),completed_at TEXT DEFAULT '')`,
`CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id)`,
`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
}{if _,err:=db.Exec(q);err!=nil{return nil,fmt.Errorf("migrate: %w",err)}};return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)CreateProject(p *Project)error{p.ID=genID();p.CreatedAt=now();if p.Color==""{p.Color="#c45d2c"};_,err:=d.db.Exec(`INSERT INTO projects VALUES(?,?,?,?)`,p.ID,p.Name,p.Color,p.CreatedAt);return err}
func(d *DB)ListProjects()[]Project{rows,_:=d.db.Query(`SELECT id,name,color,created_at FROM projects ORDER BY name`);if rows==nil{return nil};defer rows.Close()
var o []Project;for rows.Next(){var p Project;rows.Scan(&p.ID,&p.Name,&p.Color,&p.CreatedAt);d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=?`,p.ID).Scan(&p.TaskCount);d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id=? AND status='done'`,p.ID).Scan(&p.DoneCount);o=append(o,p)};return o}
func(d *DB)DeleteProject(id string)error{d.db.Exec(`DELETE FROM tasks WHERE project_id=?`,id);_,err:=d.db.Exec(`DELETE FROM projects WHERE id=?`,id);return err}
func(d *DB)CreateTask(t *Task)error{t.ID=genID();t.CreatedAt=now();if t.Status==""{t.Status="todo"};if t.Priority==""{t.Priority="medium"};if t.Tags==nil{t.Tags=[]string{}}
tj,_:=json.Marshal(t.Tags);_,err:=d.db.Exec(`INSERT INTO tasks(id,project_id,title,description,status,priority,assignee,due_date,tags_json,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,t.ID,t.ProjectID,t.Title,t.Description,t.Status,t.Priority,t.Assignee,t.DueDate,string(tj),t.CreatedAt);return err}
func(d *DB)scanTask(sc interface{Scan(...any)error})*Task{var t Task;var tj string
if sc.Scan(&t.ID,&t.ProjectID,&t.Title,&t.Description,&t.Status,&t.Priority,&t.Assignee,&t.DueDate,&tj,&t.CreatedAt,&t.CompletedAt)!=nil{return nil}
json.Unmarshal([]byte(tj),&t.Tags);if t.Tags==nil{t.Tags=[]string{}};return &t}
func(d *DB)GetTask(id string)*Task{return d.scanTask(d.db.QueryRow(`SELECT id,project_id,title,description,status,priority,assignee,due_date,tags_json,created_at,completed_at FROM tasks WHERE id=?`,id))}
func(d *DB)ListTasks(projectID,status,priority string)[]Task{where:=[]string{"1=1"};args:=[]any{}
if projectID!=""{where=append(where,"project_id=?");args=append(args,projectID)}
if status!=""&&status!="all"{where=append(where,"status=?");args=append(args,status)}
if priority!=""{where=append(where,"priority=?");args=append(args,priority)}
rows,_:=d.db.Query(`SELECT id,project_id,title,description,status,priority,assignee,due_date,tags_json,created_at,completed_at FROM tasks WHERE `+strings.Join(where," AND ")+` ORDER BY CASE priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, created_at DESC`,args...)
if rows==nil{return nil};defer rows.Close();var o []Task;for rows.Next(){if t:=d.scanTask(rows);t!=nil{o=append(o,*t)}};return o}
func(d *DB)UpdateTask(id string,t *Task)error{tj,_:=json.Marshal(t.Tags);ca:=t.CompletedAt;if t.Status=="done"&&ca==""{ca=now()};if t.Status!="done"{ca=""}
_,err:=d.db.Exec(`UPDATE tasks SET title=?,description=?,status=?,priority=?,assignee=?,due_date=?,tags_json=?,completed_at=? WHERE id=?`,t.Title,t.Description,t.Status,t.Priority,t.Assignee,t.DueDate,string(tj),ca,id);return err}
func(d *DB)DeleteTask(id string)error{_,err:=d.db.Exec(`DELETE FROM tasks WHERE id=?`,id);return err}
type Stats struct{Tasks int `json:"tasks"`;Todo int `json:"todo"`;InProgress int `json:"in_progress"`;Done int `json:"done"`;Projects int `json:"projects"`}
func(d *DB)Stats()Stats{var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&s.Tasks);d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='todo'`).Scan(&s.Todo);d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='in_progress'`).Scan(&s.InProgress);d.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE status='done'`).Scan(&s.Done);d.db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&s.Projects);return s}
