package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "git-koi-test.exe")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/git-koi")
	cmd.Dir = ".."
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build git-koi binary: %v\n%s", err, stderr.String())
	}
	return binPath
}

func TestCLIVersionAndHelp(t *testing.T) {
	bin := buildBinary(t)

	// 1. --help exits 0
	cmdHelp := exec.Command(bin, "--help")
	outHelp, err := cmdHelp.CombinedOutput()
	if err != nil {
		t.Errorf("--help failed: %v", err)
	}
	if !strings.Contains(string(outHelp), "Usage: git-koi") {
		t.Errorf("expected usage string in help output")
	}

	// 2. --version exits 0
	cmdVer := exec.Command(bin, "--version")
	outVer, err := cmdVer.CombinedOutput()
	if err != nil {
		t.Errorf("--version failed: %v", err)
	}
	if !strings.Contains(string(outVer), "git-koi 1.1.0") {
		t.Errorf("expected 'git-koi 1.1.0' in version output, got: %s", string(outVer))
	}

	// 3. No args exits 2
	cmdNoArgs := exec.Command(bin)
	errNoArgs := cmdNoArgs.Run()
	if errNoArgs == nil {
		t.Errorf("expected error for no args invocation")
	}
}

func TestCLIRenderAndAquarium(t *testing.T) {
	bin := buildBinary(t)
	fixturePath, _ := filepath.Abs("fixtures/account-with-history.json")
	outDir := t.TempDir()

	// 1. Test render command
	renderDir := filepath.Join(outDir, "render_test")
	cmdRender := exec.Command(bin, "render", fixturePath, "--output-dir", renderDir, "--body", "asagi")
	outRender, err := cmdRender.CombinedOutput()
	if err != nil {
		t.Fatalf("render command failed: %v\nOutput: %s", err, string(outRender))
	}

	for _, filename := range []string{"koi.gif", "koi.png", "preview.html"} {
		p := filepath.Join(renderDir, filename)
		if fi, err := os.Stat(p); err != nil || fi.Size() == 0 {
			t.Errorf("expected generated non-empty file %s", p)
		}
	}

	// 2. Test aquarium command
	aquariumDir := filepath.Join(outDir, "aquarium_test")
	cmdAqua := exec.Command(bin, "aquarium", fixturePath, "--output-dir", aquariumDir)
	outAqua, err := cmdAqua.CombinedOutput()
	if err != nil {
		t.Fatalf("aquarium command failed: %v\nOutput: %s", err, string(outAqua))
	}

	aquaPath := filepath.Join(aquariumDir, "aquarium.svg")
	fi, err := os.Stat(aquaPath)
	if err != nil || fi.Size() == 0 {
		t.Fatalf("expected generated aquarium.svg in %s", aquariumDir)
	}

	// 3. Test unknown body flag exits 2
	cmdUnknown := exec.Command(bin, "render", fixturePath, "--output-dir", renderDir, "--body", "rainbow")
	if err := cmdUnknown.Run(); err == nil {
		t.Errorf("expected error for unknown body 'rainbow'")
	}

	// 4. Test missing snapshot file exits 1
	cmdMissing := exec.Command(bin, "aquarium", "nonexistent-file.json", "--output-dir", aquariumDir)
	if err := cmdMissing.Run(); err == nil {
		t.Errorf("expected error for missing snapshot file")
	}
}
