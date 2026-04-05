package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-roundup/internal/server";"github.com/stockyard-dev/stockyard-roundup/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9080"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./roundup-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("roundup: %v",err)};defer db.Close();srv:=server.New(db,server.DefaultLimits())
fmt.Printf("\n  Roundup — Self-hosted task manager\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Data:       %s\n  ─────────────────────────────────\n  Questions? hello@stockyard.dev\n\n",port,port,dataDir)
log.Printf("roundup: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
