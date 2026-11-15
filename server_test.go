package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"kadai/task"
)

func newTestServer() *server {
	due := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	return &server{
		store: &store{tasks: []task.Task{{Title: "レポート", Due: due}}},
		today: func() time.Time {
			return time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
		},
	}
}

func callAPI(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestList(t *testing.T) {
	s := newTestServer()
	rec := callAPI(t, s.routes(), "GET", "/tasks", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []taskView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].DaysLeft != 6 {
		t.Errorf("got %+v", got)
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"正しい", `{"title":"実験ノート","due":"2026-11-18"}`, http.StatusCreated},
		{"タイトルなし", `{"due":"2026-11-18"}`, http.StatusBadRequest},
		{"日付が変", `{"title":"実験ノート","due":"11/18"}`, http.StatusBadRequest},
		{"JSONでない", `実験ノート`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer()
			rec := callAPI(t, s.routes(), "POST", "/tasks", tt.body)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestCreateConcurrently(t *testing.T) {
	s := newTestServer()
	h := s.routes()
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			callAPI(t, h, "POST", "/tasks", `{"title":"演習","due":"2026-11-30"}`)
		})
	}
	wg.Wait()
	if n := len(s.store.all()); n != 51 {
		t.Errorf("件数 = %d, want 51", n)
	}
}
