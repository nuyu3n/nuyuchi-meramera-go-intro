package task

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type Task struct {
	Title string
	Due   time.Time
	Done  bool
}

func (t Task) DaysLeft(today time.Time) int {
	return int(t.Due.Sub(today).Hours() / 24)
}

func (t Task) Status(today time.Time) string {
	d := t.DaysLeft(today)
	switch {
	case d < 0:
		return "期限切れ"
	case d == 0:
		return "今日まで"
	case d <= 3:
		return "もうすぐ"
	default:
		return "まだ余裕あり"
	}
}

type Filter interface {
	Keep(t Task) bool
}

type DueWithin struct {
	Today time.Time
	Days  int
}

func (f DueWithin) Keep(t Task) bool {
	d := t.DaysLeft(f.Today)
	return d >= 0 && d <= f.Days
}

func Apply(tasks []Task, f Filter) []Task {
	var out []Task
	for _, t := range tasks {
		if f.Keep(t) {
			out = append(out, t)
		}
	}
	return out
}

func New(title, dueText string) (Task, error) {
	if title == "" {
		return Task{}, errors.New("タイトルが空です")
	}
	due, err := time.Parse(dateLayout, dueText)
	if err != nil {
		return Task{}, fmt.Errorf("締切日が読めません: %w", err)
	}
	return Task{Title: title, Due: due}, nil
}

func ParseLine(line string) (Task, error) {
	title, dueText, found := strings.Cut(line, ",")
	if !found {
		return Task{}, errors.New("カンマがありません")
	}
	return New(title, dueText)
}

func Read(r io.Reader) ([]Task, error) {
	var tasks []Task
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		t, err := ParseLine(sc.Text())
		if err != nil {
			return nil, fmt.Errorf("%d行目: %w", line, err)
		}
		tasks = append(tasks, t)
	}
	return tasks, sc.Err()
}

func Print(w io.Writer, tasks []Task, today time.Time) {
	for _, t := range tasks {
		fmt.Fprintf(w, "%s: あと%d日（%s）\n", shorten(t.Title, 8), t.DaysLeft(today), t.Status(today))
	}
}

func shorten(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
