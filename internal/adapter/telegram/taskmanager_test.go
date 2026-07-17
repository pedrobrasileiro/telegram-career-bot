package telegram

import "testing"

func TestTaskManagerLifecycle(t *testing.T) {
	tm := NewTaskManager()
	chatID := int64(42)

	if tm.IsBusy(chatID) {
		t.Fatal("não deveria estar ocupado antes de Start")
	}
	if tm.Get(chatID) != nil {
		t.Fatal("Get deveria ser nil antes de Start")
	}

	tm.Start(chatID, "scan", "descrição", func() {})

	if !tm.IsBusy(chatID) {
		t.Fatal("deveria estar ocupado depois de Start")
	}
	task := tm.Get(chatID)
	if task == nil || task.Type != "scan" || task.Description != "descrição" {
		t.Fatalf("task = %+v", task)
	}

	tm.End(chatID)

	if tm.IsBusy(chatID) {
		t.Fatal("não deveria estar ocupado depois de End")
	}
	if tm.Get(chatID) != nil {
		t.Fatal("Get deveria ser nil depois de End")
	}
}

func TestTaskManagerIsolaChats(t *testing.T) {
	tm := NewTaskManager()
	tm.Start(1, "scan", "a", func() {})

	if tm.IsBusy(2) {
		t.Fatal("chat diferente não deveria estar ocupado")
	}
}

func TestTaskManagerCancel(t *testing.T) {
	tm := NewTaskManager()
	chatID := int64(9)

	if tm.Cancel(chatID) {
		t.Fatal("Cancel deveria retornar false sem job rodando")
	}

	called := false
	tm.Start(chatID, "scan", "a", func() { called = true })

	if !tm.Cancel(chatID) {
		t.Fatal("Cancel deveria retornar true com job rodando")
	}
	if !called {
		t.Fatal("cancel func deveria ter sido chamado")
	}
}

func TestAwaitingInputTakeConsome(t *testing.T) {
	ai := NewAwaitingInput()
	chatID := int64(7)

	if _, ok := ai.Take(chatID); ok {
		t.Fatal("não deveria haver pendência antes de Set")
	}

	ai.Set(chatID, "report")

	kind, ok := ai.Take(chatID)
	if !ok || kind != "report" {
		t.Fatalf("Take = %q, %v; want report, true", kind, ok)
	}

	if _, ok := ai.Take(chatID); ok {
		t.Fatal("Take deveria consumir a pendência (segunda chamada deveria falhar)")
	}
}
