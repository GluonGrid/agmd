package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type claimResp struct {
	Action        string `json:"action"`
	TakenOverFrom string `json:"taken_over_from"`
	Task          struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Assignment struct {
			WorkflowID string `json:"workflow_id"`
			UpdatedAt  string `json:"updated_at"`
		} `json:"assignment"`
		Comments []commentItem `json:"comments"`
	} `json:"task"`
}

func claim(t *testing.T, workflowID string, extra ...string) (claimResp, error) {
	t.Helper()
	args := append([]string{"task", "next", "--claim", "--workflow-id", workflowID, "--project", "p", "--json"}, extra...)
	out, err := runE(t, args...)
	var resp claimResp
	if err == nil {
		if jsonErr := json.Unmarshal([]byte(out), &resp); jsonErr != nil {
			t.Fatalf("decode claim: %v\n%s", jsonErr, out)
		}
	}
	return resp, err
}

// ageClaims rewinds every claim's updated_at by d.
func ageClaims(t *testing.T, registryPath string, d time.Duration) {
	t.Helper()
	path := filepath.Join(registryPath, "task", "p", "assignments.json")
	assignments := map[string]map[string]string{}
	if err := json.Unmarshal([]byte(readFile(t, path)), &assignments); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().Add(-d).Format(time.RFC3339)
	for id := range assignments {
		assignments[id]["updated_at"] = old
	}
	raw, _ := json.Marshal(assignments)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTaskClaim_FreshClaimIsNotTakenOver(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")

	if _, err := claim(t, "agent-1"); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if _, err := claim(t, "agent-2", "--stale-after", "1h"); err == nil || !strings.Contains(err.Error(), "no ready") {
		t.Errorf("second claim should find nothing, got %v", err)
	}
}

func TestTaskClaim_StaleClaimIsTakenOver(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")
	if _, err := claim(t, "agent-1"); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	ageClaims(t, registryPath, 3*time.Hour)

	if _, err := claim(t, "agent-2", "--stale-after", "0"); err == nil {
		t.Error("--stale-after 0 should disable expiry")
	}

	resp, err := claim(t, "agent-2", "--stale-after", "1h")
	if err != nil {
		t.Fatalf("takeover claim: %v", err)
	}
	if resp.Action != "takeover" || resp.TakenOverFrom != "agent-1" {
		t.Errorf("action=%q taken_over_from=%q", resp.Action, resp.TakenOverFrom)
	}
	if resp.Task.Assignment.WorkflowID != "agent-2" || resp.Task.Status != "in_progress" {
		t.Errorf("task after takeover = %+v", resp.Task)
	}
	if len(resp.Task.Comments) != 1 || !strings.Contains(resp.Task.Comments[0].Body, "agent-1") {
		t.Errorf("expected a takeover comment, got %+v", resp.Task.Comments)
	}

	// The default TTL (2h) also applies when no flag is given.
	ageClaims(t, registryPath, 3*time.Hour)
	if resp, err := claim(t, "agent-3"); err != nil || resp.TakenOverFrom != "agent-2" {
		t.Errorf("default TTL takeover: resp=%+v err=%v", resp, err)
	}

	t.Setenv("AGMD_CLAIM_TTL", "24h")
	ageClaims(t, registryPath, 3*time.Hour)
	if _, err := claim(t, "agent-4"); err == nil {
		t.Error("AGMD_CLAIM_TTL=24h should keep a 3h-old claim alive")
	}
}

func TestTaskClaim_ReadyTaskPreferredOverStaleOne(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")
	if _, err := claim(t, "agent-1"); err != nil {
		t.Fatal(err)
	}
	ageClaims(t, registryPath, 3*time.Hour)
	run(t, "task", "new", "b", "--project", "p", "--content", "work")

	resp, err := claim(t, "agent-2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Action != "claim" || resp.Task.Name != "b" {
		t.Errorf("expected fresh claim of b, got action=%q task=%q", resp.Action, resp.Task.Name)
	}
}

func TestTaskClaim_HeartbeatAndCommentKeepClaimAlive(t *testing.T) {
	registryPath := setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")
	if _, err := claim(t, "agent-1"); err != nil {
		t.Fatal(err)
	}

	ageClaims(t, registryPath, 3*time.Hour)
	run(t, "task", "heartbeat", "--workflow-id", "agent-1", "--project", "p")
	if _, err := claim(t, "agent-2"); err == nil {
		t.Error("heartbeat should keep the claim alive")
	}

	ageClaims(t, registryPath, 3*time.Hour)
	run(t, "task", "comment", "a", "still going", "--author", "agent-1", "--project", "p")
	if _, err := claim(t, "agent-2"); err == nil {
		t.Error("a comment from the owner should keep the claim alive")
	}

	ageClaims(t, registryPath, 3*time.Hour)
	run(t, "task", "comment", "a", "drive-by", "--author", "someone-else", "--project", "p")
	if _, err := claim(t, "agent-2"); err != nil {
		t.Errorf("a comment from another author must not refresh the claim: %v", err)
	}

	// agent-1 has lost the task: a targeted heartbeat says so.
	_, err := runE(t, "task", "heartbeat", "a", "--workflow-id", "agent-1", "--project", "p")
	if err == nil || !strings.Contains(err.Error(), "agent-2") {
		t.Errorf("expected conflict naming agent-2, got %v", err)
	}
	if _, err := runE(t, "task", "heartbeat", "--workflow-id", "agent-1", "--project", "p"); err == nil {
		t.Error("heartbeat for a workflow holding nothing should fail")
	}
}

func TestTaskClaim_InvalidStaleAfter(t *testing.T) {
	setup(t)
	chdir(t)
	run(t, "task", "new", "a", "--project", "p", "--content", "work")
	if _, err := claim(t, "agent-1", "--stale-after", "soon"); err == nil || !strings.Contains(err.Error(), "invalid --stale-after") {
		t.Errorf("expected invalid --stale-after error, got %v", err)
	}
}
