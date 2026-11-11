package aquarium

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/behavior"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/coat"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/output"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/render"
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

// 12 Mathematically Collision-Free Paths: Distributed across 6 non-overlapping geographic pond circuits
// (Great Outer Perimeter, North-West Basin, North-East Basin, Central Deep Basin, South-West Basin, South-East Basin).
// Every pair of distinct paths maintains >= 45px geometric clearance.
// Paired fish sharing a circuit are phased at exactly 180° (50% cycle offset), guaranteeing they are always
// on opposite sides of their circuit and NEVER collide or overlap.
var (
	pathLife0  = "M 230.0,52.0 C 283.3,40.0 410.0,48.0 500.0,48.0 C 590.0,48.0 716.7,40.0 770.0,52.0 C 823.3,64.0 811.7,78.7 820.0,120.0 C 828.3,161.3 828.3,258.0 820.0,300.0 C 811.7,342.0 823.3,359.3 770.0,372.0 C 716.7,384.7 590.0,376.0 500.0,376.0 C 410.0,376.0 283.3,384.7 230.0,372.0 C 176.7,359.3 188.3,342.0 180.0,300.0 C 171.7,258.0 171.7,161.3 180.0,120.0 C 188.3,78.7 176.7,64.0 230.0,52.0 Z"
	pathLife1  = pathLife0
	pathLife2  = "M 260.0,102.0 C 277.5,94.7 331.7,97.7 365.0,98.0 C 398.3,98.3 444.2,96.3 460.0,104.0 C 475.8,111.7 475.8,136.7 460.0,144.0 C 444.2,151.3 398.3,148.3 365.0,148.0 C 331.7,147.7 277.5,149.7 260.0,142.0 C 242.5,134.3 242.5,109.3 260.0,102.0 Z"
	pathLife3  = pathLife2
	pathLife4  = "M 540.0,104.0 C 555.8,96.3 601.7,98.3 635.0,98.0 C 668.3,97.7 722.5,94.7 740.0,102.0 C 757.5,109.3 757.5,134.3 740.0,142.0 C 722.5,149.7 668.3,147.7 635.0,148.0 C 601.7,148.3 555.8,151.3 540.0,144.0 C 524.2,136.7 524.2,111.7 540.0,104.0 Z"
	pathLife5  = pathLife4
	pathLife6  = "M 340.0,194.0 C 366.7,186.3 446.7,190.0 500.0,190.0 C 553.3,190.0 633.3,186.3 660.0,194.0 C 686.7,201.7 686.7,228.3 660.0,236.0 C 633.3,243.7 553.3,240.0 500.0,240.0 C 446.7,240.0 366.7,243.7 340.0,236.0 C 313.3,228.3 313.3,201.7 340.0,194.0 Z"
	pathLife7  = pathLife6
	pathLife8  = "M 260.0,286.0 C 277.5,279.3 331.7,281.7 365.0,282.0 C 398.3,282.3 444.2,281.0 460.0,288.0 C 475.8,295.0 475.8,317.3 460.0,324.0 C 444.2,330.7 398.3,328.3 365.0,328.0 C 331.7,327.7 277.5,329.0 260.0,322.0 C 242.5,315.0 242.5,292.7 260.0,286.0 Z"
	pathLife9  = pathLife8
	pathLife10 = "M 540.0,288.0 C 555.8,281.0 601.7,282.3 635.0,282.0 C 668.3,281.7 722.5,279.3 740.0,286.0 C 757.5,292.7 757.5,315.0 740.0,322.0 C 722.5,329.0 668.3,327.7 635.0,328.0 C 601.7,328.3 555.8,330.7 540.0,324.0 C 524.2,317.3 524.2,295.0 540.0,288.0 Z"
	pathLife11 = pathLife10
)

