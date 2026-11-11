package aquarium

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/reginaldalfret/git-koi/pkg/behavior"
	"github.com/reginaldalfret/git-koi/pkg/coat"
	"github.com/reginaldalfret/git-koi/pkg/koi"
	"github.com/reginaldalfret/git-koi/pkg/output"
	"github.com/reginaldalfret/git-koi/pkg/render"
)

// Pond constants
const (
	PondWidth  = 1000
	PondHeight = 420
)

// FishSpec defines trajectory, behavioral animation, and visual properties for each koi.
type FishSpec struct {
	Body      coat.KoiBody
	Name      string
	Variety   string
	Scale     float64
	Duration  string
	Begin     string
	PathID    string
	PathD     string
	WagAngle  string
	WagDur    string
	KeyTimes  string
	KeyPoints string
}

// Point represents a 2D coordinate in the pond.
type Point struct {
	X, Y float64
}

// catmullRomToBezier converts a closed sequence of waypoints into a perfectly continuous C1 cubic Bézier path.
// This guarantees zero sharp turns, zero tangent snapping, and silky-smooth orientation under rotate="auto".
func catmullRomToBezier(pts []Point) string {
	n := len(pts)
	if n < 3 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("M %.1f,%.1f", pts[0].X, pts[0].Y))

	for i := 0; i < n; i++ {
		p0 := pts[(i-1+n)%n]
		p1 := pts[i]
		p2 := pts[(i+1)%n]
		p3 := pts[(i+2)%n]

		t1x := (p2.X - p0.X) / 6.0
		t1y := (p2.Y - p0.Y) / 6.0
		t2x := (p3.X - p1.X) / 6.0
		t2y := (p3.Y - p1.Y) / 6.0

		c1x := p1.X + t1x
		c1y := p1.Y + t1y
		c2x := p2.X - t2x
		c2y := p2.Y - t2y

		sb.WriteString(fmt.Sprintf(" C %.1f,%.1f %.1f,%.1f %.1f,%.1f", c1x, c1y, c2x, c2y, p2.X, p2.Y))
	}
	sb.WriteString(" Z")
	return sb.String()
}

// buildLifeCyclePath constructs an organic, sweeping multi-stage journey for a fish that incorporates:
// 1. Straight horizontal cruise
// 2. Hover / calm drift pause zone (smooth steady glide without spinning)
// 3. Wide oval cruising loop ("0" shape)
// 4. Serpentine zig-zag / S-curves
// 5. Majestic central figure-8 (∞) loop
// 6. Smooth return glide
// All paths are bounded between x∈[185, 815] and y∈[112, 308] so NO fish swims above corner lotus pads.
func buildLifeCyclePath(seed int) string {
	flipX := 1.0
	flipY := 1.0
	if seed%2 == 1 {
		flipY = -1.0
	}
	if (seed/2)%2 == 1 {
		flipX = -1.0
	}

	s := float64(seed)
	dx := (math.Sin(s*1.7) * 35.0)
	dy := (math.Cos(s*2.1) * 20.0)
	scaleX := 0.92 + math.Sin(s*0.9)*0.08
	scaleY := 0.88 + math.Cos(s*1.3)*0.10

	baseWaypoints := []Point{
		// 1. Straight cruise
		{X: -20, Y: 0},
		{X: 100, Y: -5},
		{X: 210, Y: -10},

		// 2. Hover / calm drift pause zone
		{X: 250, Y: -15},
		{X: 265, Y: -25},

		// 3. Oval ("0") loop
		{X: 240, Y: -65},
		{X: 170, Y: -80},
		{X: 95, Y: -60},
		{X: 140, Y: -25},
		{X: 215, Y: -30},

		// 4. Serpentine Zig-Zag / S-curves
		{X: 160, Y: 10},
		{X: 60, Y: 40},
		{X: -40, Y: 5},
		{X: -130, Y: 40},
		{X: -220, Y: 10},

		// 5. Figure-8 ("∞") loop
		// First lobe
		{X: -160, Y: -45},
		{X: -90, Y: -65},
		{X: -30, Y: -35},
		{X: -90, Y: 0},
		// Second lobe
		{X: -160, Y: 25},
		{X: -230, Y: 45},
		{X: -285, Y: 25},
		{X: -285, Y: -30},
		{X: -230, Y: -55},
		{X: -160, Y: -25},

		// 6. Return glide to start
		{X: -90, Y: 0},
	}

	pts := make([]Point, len(baseWaypoints))
	for i, bp := range baseWaypoints {
		wx := 500.0 + (bp.X * flipX * scaleX) + dx
		wy := 210.0 + (bp.Y * flipY * scaleY) + dy

		if wx < 185.0 {
			wx = 185.0
		}
		if wx > 815.0 {
			wx = 815.0
		}
		if wy < 112.0 {
			wy = 112.0
		}
		if wy > 308.0 {
			wy = 308.0
		}

		pts[i] = Point{X: wx, Y: wy}
	}

	return catmullRomToBezier(pts)
}

