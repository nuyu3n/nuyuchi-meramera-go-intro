package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"kadai/task"
)

func loadTasks(path string) ([]task.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return task.Read(f)
}

func currentDate() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func list(w io.Writer, tasks []task.Task, today time.Time) {
	sorted := slices.Clone(tasks)
	slices.SortFunc(sorted, func(a, b task.Task) int {
		return a.Due.Compare(b.Due)
	})
	task.Print(w, sorted, today)
}

func summary(w io.Writer, tasks []task.Task, today time.Time) {
	counts := map[string]int{}
	for _, t := range tasks {
		counts[t.Status(today)]++
	}
	for _, s := range []string{"期限切れ", "今日まで", "もうすぐ", "まだ余裕あり"} {
		fmt.Fprintf(w, "%s: %d件\n", s, counts[s])
	}
}

func soon(w io.Writer, tasks []task.Task, today time.Time) {
	list(w, task.Apply(tasks, task.DueWithin{Today: today, Days: 3}), today)
}

func sendReminder(ctx context.Context, t task.Task, today time.Time) (string, error) {
	select {
	case <-time.After(500 * time.Millisecond):
		return fmt.Sprintf("通知: %s（あと%d日）", t.Title, t.DaysLeft(today)), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func remind(w io.Writer, tasks []task.Task, today time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	sent := 0
	for _, t := range tasks {
		wg.Go(func() {
			msg, err := sendReminder(ctx, t, today)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Fprintf(w, "送れませんでした: %s（%v）\n", t.Title, err)
				return
			}
			fmt.Fprintln(w, msg)
			sent++
		})
	}
	wg.Wait()
	fmt.Fprintf(w, "%d件送りました\n", sent)
}

func serve(_ io.Writer, tasks []task.Task, _ time.Time) {
	s := &server{
		store: &store{tasks: tasks},
		today: currentDate,
	}

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("http://localhost%s で待っています", addr)
	log.Fatal(http.ListenAndServe(addr, logRequests(s.routes())))
}

func main() {
	today := currentDate()

	tasks, err := loadTasks("tasks.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(1)
	}

	commands := map[string]func(w io.Writer, tasks []task.Task, today time.Time){
		"list":    list,
		"summary": summary,
		"soon":    soon,
		"remind":  remind,
		"serve":   serve,
	}

	name := "list"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintln(os.Stderr, "知らないコマンドです:", name)
		os.Exit(1)
	}
	cmd(os.Stdout, tasks, today)
}
