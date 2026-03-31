package contractsync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpectedStatusSchemaAllowsNullableCurrentOutputs(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	files, err := expectedFiles(repoRoot)
	if err != nil {
		t.Fatalf("expectedFiles: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(files["schema/commands/status.result.schema.json"], &schema); err != nil {
		t.Fatalf("unmarshal status schema: %v", err)
	}

	defs := schema["$defs"].(map[string]any)
	nextAction := defs["NextAction"].(map[string]any)
	nextActionProps := nextAction["properties"].(map[string]any)
	commandSchema := nextActionProps["command"].(map[string]any)
	if !schemaAllowsNull(commandSchema) {
		t.Fatalf("expected NextAction.command to allow null, got %#v", commandSchema)
	}

	statusResult := defs["StatusResult"].(map[string]any)
	statusProps := statusResult["properties"].(map[string]any)
	nextActionsSchema := statusProps["next_actions"].(map[string]any)
	if !schemaAllowsNull(nextActionsSchema) {
		t.Fatalf("expected StatusResult.next_actions to allow null, got %#v", nextActionsSchema)
	}
}

func TestCheckFilesFailsOnMissingAndUnexpectedGeneratedFiles(t *testing.T) {
	workdir := t.TempDir()
	ownedRoots := []string{
		filepath.Join(workdir, "schema"),
		filepath.Join(workdir, "docs", "reference", "contracts"),
	}

	if err := os.MkdirAll(filepath.Join(workdir, "schema"), 0o755); err != nil {
		t.Fatalf("mkdir schema: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "schema", "unexpected.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write unexpected file: %v", err)
	}

	err := checkFiles(workdir, ownedRoots, map[string][]byte{
		"schema/index.json": []byte("{\"ok\":true}\n"),
	})
	if err == nil {
		t.Fatal("expected checkFiles to fail")
	}
	message := err.Error()
	if !strings.Contains(message, "missing generated file: schema/index.json") {
		t.Fatalf("expected missing-file error, got %q", message)
	}
	if !strings.Contains(message, "unexpected generated file: schema/unexpected.json") {
		t.Fatalf("expected unexpected-file error, got %q", message)
	}
}

func TestWriteFilesReplacesOwnedRoots(t *testing.T) {
	workdir := t.TempDir()
	ownedRoots := []string{
		filepath.Join(workdir, "schema"),
		filepath.Join(workdir, "docs", "reference", "contracts"),
	}
	if err := os.MkdirAll(filepath.Join(workdir, "schema"), 0o755); err != nil {
		t.Fatalf("mkdir schema: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "schema", "stale.json"), []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	expected := map[string][]byte{
		"schema/index.json":                  []byte("{\"title\":\"ok\"}\n"),
		"docs/reference/contracts/README.md": []byte("# ok\n"),
	}
	if err := writeFiles(workdir, ownedRoots, expected); err != nil {
		t.Fatalf("writeFiles: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workdir, "schema", "stale.json")); !os.IsNotExist(err) {
		t.Fatalf("expected stale generated file to be removed, got err=%v", err)
	}
	if data, err := os.ReadFile(filepath.Join(workdir, "schema", "index.json")); err != nil || string(data) != "{\"title\":\"ok\"}\n" {
		t.Fatalf("unexpected schema/index.json contents: err=%v data=%q", err, data)
	}
}

func schemaAllowsNull(schema map[string]any) bool {
	oneOf, ok := schema["oneOf"].([]any)
	if !ok {
		return false
	}
	for _, branch := range oneOf {
		mapped, ok := branch.(map[string]any)
		if !ok {
			continue
		}
		if mapped["type"] == "null" {
			return true
		}
	}
	return false
}
