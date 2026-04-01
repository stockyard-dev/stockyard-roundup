package server
import("encoding/json";"net/http";"github.com/stockyard-dev/stockyard-roundup/internal/store")
type Server struct{db *store.DB;limits Limits;mux *http.ServeMux}
func New(db *store.DB,tier string)*Server{s:=&Server{db:db,limits:LimitsFor(tier),mux:http.NewServeMux()};s.routes();return s}
func(s *Server)ListenAndServe(addr string)error{return(&http.Server{Addr:addr,Handler:s.mux}).ListenAndServe()}
func(s *Server)routes(){
    s.mux.HandleFunc("GET /health",s.handleHealth)
    s.mux.HandleFunc("GET /api/stats",s.handleStats)
    s.mux.HandleFunc("GET /api/meetings",s.handleListMeetings)
    s.mux.HandleFunc("POST /api/meetings",s.handleCreateMeeting)
    s.mux.HandleFunc("GET /api/meetings/{id}",s.handleGetMeeting)
    s.mux.HandleFunc("PATCH /api/meetings/{id}/notes",s.handleUpdateNotes)
    s.mux.HandleFunc("DELETE /api/meetings/{id}",s.handleDeleteMeeting)
    s.mux.HandleFunc("GET /api/actions",s.handleListActions)
    s.mux.HandleFunc("POST /api/actions",s.handleCreateAction)
    s.mux.HandleFunc("POST /api/actions/{id}/toggle",s.handleToggleAction)
    s.mux.HandleFunc("DELETE /api/actions/{id}",s.handleDeleteAction)
    s.mux.HandleFunc("GET /",s.handleUI)
}
func(s *Server)handleHealth(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]string{"status":"ok","service":"stockyard-roundup"})}
func writeJSON(w http.ResponseWriter,status int,v interface{}){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);json.NewEncoder(w).Encode(v)}
func writeError(w http.ResponseWriter,status int,msg string){writeJSON(w,status,map[string]string{"error":msg})}
func(s *Server)handleUI(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html");w.Write(dashboardHTML)}