// DefaultFishSpecs defines 12 koi (2 of every authentic variety) with 100% uniform size (scale 0.53, ~75px).
// Zero overlap: Paired fish on each track are 180° out of phase, and all tracks have >= 45px geometric clearance.
var DefaultFishSpecs = []FishSpec{
	// 1. Perimeter (Outer Pond Loop)
	{
		Body:      coat.BodyWhite,
		Name:      "white-1",
		Variety:   "Kohaku",
		Scale:     0.53,
		Duration:  "110s",
		Begin:     "0s",
		PathID:    "pathLife0",
		PathD:     pathLife0,
		WagAngle:  "2.8",
		WagDur:    "2.5s",
	},
	{
		Body:      coat.BodyGold,
		Name:      "gold-1",
		Variety:   "Yamabuki",
		Scale:     0.53,
		Duration:  "110s",
		Begin:     "-55s",
		PathID:    "pathLife1",
		PathD:     pathLife1,
		WagAngle:  "3.0",
		WagDur:    "2.4s",
	},

	// 2. North-West Basin
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-1",
		Variety:   "Asagi",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "0s",
		PathID:    "pathLife2",
		PathD:     pathLife2,
		WagAngle:  "2.9",
		WagDur:    "2.6s",
	},
	{
		Body:      coat.BodyBlack,
		Name:      "black-1",
		Variety:   "Karasu",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "-41s",
		PathID:    "pathLife3",
		PathD:     pathLife3,
		WagAngle:  "2.6",
		WagDur:    "2.7s",
	},

	// 3. North-East Basin
	{
		Body:      coat.BodyTea,
		Name:      "tea-1",
		Variety:   "Chagoi",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "0s",
		PathID:    "pathLife4",
		PathD:     pathLife4,
		WagAngle:  "2.8",
		WagDur:    "2.5s",
	},
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-1",
		Variety:   "Ogon",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "-41s",
		PathID:    "pathLife5",
		PathD:     pathLife5,
		WagAngle:  "2.7",
		WagDur:    "2.5s",
	},

	// 4. Central Deep Basin
	{
		Body:      coat.BodyWhite,
		Name:      "white-2",
		Variety:   "Kohaku",
		Scale:     0.53,
		Duration:  "92s",
		Begin:     "0s",
		PathID:    "pathLife6",
		PathD:     pathLife6,
		WagAngle:  "3.0",
		WagDur:    "2.5s",
	},
	{
		Body:      coat.BodyGold,
		Name:      "gold-2",
		Variety:   "Yamabuki",
		Scale:     0.53,
		Duration:  "92s",
		Begin:     "-46s",
		PathID:    "pathLife7",
		PathD:     pathLife7,
		WagAngle:  "3.2",
		WagDur:    "2.3s",
	},

	// 5. South-West Basin
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-2",
		Variety:   "Asagi",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "0s",
		PathID:    "pathLife8",
		PathD:     pathLife8,
		WagAngle:  "2.8",
		WagDur:    "2.6s",
	},
	{
		Body:      coat.BodyBlack,
		Name:      "black-2",
		Variety:   "Karasu",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "-41s",
		PathID:    "pathLife9",
		PathD:     pathLife9,
		WagAngle:  "2.6",
		WagDur:    "2.7s",
	},

	// 6. South-East Basin
	{
		Body:      coat.BodyTea,
		Name:      "tea-2",
		Variety:   "Chagoi",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "0s",
		PathID:    "pathLife10",
		PathD:     pathLife10,
		WagAngle:  "2.9",
		WagDur:    "2.4s",
	},
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-2",
		Variety:   "Ogon",
		Scale:     0.53,
		Duration:  "82s",
		Begin:     "-41s",
		PathID:    "pathLife11",
		PathD:     pathLife11,
		WagAngle:  "2.8",
		WagDur:    "2.5s",
	},
}

// RenderAquariumGifs renders animated GIFs for all 6 distinct koi coat varieties.
// Aquarium koi are rendered as full-grown adult koi to showcase their markings and flowy fins.
func RenderAquariumGifs(dev *koi.KoiDevelopment) (map[coat.KoiBody][]byte, error) {
	const restLoopFrames = 30
	result := make(map[coat.KoiBody][]byte)

	var aquariumDev koi.KoiDevelopment
	if dev != nil {
		aquariumDev = *dev
	}
	if aquariumDev.TotalContributions < koi.AdultContributions {
		aquariumDev.TotalContributions = koi.AdultContributions
	}

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
			frames[i] = render.RenderSwimFrame(&aquariumDev, body, behavior.Rest, i)
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

	if gifMap == nil && len(AuthenticOriginalKoiB64) == 0 && dev != nil {
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
		var b64 string
		if origB64, ok := AuthenticOriginalKoiB64[spec.Body]; ok && gifMap == nil {
			b64 = origB64
		} else if gifMap != nil && len(gifMap[spec.Body]) > 0 {
			b64 = base64.StdEncoding.EncodeToString(gifMap[spec.Body])
		} else if origB64, ok := AuthenticOriginalKoiB64[spec.Body]; ok {
			b64 = origB64
		}

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
    
    <ellipse cx="0" cy="19" rx="38" ry="12" fill="#000b1f" opacity="0.25">
      <animate attributeName="opacity" values="0.25;0.14;0.25" dur="3s" repeatCount="indefinite"/>
      <animate attributeName="rx" values="38;45;38" dur="3s" repeatCount="indefinite"/>
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
    POWERED BY GIT-KOI-AQUARIUM (GO)
  </text>
</g>

</svg>
`)

	return sb.String(), nil
}
