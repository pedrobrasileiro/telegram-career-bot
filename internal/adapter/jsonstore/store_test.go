package jsonstore

import (
	"path/filepath"
	"testing"
)

func TestWriteJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "out.json")

	type payload struct {
		Foo string `json:"foo"`
	}
	if err := writeJSON(path, payload{Foo: "bar"}); err != nil {
		t.Fatalf("writeJSON erro: %v", err)
	}

	got, err := loadJSON[payload](path)
	if err != nil {
		t.Fatalf("loadJSON erro: %v", err)
	}
	if got == nil || got.Foo != "bar" {
		t.Errorf("got %+v, want Foo=bar", got)
	}
}

func TestLoadJSONMissing(t *testing.T) {
	type payload struct{ Foo string }
	got, err := loadJSON[payload]("/caminho/que/nao/existe.json")
	if err == nil {
		t.Fatal("esperava erro pra arquivo inexistente")
	}
	if got != nil {
		t.Errorf("esperava nil pra arquivo inexistente, got %+v", got)
	}
}
