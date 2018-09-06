package main

import "encoding/json"
import "flag"
import "log"
import "net/http"
import "strings"
import "time"

import "./bid"

// SourceTimeout is a timeout for sources,
// source is ignored after reaching a timeout
var SourceTimeout = 100 * time.Millisecond

func main() {
  addr := flag.String("addr", ":8080", "HTTP addr")
  flag.Parse()

  client := &http.Client{
    Timeout: SourceTimeout,
  }

  log.Printf("Listen on %s\n", *addr)
  log.Fatal(http.ListenAndServe(*addr, getWinnerHandler(client)))
}

func getWinnerHandler(client *http.Client) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    path := strings.Trim(r.URL.Path, "/")
    if path != "winner" {
      writeError(w, http.StatusNotFound, "Requested path not found")
      return
    }
    sources := r.URL.Query()["s"]
    if len(sources) == 0 {
      writeError(w, http.StatusBadRequest, "There must be at least one source")
      return
    }
    // TODO There should probably be referrer or/and sources validation
    // in order to prevent requests to unauthorized sources
    result := bid.Bid(client, sources)
    if !result.IsValid() {
      writeError(w, http.StatusUnprocessableEntity, "There is no prices")
      return
    }
    write(w, http.StatusOK, result)
  }
}

func write(w http.ResponseWriter, status int, data interface{}) {
  w.WriteHeader(status)
  w.Header().Set("Content-Type", "application/json")
  err := json.NewEncoder(w).Encode(data)
  if err != nil {
    log.Println(err)
  }
}

func writeError(w http.ResponseWriter, status int, message string) {
  data := map[string]string{
    "error": message,
  }
  write(w, status, data)
}
