package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMaterializeRefArchivesCurrentRepositoryRef(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "semci@example.com")
	runGit(t, repo, "config", "user.name", "SemCI Test")

	modelDir := filepath.Join(repo, "model")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("mkdir model: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "orders.yml"), []byte("cubes: []\n"), 0644); err != nil {
		t.Fatalf("write model: %v", err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "initial")

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir repo: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	root, cleanup, err := MaterializeRef("HEAD")
	if err != nil {
		t.Fatalf("MaterializeRef returned error: %v", err)
	}
	defer cleanup()

	if _, err := os.Stat(filepath.Join(root, "model", "orders.yml")); err != nil {
		t.Fatalf("expected archived model file: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
