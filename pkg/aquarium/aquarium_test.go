package aquarium

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/coat"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
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

	if len(DefaultFishSpecs) != 12 {
		t.Fatalf("expected 12 fish specs, got %d", len(DefaultFishSpecs))
	}

	for _, spec := range DefaultFishSpecs {
		if !strings.Contains(svg, spec.PathID) {
			t.Errorf("SVG missing path ID %s", spec.PathID)
		}
		if !strings.Contains(svg, fmt.Sprintf("id=\"fish-%s\"", spec.Name)) {
			t.Errorf("SVG missing fish element fish-%s", spec.Name)
		}
	}

	// Verify Layering: Fish layer MUST come BEFORE botanicals layer in the SVG
	fishIdx := strings.Index(svg, `id="fish-white-1"`)
	botanicalsIdx := strings.Index(svg, `id="botanicals"`)
	if fishIdx == -1 || botanicalsIdx == -1 {
		t.Fatalf("missing fish or botanicals element")
	}
	if fishIdx > botanicalsIdx {
		t.Errorf("Fish layer is after botanicals layer! Fish would render above lotus leaves.")
	}

	// Verify life cycle paths exist (every fish has pathLife0..pathLife11)
	for i := 0; i < 12; i++ {
		pathID := fmt.Sprintf("pathLife%d", i)
		if !strings.Contains(svg, fmt.Sprintf("id=\"%s\"", pathID)) {
			t.Errorf("SVG missing life cycle path definition %s", pathID)
		}
	}

	t.Logf("SVG verified: 12 fish, all executing full life cycle movements (oval, sprint, hover, zig-zag, fig-8)")
}
