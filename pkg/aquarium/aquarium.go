package aquarium

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/reginaldalfret/git-koi/pkg/behavior"
	"github.com/reginaldalfret/git-koi/pkg/coat"
	"github.com/reginaldalfret/git-koi/pkg/koi"
	"github.com/reginaldalfret/git-koi/pkg/motion"
	"github.com/reginaldalfret/git-koi/pkg/output"
	"github.com/reginaldalfret/git-koi/pkg/render"
)

// Pond constants
const (
	PondWidth  = 1000
	PondHeight = 420
)

// FishSpec defines trajectory and visual properties for a koi in the aquarium.
type FishSpec struct {
	Body       coat.KoiBody
	Name       string
	Variety    string
	Scale      float64
	Duration   string
	PathID     string
	PathD      string
	ShadowDx   int
	ShadowDy   int
}

// DefaultFishSpecs defines the 6 distinct natural swimming paths with non-repeating prime periods.
var DefaultFishSpecs = []FishSpec{
	{
		Body:     coat.BodyAsagi,
		Name:     "asagi",
		Variety:  "Asagi",
		Scale:    0.42,
		Duration: "31s",
		PathID:   "swimPathAsagi",
		PathD:    "M 150 180 C 220 120, 380 90, 520 130 C 660 170, 780 140, 860 210 C 920 270, 880 340, 780 350 C 640 360, 560 280, 440 270 C 300 260, 180 360, 110 300 C 60 250, 90 220, 150 180 Z",
		ShadowDx: 14,
		ShadowDy: 20,
	},
	{
		Body:     coat.BodyBlack,
		Name:     "black",
		Variety:  "Karasu",
		Scale:    0.40,
		Duration: "39s",
		PathID:   "swimPathBlack",
		PathD:    "M 820 140 C 740 100, 600 130, 480 200 C 340 280, 240 260, 160 330 C 100 380, 160 400, 260 380 C 400 350, 460 310, 600 320 C 760 330, 880 290, 910 220 C 930 160, 880 160, 820 140 Z",
		ShadowDx: 16,
		ShadowDy: 24,
	},
	{
		Body:     coat.BodyGold,
		Name:     "gold",
		Variety:  "Yamabuki",
		Scale:    0.44,
		Duration: "23s",
		PathID:   "swimPathGold",
		PathD:    "M 300 280 C 200 320, 120 220, 200 140 C 280 80, 420 110, 550 170 C 680 230, 750 110, 830 120 C 910 140, 920 230, 840 280 C 720 350, 620 240, 500 220 C 400 200, 360 260, 300 280 Z",
		ShadowDx: 12,
		ShadowDy: 18,
	},
	{
		Body:     coat.BodyPlatinum,
		Name:     "platinum",
		Variety:  "Ogon",
		Scale:    0.43,
		Duration: "43s",
		PathID:   "swimPathPlatinum",
		PathD:    "M 500 340 C 350 350, 220 320, 180 240 C 140 160, 230 100, 360 80 C 520 60, 680 70, 800 110 C 900 150, 940 240, 880 300 C 800 370, 660 330, 500 340 Z",
		ShadowDx: 15,
		ShadowDy: 22,
	},
	{
		Body:     coat.BodyTea,
		Name:     "tea",
		Variety:  "Chagoi",
		Scale:    0.41,
		Duration: "29s",
		PathID:   "swimPathTea",
		PathD:    "M 650 220 C 580 160, 490 150, 410 190 C 320 240, 310 320, 400 350 C 500 380, 600 330, 710 310 C 800 290, 850 210, 790 160 C 730 120, 700 260, 650 220 Z",
		ShadowDx: 14,
		ShadowDy: 20,
	},
	{
		Body:     coat.BodyWhite,
		Name:     "white",
		Variety:  "Kohaku",
		Scale:    0.42,
		Duration: "37s",
		PathID:   "swimPathWhite",
		PathD:    "M 450 90 C 580 80, 720 140, 790 200 C 860 270, 780 340, 690 320 C 580 300, 520 230, 420 250 C 300 280, 240 200, 280 140 C 320 90, 380 90, 450 90 Z",
		ShadowDx: 13,
		ShadowDy: 19,
	},
}

// RenderAquariumGifs renders 1x scaled animated GIFs for all 6 koi bodies in memory.
func RenderAquariumGifs(dev *koi.KoiDevelopment) (map[coat.KoiBody][]byte, error) {
	b := behavior.SelectBehavior(dev)
	frameCount := motion.FrameCount(b)
	result := make(map[coat.KoiBody][]byte)

	for _, spec := range DefaultFishSpecs {
		frames := make([]*output.KoiFrame, frameCount)
		for i := uint32(0); i < frameCount; i++ {
			frames[i] = render.RenderSwimFrame(dev, spec.Body, b, i)
		}
		// Scale = 1 produces compact ~35KB GIF per fish
		gifBytes, err := output.EncodeScaledGIF(frames, 1)
		if err != nil {
			return nil, fmt.Errorf("encoding %s koi: %w", spec.Name, err)
		}
		result[spec.Body] = gifBytes
	}

	return result, nil
}

