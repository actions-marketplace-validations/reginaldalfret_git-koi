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

	// Test with mock gif data
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
	for _, spec := range DefaultFishSpecs {
		if !strings.Contains(svg, spec.PathID) {
			t.Errorf("SVG missing path ID %s", spec.PathID)
		}
		if !strings.Contains(svg, fmt.Sprintf("id=\"fish-%s\"", spec.Name)) {
			t.Errorf("SVG missing fish element fish-%s", spec.Name)
		}
	}
}