// generateCaustics generates authentic animated underwater caustics:
// flowing refracted light ribbons and pulsing organic cellular dapples.
func generateCaustics() string {
	var sb strings.Builder

	// 1. Organic, undulating caustic light ribbons (refracted surface wave patterns)
	for i := 0; i < 4; i++ {
		yBase := 80.0 + float64(i)*85.0
		amp := 18.0
		dur := 6.0 + float64(i)*1.5

		var d1, d2 strings.Builder
		d1.WriteString(fmt.Sprintf("M -20,%.1f", yBase))
		d2.WriteString(fmt.Sprintf("M -20,%.1f", yBase+amp))
		for x := 0.0; x <= 1020.0; x += 120.0 {
			y1 := yBase + amp*math.Sin(x*0.015+float64(i))
			y2 := yBase - amp*math.Sin(x*0.015+float64(i)+0.5)
			d1.WriteString(fmt.Sprintf(" Q %.1f,%.1f %.1f,%.1f", x+60.0, y1+12.0, x+120.0, y1))
			d2.WriteString(fmt.Sprintf(" Q %.1f,%.1f %.1f,%.1f", x+60.0, y2-12.0, x+120.0, y2))
		}

		sb.WriteString(fmt.Sprintf(`  <!-- Refractive Wave Caustic Ribbon %d -->
  <path d="%s" fill="none" stroke="#7dd3fc" stroke-width="2.0" opacity="0.12">
    <animate attributeName="d" values="%s; %s; %s" dur="%.1fs" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="0.08;0.18;0.10;0.16;0.08" dur="%.1fs" repeatCount="indefinite"/>
  </path>
`, i+1, d1.String(), d1.String(), d2.String(), d1.String(), dur, dur))
	}

	// 2. Interlocking cellular focal nodes (bright light dapples on pond bottom)
	for row := 0; row < 3; row++ {
		for col := 0; col < 6; col++ {
			cx := 100.0 + float64(col)*155.0 + math.Sin(float64(row*2+col))*18.0
			cy := 90.0 + float64(row)*115.0 + math.Cos(float64(col*3))*14.0
			rx := 32.0 + math.Sin(float64(col))*6.0
			ry := 18.0 + math.Cos(float64(row))*4.0
			dur := 3.2 + math.Mod(float64(row*4+col*3), 2.5)

			sb.WriteString(fmt.Sprintf(`  <ellipse cx="%.1f" cy="%.1f" rx="%.1f" ry="%.1f" fill="#bae6fd" opacity="0.10">
    <animate attributeName="opacity" values="0.05;0.18;0.07;0.15;0.05" dur="%.1fs" repeatCount="indefinite"/>
    <animate attributeName="rx" values="%.1f;%.1f;%.1f" dur="%.1fs" repeatCount="indefinite"/>
  </ellipse>
`, cx, cy, rx, ry, dur, rx, rx*1.25, rx, dur+0.4))
		}
	}

	return sb.String()
}

// 12 Full-Lifecycle Paths: Every single fish performs Oval, Straight, Calm Hover, Zig-Zag, and Figure-8!
var (
	pathLife0  = buildLifeCyclePath(0)
	pathLife1  = buildLifeCyclePath(1)
	pathLife2  = buildLifeCyclePath(2)
	pathLife3  = buildLifeCyclePath(3)
	pathLife4  = buildLifeCyclePath(4)
	pathLife5  = buildLifeCyclePath(5)
	pathLife6  = buildLifeCyclePath(6)
	pathLife7  = buildLifeCyclePath(7)
	pathLife8  = buildLifeCyclePath(8)
	pathLife9  = buildLifeCyclePath(9)
	pathLife10 = buildLifeCyclePath(10)
	pathLife11 = buildLifeCyclePath(11)
)

