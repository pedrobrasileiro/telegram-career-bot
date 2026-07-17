package main

import (
	"sync"
	"time"
)

type Task struct {
	Type        string
	Description string
	StartTime   time.Time
}

type TaskManager struct {
	mu    sync.RWMutex
	tasks map[int64]*Task
}

func NewTaskManager() *TaskManager {
	return &TaskManager{tasks: make(map[int64]*Task)}
}

func (tm *TaskManager) IsBusy(chatID int64) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	_, ok := tm.tasks[chatID]
	return ok
}

func (tm *TaskManager) Start(chatID int64, taskType, desc string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks[chatID] = &Task{
		Type:        taskType,
		Description: desc,
		StartTime:   time.Now(),
	}
}

func (tm *TaskManager) End(chatID int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tasks, chatID)
}

func (tm *TaskManager) Get(chatID int64) *Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[chatID]
}

// AwaitingInput rastreia comandos que pediram um dado extra ao usuário
// (ex: /report sem número) e esperam a próxima mensagem de texto como resposta.
type AwaitingInput struct {
	mu    sync.RWMutex
	chats map[int64]string
}

func NewAwaitingInput() *AwaitingInput {
	return &AwaitingInput{chats: make(map[int64]string)}
}

func (a *AwaitingInput) Set(chatID int64, kind string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.chats[chatID] = kind
}

func (a *AwaitingInput) Take(chatID int64) (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	kind, ok := a.chats[chatID]
	if ok {
		delete(a.chats, chatID)
	}
	return kind, ok
}
