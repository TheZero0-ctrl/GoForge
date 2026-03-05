package command

import "testing"

func TestRegistryResolvesAlias(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(NewGenerateCommand()); err != nil {
		t.Fatalf("register generate command: %v", err)
	}

	cmd, ok := reg.Resolve("g")
	if !ok {
		t.Fatal("expected alias to resolve")
	}

	if got := cmd.Spec().ID; got != "generate" {
		t.Fatalf("expected generate, got %q", got)
	}
}