// DefaultFishSpecs defines 12 koi (2 of every variety).
// Swimming speed is calm, steady, and realistic (90s - 128s per full pond cycle).
// Lateral body turning uses subtle, gentle angles (2.5° - 3.5°) over relaxed breathing cycles (2.2s - 2.8s).
var DefaultFishSpecs = []FishSpec{
	{
		Body:      coat.BodyWhite,
		Name:      "white-1",
		Variety:   "Kohaku",
		Scale:     0.45,
		Duration:  "108s",
		Begin:     "0s",
		PathID:    "pathLife0",
		PathD:     pathLife0,
		WagAngle:  "2.8",
		WagDur:    "2.5s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyGold,
		Name:      "gold-1",
		Variety:   "Yamabuki",
		Scale:     0.46,
		Duration:  "96s",
		Begin:     "-18s",
		PathID:    "pathLife1",
		PathD:     pathLife1,
		WagAngle:  "3.2",
		WagDur:    "2.3s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-1",
		Variety:   "Asagi",
		Scale:     0.43,
		Duration:  "112s",
		Begin:     "-36s",
		PathID:    "pathLife2",
		PathD:     pathLife2,
		WagAngle:  "3.0",
		WagDur:    "2.6s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyBlack,
		Name:      "black-1",
		Variety:   "Karasu",
		Scale:     0.42,
		Duration:  "124s",
		Begin:     "-54s",
		PathID:    "pathLife3",
		PathD:     pathLife3,
		WagAngle:  "2.5",
		WagDur:    "2.8s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyTea,
		Name:      "tea-1",
		Variety:   "Chagoi",
		Scale:     0.44,
		Duration:  "100s",
		Begin:     "-28s",
		PathID:    "pathLife4",
		PathD:     pathLife4,
		WagAngle:  "2.9",
		WagDur:    "2.4s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-1",
		Variety:   "Ogon",
		Scale:     0.43,
		Duration:  "106s",
		Begin:     "-72s",
		PathID:    "pathLife5",
		PathD:     pathLife5,
		WagAngle:  "2.7",
		WagDur:    "2.5s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyWhite,
		Name:      "white-2",
		Variety:   "Kohaku",
		Scale:     0.44,
		Duration:  "116s",
		Begin:     "-45s",
		PathID:    "pathLife6",
		PathD:     pathLife6,
		WagAngle:  "3.0",
		WagDur:    "2.6s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyGold,
		Name:      "gold-2",
		Variety:   "Yamabuki",
		Scale:     0.43,
		Duration:  "92s",
		Begin:     "-24s",
		PathID:    "pathLife7",
		PathD:     pathLife7,
		WagAngle:  "3.4",
		WagDur:    "2.2s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-2",
		Variety:   "Asagi",
		Scale:     0.42,
		Duration:  "110s",
		Begin:     "-62s",
		PathID:    "pathLife8",
		PathD:     pathLife8,
		WagAngle:  "2.8",
		WagDur:    "2.5s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyBlack,
		Name:      "black-2",
		Variety:   "Karasu",
		Scale:     0.41,
		Duration:  "128s",
		Begin:     "-80s",
		PathID:    "pathLife9",
		PathD:     pathLife9,
		WagAngle:  "2.6",
		WagDur:    "2.7s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyTea,
		Name:      "tea-2",
		Variety:   "Chagoi",
		Scale:     0.45,
		Duration:  "102s",
		Begin:     "-10s",
		PathID:    "pathLife10",
		PathD:     pathLife10,
		WagAngle:  "2.9",
		WagDur:    "2.4s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-2",
		Variety:   "Ogon",
		Scale:     0.42,
		Duration:  "114s",
		Begin:     "-32s",
		PathID:    "pathLife11",
		PathD:     pathLife11,
		WagAngle:  "2.8",
		WagDur:    "2.6s",
		KeyTimes:  "0; 0.20; 0.36; 0.52; 0.72; 0.88; 1",
		KeyPoints: "0; 0.22; 0.35; 0.50; 0.72; 0.88; 1",
	},
}

