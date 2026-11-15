package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"kadai/task"
)

type taskView struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Due      string `json:"due"`
	DaysLeft int    `json:"days_left"`
	Status   string `json:"status"`
	Done     bool   `json:"done"`
}

func newTaskView(id int, t task.Task, today time.Time) taskView {
	return taskView{
		ID:       id,
		Title:    t.Title,
		Due:      t.Due.Format("2006-01-02"),
		DaysLeft: t.DaysLeft(today),
		Status:   t.Status(today),
		Done:     t.Done,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("JSONの書き出しに失敗:", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

type server struct {
	store *store
	today func() time.Time
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", s.handleList)
	mux.HandleFunc("POST /tasks", s.handleCreate)
	mux.HandleFunc("POST /tasks/{id}/done", s.handleDone)
	return mux
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	today := s.today()
	pendingOnly := r.URL.Query().Get("pending") == "true"
	views := []taskView{}
	for i, t := range s.store.all() {
		if pendingOnly && t.Done {
			continue
		}
		views = append(views, newTaskView(i+1, t, today))
	}
	writeJSON(w, http.StatusOK, views)
}

type createRequest struct {
	Title string `json:"title"`
	Due   string `json:"due"`
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSONが読めません")
		return
	}
	t, err := task.New(req.Title, req.Due)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := s.store.add(t)
	writeJSON(w, http.StatusCreated, newTaskView(id, t, s.today()))
}

func (s *server) handleDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "課題が見つかりません")
		return
	}
	t, ok := s.store.finish(id)
	if !ok {
		writeError(w, http.StatusNotFound, "課題が見つかりません")
		return
	}
	writeJSON(w, http.StatusOK, newTaskView(id, t, s.today()))
}
