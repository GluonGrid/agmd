package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildGoalPrompt(t *testing.T) {
	tasks := []*Task{
		{Name: "web-api", ProjectName: "agent-md", Subject: "Web API server", Content: "Add HTTP JSON API.\nReuse helpers."},
		{Name: "web-spa", ProjectName: "agent-md", Subject: "", Content: ""},
	}
	prompt := buildGoalPrompt(tasks)

	if !strings.HasPrefix(prompt, "/goal ") {
		t.Errorf("prompt should start with '/goal ', got: %q", prompt[:20])
	}
	for _, want := range []string{
		"agent-md/web-api — Web API server",
		"Add HTTP JSON API.",
		"  Reuse helpers.",
		"agent-md/web-spa — web-spa", // subject falls back to name
		"agmd task status",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q\n---\n%s", want, prompt)
		}
	}
}

func TestSessionTranscriptPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := sessionTranscriptPath("/Users/sky/git/agent-md", "6d437ba2-dead-beef")
	if err != nil {
		t.Fatalf("sessionTranscriptPath: %v", err)
	}
	want := filepath.Join(home, ".claude", "projects", "-Users-sky-git-agent-md", "6d437ba2-dead-beef.jsonl")
	if got != want {
		t.Errorf("transcript path = %q, want %q", got, want)
	}
}

func TestNewSessionIDFormat(t *testing.T) {
	id, err := newSessionID()
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("session id %q is not a 5-part UUID", id)
	}
	lengths := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != lengths[i] {
			t.Errorf("uuid part %d = %q (len %d), want len %d", i, p, len(p), lengths[i])
		}
	}
	if parts[2][0] != '4' {
		t.Errorf("expected version-4 uuid, got %q", id)
	}
}

func TestTranscriptLineToEvent(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantKind string
		wantText string
		wantOK   bool
	}{
		{
			name:     "assistant text + tool",
			line:     `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Working on it"},{"type":"tool_use","name":"Bash"}]}}`,
			wantKind: "assistant",
			wantText: "Working on it\n🔧 Bash",
			wantOK:   true,
		},
		{
			name:     "user string content",
			line:     `{"type":"user","message":{"role":"user","content":"hello there"}}`,
			wantKind: "user",
			wantText: "hello there",
			wantOK:   true,
		},
		{
			name:     "summary",
			line:     `{"type":"summary","summary":"Implemented the feature"}`,
			wantKind: "summary",
			wantText: "Implemented the feature",
			wantOK:   true,
		},
		{
			name:   "unparseable",
			line:   `not json`,
			wantOK: false,
		},
		{
			name:   "empty assistant text is skipped",
			line:   `{"type":"assistant","message":{"role":"assistant","content":[]}}`,
			wantOK: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev, ok := transcriptLineToEvent(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if ev.Kind != tc.wantKind {
				t.Errorf("kind = %q, want %q", ev.Kind, tc.wantKind)
			}
			if ev.Text != tc.wantText {
				t.Errorf("text = %q, want %q", ev.Text, tc.wantText)
			}
		})
	}
}