// RenderAquariumGifs renders animated GIFs for all 6 distinct koi coat varieties.
func RenderAquariumGifs(dev *koi.KoiDevelopment) (map[coat.KoiBody][]byte, error) {
	const restLoopFrames = 30
	result := make(map[coat.KoiBody][]byte)

	distinctBodies := []coat.KoiBody{
		coat.BodyWhite,
		coat.BodyGold,
		coat.BodyAsagi,
		coat.BodyBlack,
		coat.BodyTea,
		coat.BodyPlatinum,
	}

	for _, body := range distinctBodies {
		frames := make([]*output.KoiFrame, restLoopFrames)
		for i := uint32(0); i < restLoopFrames; i++ {
			frames[i] = render.RenderSwimFrame(dev, body, behavior.Rest, i)
		}
		gifBytes, err := output.EncodeScaledGIF(frames, 1)
		if err != nil {
			return nil, fmt.Errorf("encoding %s koi: %w", body.String(), err)
		}
		result[body] = gifBytes
	}

	return result, nil
}

// GenerateAquariumSVG builds the animated SVG aquarium containing all 12 diverse-swimming fishes.
func GenerateAquariumSVG(dev *koi.KoiDevelopment, koiGifs interface{}) (string, error) {
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

	causticsSVG := generateCaustics()

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
    <feGaussianBlur stdDeviation="4" result="blur"/>
    <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
  </filter>

  <filter id="padShadow" x="-20%%" y="-20%%" width="140%%" height="140%%">
    <feDropShadow dx="2" dy="4" stdDeviation="4" flood-color="#000b1f" flood-opacity="0.5"/>
  </filter>

  <!-- ═══ LILY PAD SYMBOLS ═══ -->
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

  <g id="lilySm">
    <path d="M 0 0 C 4 -3, 10 -5, 15 -4 C 20 -3, 22 2, 20 6 C 18 10, 12 14, 6 14 C 0 14, -6 10, -8 5 C -10 0, -8 -4, -4 -6 C -2 -7, -1 -4, 0 0 Z"
      fill="#1e5c33" stroke="#2d8a4e" stroke-width="0.8" opacity="0.88"/>
    <path d="M 0 0 C 4 -2, 10 -4, 15 -4" stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
    <path d="M 0 0 C 6 4, 10 10, 6 14" stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
    <path d="M 0 0 C -4 2, -8 4, -8 5" stroke="#4ade80" stroke-width="0.6" fill="none" opacity="0.38"/>
  </g>

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

  <g id="lotusBud">
    <path d="M 0 -10 C -4 -12, -6 -6, -4 0 C -2 4, 2 4, 4 0 C 6 -6, 4 -12, 0 -10 Z" fill="#f9a8d4" stroke="#ec4899" stroke-width="0.6"/>
    <path d="M 0 -8 C -2 -10, -3 -4, -2 0 C -1 3, 1 3, 2 0 C 3 -4, 2 -10, 0 -8 Z" fill="#fdf2f8" opacity="0.9"/>
    <line x1="0" y1="0" x2="0" y2="8" stroke="#2d7a46" stroke-width="1.2"/>
  </g>

  <g id="algaeWisp">
    <path d="M 0 0 C 4 -8, -4 -16, 2 -24 C 6 -18, -2 -10, 2 -4 C 4 0, 2 2, 0 0 Z" fill="#166534" opacity="0.65"/>
    <path d="M 0 0 C -6 -6, 2 -14, -2 -20 C -4 -14, 4 -8, 0 -4 Z" fill="#15803d" opacity="0.55"/>
  </g>

  <!-- ═══ 12 FULL LIFECYCLE PATHS: EVERY FISH PERFORMS OVAL, SPRINT, HOVER, ZIG-ZAG, FIG-8 ═══ -->
  <path id="pathLife0" d="%s" fill="none"/>
  <path id="pathLife1" d="%s" fill="none"/>
  <path id="pathLife2" d="%s" fill="none"/>
  <path id="pathLife3" d="%s" fill="none"/>
  <path id="pathLife4" d="%s" fill="none"/>
  <path id="pathLife5" d="%s" fill="none"/>
  <path id="pathLife6" d="%s" fill="none"/>
  <path id="pathLife7" d="%s" fill="none"/>
  <path id="pathLife8" d="%s" fill="none"/>
  <path id="pathLife9" d="%s" fill="none"/>
  <path id="pathLife10" d="%s" fill="none"/>
  <path id="pathLife11" d="%s" fill="none"/>

</defs>

<!-- ═══ LAYER 1: POND BED ═══ -->
<rect width="100%%" height="100%%" fill="url(#pondDeep)"/>
<rect width="100%%" height="100%%" fill="url(#surfaceShimmer)"/>

<!-- ═══ LAYER 2: NATURAL WATER CAUSTICS (REFRACTED LIGHT MESH & CELLULAR DAPPLES) ═══ -->
<g filter="url(#causticGlow)">
%s
</g>

<!-- ═══ LAYER 3: POND GRAVEL BED ═══ -->
<g opacity="0.15" fill="#1e3a5f">
  <circle cx="80" cy="390" r="3"/><circle cx="150" cy="405" r="2"/><circle cx="220" cy="400" r="4"/>
  <circle cx="310" cy="410" r="2"/><circle cx="400" cy="395" r="3"/><circle cx="490" cy="408" r="2"/>
  <circle cx="560" cy="402" r="4"/><circle cx="650" cy="412" r="2"/><circle cx="730" cy="398" r="3"/>
  <circle cx="820" cy="406" r="2"/><circle cx="900" cy="394" r="4"/><circle cx="960" cy="410" r="2"/>
</g>

<!-- ═══ LAYER 4: 12 SWIMMING KOI FISH (SWIMMING IN WATER BENEATH SURFACE PLANTS) ═══ -->
`, PondWidth, PondHeight,
		pathLife0, pathLife1, pathLife2, pathLife3,
		pathLife4, pathLife5, pathLife6, pathLife7,
		pathLife8, pathLife9, pathLife10, pathLife11,
		causticsSVG))

	for _, spec := range DefaultFishSpecs {
		gifData := gifMap[spec.Body]
		b64 := base64.StdEncoding.EncodeToString(gifData)

		motionAttrs := fmt.Sprintf(`dur="%s" begin="%s" repeatCount="indefinite" rotate="auto" calcMode="linear"`,
			spec.Duration, spec.Begin)
		if spec.KeyTimes != "" && spec.KeyPoints != "" {
			motionAttrs = fmt.Sprintf(`dur="%s" begin="%s" repeatCount="indefinite" rotate="auto" calcMode="linear" keyTimes="%s" keyPoints="%s"`,
				spec.Duration, spec.Begin, spec.KeyTimes, spec.KeyPoints)
		}

		sb.WriteString(fmt.Sprintf(`
<!-- Koi: %s (%s) — Full Life Cycle Movement (Oval, Sprint, Hover, Zig-Zag, Fig-8) -->
<g id="fish-%s">
  <animateMotion %s>
    <mpath xlink:href="#%s"/>
  </animateMotion>
  
  <g>
    <animateTransform attributeName="transform" type="rotate"
      values="-%s; %s; -%s" dur="%s" repeatCount="indefinite"/>
    
    <ellipse cx="0" cy="18" rx="34" ry="11" fill="#000b1f" opacity="0.28">
      <animate attributeName="opacity" values="0.28;0.16;0.28" dur="3s" repeatCount="indefinite"/>
      <animate attributeName="rx" values="34;40;34" dur="3s" repeatCount="indefinite"/>
    </ellipse>

    <g transform="scale(%.3f)">
      <image xlink:href="data:image/gif;base64,%s"
        x="-128" y="-128" width="256" height="256"
        style="image-rendering:pixelated; image-rendering:crisp-edges;"/>
    </g>
  </g>
</g>
`, spec.Variety, spec.Name, spec.Name, motionAttrs, spec.PathID, spec.WagAngle, spec.WagAngle, spec.WagAngle, spec.WagDur, spec.Scale, b64))
	}

	sb.WriteString(`
<!-- ═══ LAYER 5: FLOATING BOTANICALS (ON TOP OF WATER — FISH SWIM UNDERNEATH) ═══ -->
<g id="botanicals">
  <!-- Top-Left Lotus & Lily Cluster -->
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

  <!-- Bottom-Right Lotus & Lily Cluster -->
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

  <!-- Side Pond Accents -->
  <g transform="translate(870, 50) rotate(15) scale(0.65)">
    <use href="#lilyMd"/>
  </g>
  <g transform="translate(80, 362) rotate(20) scale(0.6)">
    <use href="#lilySm"/>
  </g>
</g>

<!-- ═══ LAYER 6: SURFACE RIPPLES ═══ -->
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

<!-- ═══ LAYER 7: VIGNETTE & FRAME ═══ -->
<rect width="100%" height="100%" fill="url(#vignette)"/>
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
    LIVE KOI POND · 12 NATIVE ECOSYSTEM FISH
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
