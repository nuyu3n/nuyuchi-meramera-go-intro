package task

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestDaysLeft(t *testing.T) {
	tk := Task{Title: "レポート", Due: date(2026, 11, 20)}
	got := tk.DaysLeft(date(2026, 11, 14))
	if got != 6 {
		t.Errorf("DaysLeft() = %d, want 6", got)
	}
}

func TestStatus(t *testing.T) {
	today := date(2026, 11, 14)
	tests := []struct {
		name string
		due  time.Time
		want string
	}{
		{"過ぎている", date(2026, 11, 13), "期限切れ"},
		{"今日", date(2026, 11, 14), "今日まで"},
		{"3日後", date(2026, 11, 17), "もうすぐ"},
		{"4日後", date(2026, 11, 18), "まだ余裕あり"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Task{Due: tt.due}.Status(today)
			if got != tt.want {
				t.Errorf("Status() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Task
		wantErr bool
	}{
		{"正しい行", "レポート,2026-11-20", Task{Title: "レポート", Due: date(2026, 11, 20)}, false},
		{"カンマなし", "レポート 2026-11-20", Task{}, true},
		{"タイトルが空", ",2026-11-20", Task{}, true},
		{"日付が変", "レポート,11/20", Task{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseLine(%q) err = %v, wantErr %v", tt.line, err, tt.wantErr)
			}
			if got.Title != tt.want.Title || !got.Due.Equal(tt.want.Due) {
				t.Errorf("ParseLine(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}

func readString(t *testing.T, s string) []Task {
	t.Helper()
	tasks, err := Read(strings.NewReader(s))
	if err != nil {
		t.Fatalf("Read(%q) でエラー: %v", s, err)
	}
	return tasks
}

func TestRead(t *testing.T) {
	got := readString(t, "レポート,2026-11-20\n小テスト勉強,2026-11-15\n")
	want := []Task{
		{Title: "レポート", Due: date(2026, 11, 20)},
		{Title: "小テスト勉強", Due: date(2026, 11, 15)},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Read() の結果が違います (-want +got):\n%s", diff)
	}
}

func TestPrint(t *testing.T) {
	tasks := readString(t, "レポート,2026-11-20\n小テスト勉強,2026-11-15\n")

	var buf bytes.Buffer
	Print(&buf, tasks, date(2026, 11, 14))

	want := "レポート: あと6日（まだ余裕あり）\n小テスト勉強: あと1日（もうすぐ）\n"
	if buf.String() != want {
		t.Errorf("got\n%s\nwant\n%s", buf.String(), want)
	}
}
