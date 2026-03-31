package smoke_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/catu-ai/easyharness/tests/support"
)

func TestSyncContractArtifactsCheckPassesForCurrentRepo(t *testing.T) {
	repoRoot := support.RepoRoot(t)
	cmd := exec.Command(filepath.Join(repoRoot, "scripts", "sync-contract-artifacts"), "--check")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync-contract-artifacts --check: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Contract schemas and reference docs are in sync.") {
		t.Fatalf("unexpected check output:\n%s", output)
	}
}
