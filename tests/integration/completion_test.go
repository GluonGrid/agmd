package integration

import (
	"strings"
	"testing"
)

func TestCompletion_Bash(t *testing.T) {
	setup(t)

	out := run(t, "completion", "bash")
	if !strings.Contains(out, "bash") {
		t.Error("completion bash should output bash completion script")
	}
}

func TestCompletion_Zsh(t *testing.T) {
	setup(t)

	out := run(t, "completion", "zsh")
	if !strings.Contains(out, "zsh") {
		t.Error("completion zsh should output zsh completion script")
	}
}
