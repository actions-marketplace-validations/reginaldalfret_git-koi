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
	Body    coat.KoiBody
	Name    string
	Variety string
	Scale   float64
	// Animation
	Duration string
	Begin    string // negative offset to stagger fish start positions on their figure-8
	PathID   string
	// figure-8 path: M cx,cy C ... (two mirrored loops)
	PathD string
}

// figure8Path generates a smooth figure-8 (lemniscate) SVG path.
// cx,cy = centre crossover point; rx = half-width of each loop; ry = half-height of each loop.
// The fish starts at the centre, loops right, returns, loops left, returns.
func figure8Path(cx, cy, rx, ry int) string {
	// Control-point offset — 0.55 * radius gives a near-circular bezier loop
	cpx := int(float64(rx) * 0.72)
	cpy := int(float64(ry) * 0.72)
	rx1 := cx + rx
	rx0 := cx - rx

	return fmt.Sprintf(
		"M %d,%d C %d,%d %d,%d %d,%d C %d,%d %d,%d %d,%d C %d,%d %d,%d %d,%d C %d,%d %d,%d %d,%d Z",
		// Start at centre
		cx, cy,
		// Right loop — upper arc
		cx+cpx, cy-cpy, rx1, cy-cpy, rx1, cy,
		// Right loop — lower arc  
		rx1, cy+cpy, cx+cpx, cy+cpy, cx, cy,
		// Left loop — lower arc
		cx-cpx, cy+cpy, rx0, cy+cpy, rx0, cy,
		// Left loop — upper arc
		rx0, cy-cpy, cx-cpx, cy-cpy, cx, cy,
	)
}

// DefaultFishSpecs defines 6 koi with individual figure-8 swim paths.
// Paths vary in size and centre so fish spread naturally across the pond.
// Begin offsets place each fish at a different starting point on its path.
var DefaultFishSpecs = []FishSpec{
	{
		Body:     coat.BodyAsagi,
		Name:     "asagi",
		Variety:  "Asagi",
		Scale:    0.44,
		Duration: "31s",
		Begin:    "0s",
		PathID:   "swimPathAsagi",
		// Large figure-8 spanning full pond, centred at (500,200)
		PathD: figure8Path(500, 200, 360, 115),
	},
	{
		Body:     coat.BodyBlack,
		Name:     "black",
		Variety:  "Karasu",
		Scale:    0.42,
		Duration: "43s",
		Begin:    "-20s",
		PathID:   "swimPathBlack",
		// Wide but lower figure-8, centred at (500,240)
		PathD: figure8Path(500, 240, 380, 125),
	},
	{
		Body:     coat.BodyGold,
		Name:     "gold",
		Variety:  "Yamabuki",
		Scale:    0.46,
		Duration: "23s",
		Begin:    "-8s",
		PathID:   "swimPathGold",
		// Smaller upper figure-8, centred at (500,170)
		PathD: figure8Path(500, 170, 270, 90),
	},
	{
		Body:     coat.BodyPlatinum,
		Name:     "platinum",
		Variety:  "Ogon",
		Scale:    0.44,
		Duration: "37s",
		Begin:    "-30s",
		PathID:   "swimPathPlatinum",
		// Medium figure-8, offset left, centred at (460,210)
		PathD: figure8Path(460, 210, 320, 110),
	},
	{
		Body:     coat.BodyTea,
		Name:     "tea",
		Variety:  "Chagoi",
		Scale:    0.42,
		Duration: "29s",
		Begin:    "-13s",
		PathID:   "swimPathTea",
		// Medium figure-8, offset right, centred at (540,195)
		PathD: figure8Path(540, 195, 300, 100),
	},
	{
		Body:     coat.BodyWhite,
		Name:     "white",
		Variety:  "Kohaku",
		Scale:    0.44,
		Duration: "41s",
		Begin:    "-36s",
		PathID:   "swimPathWhite",
		// Large figure-8, slight downward offset, centred at (500,220)
		PathD: figure8Path(500, 220, 350, 118),
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
		gifBytes, err := output.EncodeScaledGIF(frames, 1)
		if err != nil {
			return nil, fmt.Errorf("encoding %s koi: %w", spec.Name, err)
		}
		result[spec.Body] = gifBytes
	}

	return result, nil
}

