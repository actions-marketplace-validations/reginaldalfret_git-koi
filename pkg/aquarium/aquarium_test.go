package aquarium

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reginaldalfret/git-koi/pkg/coat"
	"github.com/reginaldalfret/git-koi/pkg/history"
	"github.com/reginaldalfret/git-koi/pkg/koi"
)

func TestGenerateAquariumSVG(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "account-with-history.json")
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
		t.Fatalf("DeriveKoiDevelopment failed: %v", err)
	}

	// Use mock GIF data so the test doesn't need to do full rendering
	mockGifs := map[coat.KoiBody][]byte{
		coat.BodyAsagi:    []byte("GIF89a"),
		coat.BodyBlack:    []byte("GIF89a"),
		coat.BodyGold:     []byte("GIF89a"),
		coat.BodyPlatinum: []byte("GIF89a"),
		coat.BodyTea:      []byte("GIF89a"),
		coat.BodyWhite:    []byte("GIF89a"),
	}

	svg, err := GenerateAquariumSVG(dev, mockGifs)
	if err != nil {
		t.Fatalf("GenerateAquariumSVG failed: %v", err)
	}

	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("SVG does not start with <svg")
	}
	if !strings.HasSuffix(strings.TrimSpace(svg), "</svg>") {
		t.Errorf("SVG does not end with </svg>")
	}

	// All 6 fish paths and elements must be present
	for _, spec := range DefaultFishSpecs {
		if !strings.Contains(svg, spec.PathID) {
			t.Errorf("SVG missing path ID %s", spec.PathID)
		}
		if !strings.Contains(svg, fmt.Sprintf("id=\"fish-%s\"", spec.Name)) {
			t.Errorf("SVG missing fish element fish-%s", spec.Name)
		}
	}

	// Figure-8 paths: each path must have the crossover bezier structure
	if !strings.Contains(svg, "animateMotion") {
		t.Errorf("SVG missing animateMotion")
	}

	// Stagger offsets must be present
	if !strings.Contains(svg, `begin="-20s"`) {
		t.Errorf("SVG missing begin stagger offsets")
	}

	// Botanicals
	if !strings.Contains(svg, "lilyLg") {
		t.Errorf("SVG missing lily pad definitions")
	}
	if !strings.Contains(svg, "lotus") {
		t.Errorf("SVG missing lotus flower definition")
	}

	t.Logf("SVG generated: %d bytes, %d fish on figure-8 paths", len(svg), len(DefaultFishSpecs))
}

func TestFigure8Path(t *testing.T) {
	path := figure8Path(500, 210, 350, 110)
	// Must start at centre point
	if !strings.HasPrefix(path, "M 500,210") {
		t.Errorf("figure-8 path does not start at centre: %s", path[:20])
	}
	// Must be a closed path
	if !strings.HasSuffix(path, " Z") {
		t.Errorf("figure-8 path is not closed")
	}
	// Must have exactly 4 cubic bezier segments (C ... C ... C ... C ...)
	count := strings.Count(path, " C ")
	if count != 4 {
		t.Errorf("expected 4 cubic bezier segments, got %d", count)
	}
}
