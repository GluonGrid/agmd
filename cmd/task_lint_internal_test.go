package cmd

import "testing"

func TestSaysReplaced(t *testing.T) {
	taskMap := buildTaskMap([]*Task{{Name: "auth"}, {Name: "fix-auth"}})
	for _, tc := range []struct {
		body string
		want bool
	}{
		{"Replaced by fix-auth", true},
		{"replaced by auth", true},
		{"This task was superseded by `fix-auth`.", true},
		{"Superseded by agents/move-board", true},
		{"replaced by new-name (not created yet)", true},
		{"Closed: superseded by fix-auth; see there", true},
		// plans, negations and questions
		{"Should be replaced by fix-auth once it lands", false},
		{"this is not replaced by fix-auth", false},
		{"Could this be superseded by fix-auth?", false},
		{"will be replaced by fix-auth", false},
		// the word alone, or "by" something that is not a task
		{"The config file is replaced on every deploy", false},
		{"superseded", false},
		{"Replaced the old token with a new one", false},
		{"replaced by the team", false},
		{"replaced by hand", false},
	} {
		if got := saysReplaced(tc.body, taskMap); got != tc.want {
			t.Errorf("saysReplaced(%q) = %v, want %v", tc.body, got, tc.want)
		}
	}
}