// GenerateAquariumSVG builds the self-contained animated SVG aquarium.
// koiGifs: pre-rendered GIF bytes per body. If nil, they are rendered from dev.
func GenerateAquariumSVG(dev *koi.KoiDevelopment, koiGifs interface{}) (string, error) {
	// Accept map[coat.KoiBody][]byte or nil
	var gifMap map[coat.KoiBody][]byte
	switch v := koiGifs.(type) {
	case map[coat.KoiBody][]byte:
		gifMap = v
	}

	if gifMap == nil {
		var err error
		gifMap, err = RenderAquariumGifs(dev)
		if err != nil {
			return "", err
		}
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %d %d" width="100%%" style="border-radius:16px; overflow:hidden; display:block; background:#050d1a;">
<defs>

  <!-- ═══ WATER GRADIENTS ═══ -->
  <radialGradient id="pondDeep" cx="50%%" cy="45%%" r="65%%">
    <stop offset="0%%" stop-color="#0a2040" stop-opacity="0.95"/>
    <stop offset="50%%" stop-color="#061428" stop-opacity="0.98"/>
    <stop offset="100%%" stop-color="#020810" stop-opacity="1"/>
  </radialGradient>

  <linearGradient id="surfaceShimmer" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#0ea5e9" stop-opacity="0.08"/>
    <stop offset="50%%" stop-color="#06b6d4" stop-opacity="0.04"/>
    <stop offset="100%%" stop-color="#0284c7" stop-opacity="0.10"/>
  </linearGradient>

  <radialGradient id="vignette" cx="50%%" cy="50%%" r="70%%">
    <stop offset="60%%" stop-color="#000000" stop-opacity="0"/>
    <stop offset="100%%" stop-color="#000000" stop-opacity="0.55"/>
  </radialGradient>

  <!-- ═══ FILTERS ═══ -->
  <filter id="causticGlow" x="-20%%" y="-20%%" width="140%%" height="140%%">
    <feGaussianBlur stdDeviation="6" result="blur"/>
    <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
  </filter>

  <filter id="padShadow" x="-20%%" y="-20%%" width="140%%" height="140%%">
    <feDropShadow dx="2" dy="4" stdDeviation="4" flood-color="#000b1f" flood-opacity="0.5"/>
  </filter>

  <!-- ═══ LILY PAD SYMBOLS ═══ -->
  <!-- Large organic lily pad — wedge notch, curved veins -->
  <g id="lilyLg">
    <path d="M 0 0 C 8 -6, 20 -12, 32 -10 C 42 -8, 46 2, 44 10 C 42 20, 36 28, 26 32 C 14 36, 2 34, -8 28 C -20 22, -28 10, -26 -2 C -24 -12, -14 -18, -6 -16 C -4 -16, -2 -14, 0 0 Z"
      fill="#1a4d2e" stroke="#2d7a46" stroke-width="1.2" opacity="0.92"/>
    <path d="M 0 0 L -4 -16" stroke="#3b9657" stroke-width="1" opacity="0.7"/>
    <path d="M 0 0 C 6 -4, 20 -8, 32 -10" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 0 0 C 10 2, 30 4, 44 10" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 0 0 C 8 10, 18 22, 26 32" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 0 0 C -4 10, -10 22, -8 28" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 0 0 C -10 2, -22 6, -26 -2" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 0 0 C -8 -6, -16 -12, -6 -16" stroke="#4ade80" stroke-width="0.9" fill="none" opacity="0.45"/>
    <path d="M 10 -6 C 16 -8, 22 -6, 24 -2" stroke="#86efac" stroke-width="1.2" fill="none" opacity="0.50"/>
  </g>

  <!-- Medium lily pad -->
  <g id="lilyMd">
    <path d="M 0 0 C 6 -4, 14 -8, 22 -7 C 30 -6, 32 2, 30 8 C 28 15, 22 20, 14 22 C 6 24, -2 22, -8 16 C -14 10, -16 2, -14 -4 C -12 -10, -6 -12, -2 -10 C -1 -10, 0 -6, 0 0 Z"
      fill="#235c37" stroke="#3b9657" stroke-width="1" opacity="0.90"/>
    <path d="M 0 0 C 4 -3, 14 -6, 22 -7" stroke="#4ade80" stroke-width="0.7" fill="none" opacity="0.40"/>
    <path d="M 0 0 C 8 2, 22 4, 30 8" stroke="#4ade80" stroke-width="0.7" fill="none" opacity="0.40"/>
    <path d="M 0 0 C 6 8, 12 16, 14 22" stroke="#4ade80" stroke-width="0.7" fill="none" opacity="0.40"/>
    <path d="M 0 0 C -6 6, -10 14, -8 16" stroke="#4ade80" stroke-width="0.7" fill="none" opacity="0.40"/>
    <path d="M 0 0 C -8 0, -14 -2, -14 -4" stroke="#4ade80" stroke-width="0.7" fill="none" opacity="0.40"/>
    <path d="M 6 -4 C 10 -6, 14 -4, 15 -2" stroke="#86efac" stroke-width="0.9" fill="none" opacity="0.45"/>
  </g>

  <!-- Small lily pad -->
  <g id="lilySm">
    <path d="M 0 0 C 4 -3, 10 -5, 15 -4 C 20 -3, 22 2, 20 6 C 18 10, 12 14, 6 14 C 0 14, -6 10, -8 5 C -10 0, -8 -4, -4 -6 C -2 -7, -1 -4, 0 0 Z"
      fill="#1e5c33" stroke="#2d8a4e" stroke-width="0.8" opacity="0.88"/>
    <path d="M 0 0 C 4 -2, 10 -4, 15 -4" stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
    <path d="M 0 0 C 6 4, 10 10, 6 14"  stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
    <path d="M 0 0 C -4 2, -8 4, -8 5"  stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
  </g>

  <!-- Lotus flower -->
  <g id="lotus">
    <path d="M 0 0 C -3 -6, -2 -14, 0 -16 C 2 -14, 3 -6, 0 0 Z" fill="#fbcfe8" stroke="#f9a8d4" stroke-width="0.5" opacity="0.88"/>
    <path d="M 0 0 C -3 -6, -2 -14, 0 -16 C 2 -14, 3 -6, 0 0 Z" fill="#fbcfe8" stroke="#f9a8d4" stroke-width="0.5" opacity="0.88" transform="rotate(72)"/>
    <path d="M 0 0 C -3 -6, -2 -14, 0 -16 C 2 -14, 3 -6, 0 0 Z" fill="#fbcfe8" stroke="#f9a8d4" stroke-width="0.5" opacity="0.88" transform="rotate(144)"/>
    <path d="M 0 0 C -3 -6, -2 -14, 0 -16 C 2 -14, 3 -6, 0 0 Z" fill="#fbcfe8" stroke="#f9a8d4" stroke-width="0.5" opacity="0.88" transform="rotate(216)"/>
    <path d="M 0 0 C -3 -6, -2 -14, 0 -16 C 2 -14, 3 -6, 0 0 Z" fill="#fbcfe8" stroke="#f9a8d4" stroke-width="0.5" opacity="0.88" transform="rotate(288)"/>
    <path d="M 0 0 C -2 -4, -1 -10, 0 -11 C 1 -10, 2 -4, 0 0 Z" fill="#fdf2f8" opacity="0.95" transform="rotate(36)"/>
    <path d="M 0 0 C -2 -4, -1 -10, 0 -11 C 1 -10, 2 -4, 0 0 Z" fill="#fdf2f8" opacity="0.95" transform="rotate(108)"/>
    <path d="M 0 0 C -2 -4, -1 -10, 0 -11 C 1 -10, 2 -4, 0 0 Z" fill="#fdf2f8" opacity="0.95" transform="rotate(180)"/>
    <path d="M 0 0 C -2 -4, -1 -10, 0 -11 C 1 -10, 2 -4, 0 0 Z" fill="#fdf2f8" opacity="0.95" transform="rotate(252)"/>
    <path d="M 0 0 C -2 -4, -1 -10, 0 -11 C 1 -10, 2 -4, 0 0 Z" fill="#fdf2f8" opacity="0.95" transform="rotate(324)"/>
    <circle cx="0" cy="0" r="4" fill="#fef08a"/>
    <circle cx="0" cy="0" r="2" fill="#fde047"/>
    <circle cx="0" cy="-3" r="0.8" fill="#f97316"/>
    <circle cx="2.8" cy="-1" r="0.8" fill="#f97316"/>
    <circle cx="1.8" cy="2.2" r="0.8" fill="#f97316"/>
    <circle cx="-1.8" cy="2.2" r="0.8" fill="#f97316"/>
    <circle cx="-2.8" cy="-1" r="0.8" fill="#f97316"/>
  </g>

  <!-- Lotus bud -->
  <g id="lotusBud">
    <path d="M 0 -10 C -4 -12, -6 -6, -4 0 C -2 4, 2 4, 4 0 C 6 -6, 4 -12, 0 -10 Z" fill="#f9a8d4" stroke="#ec4899" stroke-width="0.6"/>
    <path d="M 0 -8 C -2 -10, -3 -4, -2 0 C -1 3, 1 3, 2 0 C 3 -4, 2 -10, 0 -8 Z" fill="#fdf2f8" opacity="0.9"/>
    <line x1="0" y1="0" x2="0" y2="8" stroke="#2d7a46" stroke-width="1.2"/>
  </g>

  <!-- Algae wisp -->
  <g id="algaeWisp">
    <path d="M 0 0 C 4 -8, -4 -16, 2 -24 C 6 -18, -2 -10, 2 -4 C 4 0, 2 2, 0 0 Z" fill="#166534" opacity="0.65"/>
    <path d="M 0 0 C -6 -6, 2 -14, -2 -20 C -4 -14, 4 -8, 0 -4 Z" fill="#15803d" opacity="0.55"/>
  </g>

  <!-- ═══ FIGURE-8 SWIM PATHS ═══ -->
`, PondWidth, PondHeight))

	for _, spec := range DefaultFishSpecs {
		sb.WriteString(fmt.Sprintf(`  <path id="%s" d="%s" fill="none"/>
`, spec.PathID, spec.PathD))
	}

	sb.WriteString(`</defs>

<!-- ═══ POND BED ═══ -->
<rect width="100%" height="100%" fill="url(#pondDeep)"/>
<rect width="100%" height="100%" fill="url(#surfaceShimmer)"/>

<!-- ═══ WATER CAUSTICS ═══ -->
<g filter="url(#causticGlow)">
  <polygon points="60,0 180,0 320,420 160,420" fill="#38bdf8" opacity="0.07">
    <animate attributeName="opacity" values="0.04;0.13;0.07;0.11;0.04" dur="7s" repeatCount="indefinite"/>
  </polygon>
  <polygon points="380,0 500,0 580,420 380,420" fill="#06b6d4" opacity="0.06">
    <animate attributeName="opacity" values="0.03;0.11;0.05;0.09;0.03" dur="9s" begin="1.5s" repeatCount="indefinite"/>
  </polygon>
  <polygon points="720,0 840,0 940,420 820,420" fill="#38bdf8" opacity="0.05">
    <animate attributeName="opacity" values="0.03;0.10;0.06;0.08;0.03" dur="8s" begin="3s" repeatCount="indefinite"/>
  </polygon>
  <polygon points="240,0 280,0 340,420 240,420" fill="#bae6fd" opacity="0.05">
    <animate attributeName="opacity" values="0.02;0.09;0.03;0.07;0.02" dur="11s" begin="2s" repeatCount="indefinite"/>
  </polygon>

  <!-- Dancing light patches on pond floor -->
  <ellipse cx="160" cy="130" rx="22" ry="9" fill="#7dd3fc" opacity="0.14" transform="rotate(-15, 160, 130)">
    <animate attributeName="opacity" values="0.08;0.22;0.10;0.18;0.08" dur="4s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="22;30;20;28;22" dur="4s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="460" cy="80" rx="26" ry="10" fill="#38bdf8" opacity="0.12">
    <animate attributeName="opacity" values="0.06;0.20;0.08;0.16;0.06" dur="5s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="26;18;30;22;26" dur="5s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="740" cy="150" rx="18" ry="7" fill="#7dd3fc" opacity="0.13">
    <animate attributeName="opacity" values="0.07;0.21;0.09;0.17;0.07" dur="3.5s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="18;26;16;24;18" dur="3.5s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="300" cy="320" rx="24" ry="9" fill="#2dd4bf" opacity="0.11">
    <animate attributeName="opacity" values="0.05;0.17;0.07;0.14;0.05" dur="6s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="24;16;28;20;24" dur="6s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="650" cy="350" rx="20" ry="8" fill="#38bdf8" opacity="0.12">
    <animate attributeName="opacity" values="0.06;0.19;0.08;0.15;0.06" dur="4.5s" begin="1s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="20;28;18;26;20" dur="4.5s" begin="1s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="850" cy="260" rx="14" ry="6" fill="#7dd3fc" opacity="0.14">
    <animate attributeName="opacity" values="0.08;0.24;0.10;0.20;0.08" dur="3s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="14;20;12;18;14" dur="3s" repeatCount="indefinite"/>
  </ellipse>
  <ellipse cx="520" cy="250" rx="16" ry="6" fill="#93c5fd" opacity="0.10">
    <animate attributeName="opacity" values="0.05;0.16;0.07;0.13;0.05" dur="5.5s" begin="2s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="16;22;14;20;16" dur="5.5s" begin="2s" repeatCount="indefinite"/>
  </ellipse>
</g>

<!-- ═══ POND GRAVEL BED (subtle texture) ═══ -->
<g opacity="0.15" fill="#1e3a5f">
  <circle cx="80" cy="390" r="3"/><circle cx="150" cy="405" r="2"/><circle cx="220" cy="400" r="4"/>
  <circle cx="310" cy="410" r="2"/><circle cx="400" cy="395" r="3"/><circle cx="490" cy="408" r="2"/>
  <circle cx="560" cy="402" r="4"/><circle cx="650" cy="412" r="2"/><circle cx="730" cy="398" r="3"/>
  <circle cx="820" cy="406" r="2"/><circle cx="900" cy="394" r="4"/><circle cx="960" cy="410" r="2"/>
</g>

<!-- ═══ BOTANICAL LAYER ═══ -->
<g id="botanicals">
  <!-- Top-left cluster -->
  <g transform="translate(78, 68) rotate(-12)">
    <use href="#lilyLg" filter="url(#padShadow)"/>
  </g>
  <g transform="translate(42, 110) rotate(35) scale(0.75)">
    <use href="#lilyMd" filter="url(#padShadow)"/>
  </g>
  <g transform="translate(130, 55) rotate(5) scale(0.60)">
    <use href="#lilySm"/>
  </g>
  <g transform="translate(84, 54) scale(1.4)">
    <use href="#lotus"/>
  </g>
  <g transform="translate(52, 100) scale(1.0)">
    <use href="#lotusBud"/>
  </g>
  <g transform="translate(20, 180) rotate(-8) scale(1.2)">
    <use href="#algaeWisp"/>
  </g>
  <g transform="translate(10, 260) rotate(12) scale(0.9)">
    <use href="#algaeWisp"/>
  </g>

  <!-- Bottom-right cluster -->
  <g transform="translate(892, 318) rotate(-25) scale(1.15)">
    <use href="#lilyLg" filter="url(#padShadow)"/>
  </g>
  <g transform="translate(946, 274) rotate(80) scale(0.80)">
    <use href="#lilyMd" filter="url(#padShadow)"/>
  </g>
  <g transform="translate(850, 362) rotate(-5) scale(0.65)">
    <use href="#lilySm"/>
  </g>
  <g transform="translate(892, 300) scale(1.3)">
    <use href="#lotus"/>
  </g>
  <g transform="translate(958, 265) scale(0.9)">
    <use href="#lotusBud"/>
  </g>
  <g transform="translate(978, 140) rotate(5) scale(1.1)">
    <use href="#algaeWisp"/>
  </g>
  <g transform="translate(985, 310) rotate(-10) scale(0.85)">
    <use href="#algaeWisp"/>
  </g>

  <!-- Small corner accents -->
  <g transform="translate(870, 50) rotate(15) scale(0.65)">
    <use href="#lilyMd"/>
  </g>
  <g transform="translate(80, 362) rotate(20) scale(0.6)">
    <use href="#lilySm"/>
  </g>
</g>

<!-- ═══ SURFACE RIPPLES ═══ -->
<g fill="none" stroke="#7dd3fc" stroke-width="1.0">
  <circle cx="430" cy="95" r="3" opacity="0">
    <animate attributeName="r" values="3;60" dur="8s" repeatCount="indefinite" begin="0s"/>
    <animate attributeName="opacity" values="0.5;0" dur="8s" repeatCount="indefinite" begin="0s"/>
  </circle>
  <circle cx="430" cy="95" r="3" opacity="0">
    <animate attributeName="r" values="3;60" dur="8s" repeatCount="indefinite" begin="2.6s"/>
    <animate attributeName="opacity" values="0.5;0" dur="8s" repeatCount="indefinite" begin="2.6s"/>
  </circle>
  <circle cx="260" cy="210" r="3" opacity="0">
    <animate attributeName="r" values="3;70" dur="10s" repeatCount="indefinite" begin="1s"/>
    <animate attributeName="opacity" values="0.45;0" dur="10s" repeatCount="indefinite" begin="1s"/>
  </circle>
  <circle cx="750" cy="250" r="3" opacity="0">
    <animate attributeName="r" values="3;55" dur="7s" repeatCount="indefinite" begin="3s"/>
    <animate attributeName="opacity" values="0.50;0" dur="7s" repeatCount="indefinite" begin="3s"/>
  </circle>
  <circle cx="540" cy="350" r="3" opacity="0">
    <animate attributeName="r" values="3;50" dur="6s" repeatCount="indefinite" begin="2s"/>
    <animate attributeName="opacity" values="0.40;0" dur="6s" repeatCount="indefinite" begin="2s"/>
  </circle>
</g>

<!-- ═══ SWIMMING KOI — ORIGINAL PIXEL-ART GIF SPRITES ON FIGURE-8 PATHS ═══ -->
`)

	for _, spec := range DefaultFishSpecs {
		gifData := gifMap[spec.Body]
		b64 := base64.StdEncoding.EncodeToString(gifData)

		sb.WriteString(fmt.Sprintf(`
<!-- Koi: %s (%s) — figure-8 path -->
<g id="fish-%s">
  <animateMotion dur="%s" begin="%s" repeatCount="indefinite" rotate="auto" calcMode="linear">
    <mpath xlink:href="#%s"/>
  </animateMotion>
  <!-- Soft pond-floor shadow -->
  <ellipse cx="0" cy="20" rx="36" ry="12" fill="#000b1f" opacity="0.30">
    <animate attributeName="opacity" values="0.30;0.18;0.30" dur="3s" repeatCount="indefinite"/>
    <animate attributeName="rx" values="36;42;36" dur="3s" repeatCount="indefinite"/>
  </ellipse>
  <!-- Pixel-art koi sprite, centered, facing right (rotate="auto" steers it) -->
  <g transform="scale(%.3f)">
    <image xlink:href="data:image/gif;base64,%s"
      x="-128" y="-128" width="256" height="256"
      style="image-rendering:pixelated; image-rendering:crisp-edges;"/>
  </g>
</g>
`, spec.Variety, spec.Name, spec.Name, spec.Duration, spec.Begin, spec.PathID, spec.Scale, b64))
	}

	sb.WriteString(`
<!-- ═══ VIGNETTE ═══ -->
<rect width="100%" height="100%" fill="url(#vignette)"/>

<!-- ═══ GLASS FRAME ═══ -->
<rect width="100%" height="100%" rx="16" ry="16" fill="none"
  stroke="#38bdf8" stroke-width="1.2" stroke-opacity="0.28"/>
<rect x="1.5" y="1.5" width="997" height="417" rx="14.5" ry="14.5" fill="none"
  stroke="#0f172a" stroke-width="1.5" stroke-opacity="0.70"/>
<rect x="8" y="2" width="984" height="3" rx="1.5" fill="white" opacity="0.06"/>

<!-- ═══ STATUS BADGE ═══ -->
<g transform="translate(22, 28)" opacity="0.88">
  <circle cx="0" cy="0" r="3.5" fill="#34d399">
    <animate attributeName="opacity" values="0.3;1;0.3" dur="2s" repeatCount="indefinite"/>
  </circle>
  <text x="11" y="4"
    font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
    font-size="10.5" font-weight="600" fill="#94a3b8" letter-spacing="1.1">
    LIVE KOI POND · 6 NATIVE ECOSYSTEM FISH
  </text>
</g>
<g transform="translate(978, 28)" text-anchor="end" opacity="0.55">
  <text x="0" y="4"
    font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
    font-size="9.5" font-weight="400" fill="#64748b" letter-spacing="0.8">
    POWERED BY GIT-KOI (GO)
  </text>
</g>

</svg>
`)

	return sb.String(), nil
}