// GenerateAquariumSVG builds the self-contained animated SVG aquarium containing all 6 moving fishes.
func GenerateAquariumSVG(dev *koi.KoiDevelopment, koiGifs map[coat.KoiBody][]byte) (string, error) {
	if koiGifs == nil {
		var err error
		koiGifs, err = RenderAquariumGifs(dev)
		if err != nil {
			return "", err
		}
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %d %d" width="100%%" height="auto" style="border-radius:16px; overflow:hidden; background:#070d19;">
<defs>
  <!-- Water Gradients -->
  <radialGradient id="pondDeep" cx="50%%" cy="45%%" r="60%%">
    <stop offset="0%%" stop-color="#0c233c" stop-opacity="0.9"/>
    <stop offset="55%%" stop-color="#071526" stop-opacity="0.95"/>
    <stop offset="100%%" stop-color="#030710" stop-opacity="1"/>
  </radialGradient>

  <linearGradient id="surfaceGlow" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#14b8a6" stop-opacity="0.12"/>
    <stop offset="50%%" stop-color="#38bdf8" stop-opacity="0.05"/>
    <stop offset="100%%" stop-color="#0284c7" stop-opacity="0.15"/>
  </linearGradient>

  <!-- Soft Shadow Filter for Fish Depth -->
  <filter id="pondShadow" x="-30%%" y="-30%%" width="160%%" height="160%%">
    <feGaussianBlur in="SourceAlpha" stdDeviation="4"/>
    <feColorMatrix type="matrix" values="0 0 0 0 0   0 0 0 0 0   0 0 0 0 0  0 0 0 0.35 0"/>
    <feOffset dx="6" dy="10"/>
    <feBlend in="SourceGraphic" in2="blurOut" mode="normal"/>
  </filter>

  <!-- Caustic Light Shaft Filter -->
  <filter id="causticBlur">
    <feGaussianBlur stdDeviation="8"/>
  </filter>

  <!-- Reusable Lily Pad -->
  <g id="lilyPadLarge">
    <path d="M 0 0 L 35 -15 A 38 38 0 1 1 -35 -15 Z" fill="#1b4d2e" stroke="#2d7a46" stroke-width="1.5"/>
    <path d="M 0 0 L 28 -30 M 0 0 L -28 -30 M 0 0 L -38 0 M 0 0 L 38 0 M 0 0 L -20 28 M 0 0 L 20 28" stroke="#163e25" stroke-width="0.8" opacity="0.6"/>
  </g>
  <g id="lilyPadMedium">
    <path d="M 0 0 L 25 -10 A 28 28 0 1 1 -25 -10 Z" fill="#235c37" stroke="#3b9657" stroke-width="1.2"/>
    <path d="M 0 0 L 20 -20 M 0 0 L -20 -20 M 0 0 L -28 0 M 0 0 L 28 0" stroke="#1a462a" stroke-width="0.7" opacity="0.6"/>
  </g>

  <!-- Reusable Floating Lotus Flower -->
  <g id="lotusFlower">
    <ellipse cx="-4" cy="-3" rx="10" ry="5" fill="#fbcfe8" transform="rotate(-30)" opacity="0.85"/>
    <ellipse cx="4" cy="-3" rx="10" ry="5" fill="#fbcfe8" transform="rotate(30)" opacity="0.85"/>
    <ellipse cx="0" cy="-6" rx="10" ry="5" fill="#fdf2f8" opacity="0.95"/>
    <ellipse cx="-5" cy="2" rx="9" ry="4" fill="#f472b6" transform="rotate(-60)" opacity="0.75"/>
    <ellipse cx="5" cy="2" rx="9" ry="4" fill="#f472b6" transform="rotate(60)" opacity="0.75"/>
    <circle cx="0" cy="-1" r="3.5" fill="#fef08a"/>
  </g>

  <!-- Motion Paths -->
`, PondWidth, PondHeight))

	// Write motion paths into <defs>
	for _, spec := range DefaultFishSpecs {
		sb.WriteString(fmt.Sprintf(`  <path id="%s" d="%s" fill="none"/>
`, spec.PathID, spec.PathD))
	}

	sb.WriteString(`</defs>

<!-- Pond Bed & Water Column -->
<rect width="100%" height="100%" fill="url(#pondDeep)"/>
<rect width="100%" height="100%" fill="url(#surfaceGlow)"/>

<!-- Ambient Sunlight Caustics -->
<g opacity="0.18" filter="url(#causticBlur)">
  <polygon points="120,0 280,0 420,420 180,420" fill="#38bdf8">
    <animate attributeName="opacity" values="0.12;0.22;0.12" dur="8s" repeatCount="indefinite"/>
  </polygon>
  <polygon points="460,0 640,0 800,420 540,420" fill="#2dd4bf">
    <animate attributeName="opacity" values="0.15;0.26;0.15" dur="11s" repeatCount="indefinite"/>
  </polygon>
  <polygon points="760,0 920,0 1000,420 860,420" fill="#38bdf8">
    <animate attributeName="opacity" values="0.10;0.20;0.10" dur="9s" repeatCount="indefinite"/>
  </polygon>
</g>

<!-- Animated Surface Ripples -->
<g fill="none" stroke="#38bdf8" stroke-width="1.2" opacity="0.4">
  <!-- Feeding Ring 1 (Center-Left) -->
  <circle cx="340" cy="170" r="2">
    <animate attributeName="r" values="2;55" dur="7s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.6;0" dur="7s" repeatCount="indefinite"/>
  </circle>
  <circle cx="340" cy="170" r="2">
    <animate attributeName="r" values="2;55" dur="7s" begin="2.5s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.6;0" dur="7s" begin="2.5s" repeatCount="indefinite"/>
  </circle>

  <!-- Feeding Ring 2 (Center-Right) -->
  <circle cx="720" cy="250" r="2">
    <animate attributeName="r" values="2;65" dur="9s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.55;0" dur="9s" repeatCount="indefinite"/>
  </circle>
  <circle cx="720" cy="250" r="2">
    <animate attributeName="r" values="2;65" dur="9s" begin="3.2s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.55;0" dur="9s" begin="3.2s" repeatCount="indefinite"/>
  </circle>

  <!-- Feeding Ring 3 (Upper Right) -->
  <circle cx="580" cy="90" r="2">
    <animate attributeName="r" values="2;45" dur="6s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.5;0" dur="6s" repeatCount="indefinite"/>
  </circle>
</g>

<!-- SWIMMING KOI FISH (6 NATIVE BODIES) -->
`)

	// Render each fish
	for _, spec := range DefaultFishSpecs {
		gifData, ok := koiGifs[spec.Body]
		if !ok || len(gifData) == 0 {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(gifData)

		sb.WriteString(fmt.Sprintf(`
<!-- Koi: %s (%s) -->
<g id="fish-%s">
  <!-- Fish Body & Dynamic Sprite -->
  <g>
    <animateMotion dur="%s" repeatCount="indefinite" rotate="auto">
      <mpath xlink:href="#%s"/>
    </animateMotion>
    
    <!-- Pond floor depth shadow -->
    <ellipse cx="-%d" cy="%d" rx="42" ry="16" fill="#000000" opacity="0.32" filter="url(#pondShadow)"/>
    
    <!-- Swimming Animated Koi Sprite (centered at origin, facing tangent) -->
    <g transform="scale(%.3f)">
      <image xlink:href="data:image/gif;base64,%s" x="-128" y="-128" width="256" height="256" style="image-rendering:pixelated; image-rendering:crisp-edges;"/>
    </g>
  </g>
</g>
`, spec.Name, spec.Variety, spec.Name, spec.Duration, spec.PathID, spec.ShadowDx, spec.ShadowDy, spec.Scale, b64))
	}

	sb.WriteString(`
<!-- FLOATING BOTANICAL LILY PADS (Top Aquatic Layer) -->
<g id="botanicals" opacity="0.92">
  <!-- Top Left Cluster -->
  <g transform="translate(110, 80) scale(1.1)">
    <use href="#lilyPadLarge"/>
  </g>
  <g transform="translate(65, 120) scale(0.9) rotate(45)">
    <use href="#lilyPadMedium"/>
  </g>
  <g transform="translate(115, 75) scale(0.8)">
    <use href="#lotusFlower"/>
  </g>

  <!-- Bottom Right Cluster -->
  <g transform="translate(890, 330) scale(1.25) rotate(-30)">
    <use href="#lilyPadLarge"/>
  </g>
  <g transform="translate(945, 290) scale(0.85) rotate(110)">
    <use href="#lilyPadMedium"/>
  </g>
  <g transform="translate(895, 325) scale(0.9) rotate(-15)">
    <use href="#lotusFlower"/>
  </g>
</g>

<!-- Atmospheric HUD Overlay & Glass Edge -->
<rect width="100%" height="100%" rx="16" ry="16" fill="none" stroke="#38bdf8" stroke-width="1.5" stroke-opacity="0.35"/>
<rect x="1" y="1" width="998" height="418" rx="15" ry="15" fill="none" stroke="#0f172a" stroke-width="1" stroke-opacity="0.8"/>

<!-- Subtle Aquarium Title Badge -->
<g transform="translate(24, 30)" opacity="0.85">
  <circle cx="0" cy="0" r="4" fill="#34d399">
    <animate attributeName="opacity" values="0.4;1;0.4" dur="2s" repeatCount="indefinite"/>
  </circle>
  <text x="12" y="4" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif" font-size="11" font-weight="600" fill="#94a3b8" letter-spacing="1.2">LIVE KOI AQUARIUM · 6 NATIVE ECOSYSTEM BODIES</text>
</g>

<g transform="translate(976, 30)" text-anchor="end" opacity="0.65">
  <text x="0" y="4" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif" font-size="10" font-weight="500" fill="#64748b" letter-spacing="0.8">POWERED BY GIT-KOI (GO)</text>
</g>
</svg>
`)

	return sb.String(), nil
}
