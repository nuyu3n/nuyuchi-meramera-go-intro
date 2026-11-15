package main

import (
	"slices"
	"sync"

	"kadai/task"
)

type store struct {
	mu    sync.Mutex
	tasks []task.Task
}

func (s *store) all() []task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.tasks)
}

func (s *store) add(t task.Task) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, t)
	return len(s.tasks)
}

func (s *store) finish(id int) (task.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id < 1 || id > len(s.tasks) {
		return task.Task{}, false
	}
	s.tasks[id-1].Done = true
	return s.tasks[id-1], true
}
