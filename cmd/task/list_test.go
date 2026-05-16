package task

import (
	"testing"
	"time"

	"github.com/botre/tickli/internal/types"
	"github.com/botre/tickli/internal/types/task"
)

func taskIDs(tasks []types.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}

func TestFilterTasksCompletionStatus(t *testing.T) {
	tasks := []types.Task{
		{ID: "open", Status: task.StatusNormal},
		{ID: "done", Status: task.StatusComplete},
	}

	got := filterTasks(tasks, &listOptions{})
	if len(got) != 1 || got[0].ID != "open" {
		t.Errorf("default filter = %v, want [open]", taskIDs(got))
	}

	got = filterTasks(tasks, &listOptions{all: true})
	if len(got) != 2 {
		t.Errorf("--all filter = %v, want both tasks", taskIDs(got))
	}
}

func TestFilterTasksPriority(t *testing.T) {
	tasks := []types.Task{
		{ID: "low", Priority: task.PriorityLow},
		{ID: "high", Priority: task.PriorityHigh},
	}

	got := filterTasks(tasks, &listOptions{priority: task.PriorityMedium})
	if len(got) != 1 || got[0].ID != "high" {
		t.Errorf("priority filter = %v, want [high]", taskIDs(got))
	}
}

func TestFilterTasksTag(t *testing.T) {
	tasks := []types.Task{
		{ID: "a", Tags: []string{"work"}},
		{ID: "b", Tags: []string{"home"}},
	}

	got := filterTasks(tasks, &listOptions{tag: "work"})
	if len(got) != 1 || got[0].ID != "a" {
		t.Errorf("tag filter = %v, want [a]", taskIDs(got))
	}
}

func TestFilterTasksDueDate(t *testing.T) {
	now := time.Now()
	todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	tasks := []types.Task{
		{ID: "today", DueDate: types.TickTickTime(todayNoon)},
		{ID: "next-week", DueDate: types.TickTickTime(todayNoon.AddDate(0, 0, 10))},
		{ID: "no-date"},
	}

	got := filterTasks(tasks, &listOptions{dueDate: "today"})
	if len(got) != 1 || got[0].ID != "today" {
		t.Errorf("due-within today = %v, want [today]", taskIDs(got))
	}
}
