package tests_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reginaldalfret/git-koi/pkg/behavior"
	"github.com/reginaldalfret/git-koi/pkg/coat"
	"github.com/reginaldalfret/git-koi/pkg/history"
	"github.com/reginaldalfret/git-koi/pkg/koi"
	"github.com/reginaldalfret/git-koi/pkg/motion"
	"github.com/reginaldalfret/git-koi/pkg/output"
	"github.com/reginaldalfret/git-koi/pkg/render"
)

func TestFullRenderPipeline(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "account-with-history.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var snapshot history.AccountSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("failed to unmarshal snapshot: %v", err)
	}

	dev, err := koi.DeriveKoiDevelopment(&snapshot)
	if err != nil {
		t.Fatalf("failed to derive koi development: %v", err)
	}

	b := behavior.SelectBehavior(dev)
	t.Logf("Selected behavior: %s", b)

	stillFrame := render.RenderKoi(dev, coat.BodyWhite)
	pngBytes, err := output.EncodePNG(stillFrame)
	if err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("empty PNG bytes")
	}
	t.Logf("Encoded PNG: %d bytes", len(pngBytes))

	frameCount := motion.FrameCount(b)
	t.Logf("Rendering %d swim frames...", frameCount)

	frames := make([]*output.KoiFrame, frameCount)
	for i := uint32(0); i < frameCount; i++ {
		frames[i] = render.RenderSwimFrame(dev, coat.BodyWhite, b, i)
	}

	gifBytes, err := output.EncodeGIF(frames)
	if err != nil {
		t.Fatalf("failed to encode GIF: %v", err)
	}
	if len(gifBytes) == 0 {
		t.Fatal("empty GIF bytes")
	}
	t.Logf("Encoded GIF: %d bytes", len(gifBytes))
}
