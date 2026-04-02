package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-roundup/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux}
func New(db *store.DB)*Server{s:=&Server{db:db,mux:http.NewServeMux()}
s.mux.HandleFunc("GET /api/projects",s.listProjects);s.mux.HandleFunc("POST /api/projects",s.createProject);s.mux.HandleFunc("DELETE /api/projects/{id}",s.deleteProject)
s.mux.HandleFunc("GET /api/tasks",s.listTasks);s.mux.HandleFunc("POST /api/tasks",s.createTask);s.mux.HandleFunc("GET /api/tasks/{id}",s.getTask);s.mux.HandleFunc("PUT /api/tasks/{id}",s.updateTask);s.mux.HandleFunc("DELETE /api/tasks/{id}",s.deleteTask)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listProjects(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"projects":oe(s.db.ListProjects())})}
func(s *Server)createProject(w http.ResponseWriter,r *http.Request){var p store.Project;json.NewDecoder(r.Body).Decode(&p);if p.Name==""{we(w,400,"name required");return};s.db.CreateProject(&p);wj(w,201,p)}
func(s *Server)deleteProject(w http.ResponseWriter,r *http.Request){s.db.DeleteProject(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)listTasks(w http.ResponseWriter,r *http.Request){q:=r.URL.Query();wj(w,200,map[string]any{"tasks":oe(s.db.ListTasks(q.Get("project_id"),q.Get("status"),q.Get("priority")))})}
func(s *Server)createTask(w http.ResponseWriter,r *http.Request){var t store.Task;json.NewDecoder(r.Body).Decode(&t);if t.Title==""{we(w,400,"title required");return};s.db.CreateTask(&t);wj(w,201,s.db.GetTask(t.ID))}
func(s *Server)getTask(w http.ResponseWriter,r *http.Request){t:=s.db.GetTask(r.PathValue("id"));if t==nil{we(w,404,"not found");return};wj(w,200,t)}
func(s *Server)updateTask(w http.ResponseWriter,r *http.Request){id:=r.PathValue("id");ex:=s.db.GetTask(id);if ex==nil{we(w,404,"not found");return};var t store.Task;json.NewDecoder(r.Body).Decode(&t);if t.Title==""{t.Title=ex.Title};if t.Status==""{t.Status=ex.Status};if t.Priority==""{t.Priority=ex.Priority};if t.Tags==nil{t.Tags=ex.Tags};s.db.UpdateTask(id,&t);wj(w,200,s.db.GetTask(id))}
func(s *Server)deleteTask(w http.ResponseWriter,r *http.Request){s.db.DeleteTask(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"roundup","tasks":st.Tasks,"todo":st.Todo})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
