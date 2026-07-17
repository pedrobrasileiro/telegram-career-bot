package telegram

import (
	"context"
	"sync"
	"time"

	"telegram-career-bot/internal/domain"
)

type taskEntry struct {
	task   *domain.Task
	cancel context.CancelFunc
}

// TaskManager rastreia o job em execução por chat (no máximo um job
// concorrente por chat) — estado de sessão do Telegram, não domínio.
// Guarda também o cancel func do contexto do job, pra suportar /cancel.
type TaskManager struct {
	mu    sync.RWMutex
	tasks map[int64]*taskEntry
}

func NewTaskManager() *TaskManager {
	return &TaskManager{tasks: make(map[int64]*taskEntry)}
}

func (tm *TaskManager) IsBusy(chatID int64) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	_, ok := tm.tasks[chatID]
	return ok
}

func (tm *TaskManager) Start(chatID int64, taskType, desc string, cancel context.CancelFunc) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks[chatID] = &taskEntry{
		task: &domain.Task{
			Type:        taskType,
			Description: desc,
			StartTime:   time.Now(),
		},
		cancel: cancel,
	}
}

func (tm *TaskManager) End(chatID int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tasks, chatID)
}

func (tm *TaskManager) Get(chatID int64) *domain.Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	entry, ok := tm.tasks[chatID]
	if !ok {
		return nil
	}
	return entry.task
}

// Cancel dispara o cancel func do job em execução pro chat. Retorna false
// se não havia job rodando. Não remove a entrada — quem roda o job é
// responsável por chamar End() quando notar o cancelamento.
func (tm *TaskManager) Cancel(chatID int64) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	entry, ok := tm.tasks[chatID]
	if !ok {
		return false
	}
	entry.cancel()
	return true
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
