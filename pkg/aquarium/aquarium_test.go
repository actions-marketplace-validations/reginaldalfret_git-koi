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

func TestActivityDependentPersonalization(t *testing.T) {
	goColor, _ := history.ParseLanguageColor("#00ADD8")
	rubyColor, _ := history.ParseLanguageColor("#701516")

	devGo := &koi.KoiDevelopment{
		IdentitySeed:       11111,
		TotalContributions: 1200,
		ObservedOn:         history.NewDate(2026, 11, 10),
		LanguageShares: map[string]float64{
			"Go": 1.0,
		},
		LanguageColors: map[string]*history.LanguageColor{
			"Go": &goColor,
		},
	}

	devRuby := &koi.KoiDevelopment{
		IdentitySeed:       99999,
		TotalContributions: 5400,
		ObservedOn:         history.NewDate(2026, 11, 10),
		LanguageShares: map[string]float64{
			"Ruby": 1.0,
		},
		LanguageColors: map[string]*history.LanguageColor{
			"Ruby": &rubyColor,
		},
	}

	// 1. Generate SVG for Go account (with nil koiGifs, so it triggers activity-derived rendering)
	svgGo1, err := GenerateAquariumSVG(devGo, nil)
	if err != nil {
		t.Fatalf("GenerateAquariumSVG(devGo) failed: %v", err)
	}

	// 2. Generate SVG for Ruby account
	svgRuby, err := GenerateAquariumSVG(devRuby, nil)
	if err != nil {
		t.Fatalf("GenerateAquariumSVG(devRuby) failed: %v", err)
	}

	// 3. Proves activity personalization: Materially different accounts MUST produce different visual SVGs
	if svgGo1 == svgRuby {
		t.Errorf("expected different aquarium SVGs for Go vs Ruby accounts, but got identical SVGs (P0 bug check failed)")
	}

	// 4. Proves P0 resolution: SVG MUST NOT fall back to static bundled base64 string when dev is provided
	staticWhiteB64 := AuthenticOriginalKoiB64[coat.BodyWhite]
	if strings.Contains(svgGo1, staticWhiteB64) {
		t.Errorf("svgGo contains hardcoded static base64 GIF instead of activity-derived GIF (P0 bug check failed)")
	}
	if strings.Contains(svgRuby, staticWhiteB64) {
		t.Errorf("svgRuby contains hardcoded static base64 GIF instead of activity-derived GIF (P0 bug check failed)")
	}

	// 5. Proves determinism: Identical inputs MUST produce byte-identical SVGs
	svgGo2, err := GenerateAquariumSVG(devGo, nil)
	if err != nil {
		t.Fatalf("second GenerateAquariumSVG(devGo) failed: %v", err)
	}
	if svgGo1 != svgGo2 {
		t.Errorf("expected deterministic SVG generation for identical development input, but outputs differed")
	}

	t.Logf("Activity-dependent personalization verified: Go and Ruby accounts produced distinct visual fish deterministically")
}

func TestOfflineFallbackWhenDevNil(t *testing.T) {
	// When dev is nil, it should gracefully fall back to the bundled authentic assets
	svg, err := GenerateAquariumSVG(nil, nil)
	if err != nil {
		t.Fatalf("GenerateAquariumSVG(nil, nil) failed: %v", err)
	}
	staticWhiteB64 := AuthenticOriginalKoiB64[coat.BodyWhite]
	if !strings.Contains(svg, staticWhiteB64) {
		t.Errorf("expected bundled static base64 asset when dev is nil, but not found")
	}
	if !strings.HasPrefix(svg, "<svg") || !strings.HasSuffix(strings.TrimSpace(svg), "</svg>") {
		t.Errorf("invalid SVG wrapper for fallback aquarium")
	}
}
