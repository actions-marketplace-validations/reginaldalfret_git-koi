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
	KeyTimes  string // Optional SMIL keyTimes for dynamic acceleration/coasting/pause
	KeyPoints string // Optional SMIL keyPoints along path
}

// ovalPath generates a smooth closed oval ("0" shaped loop).
func ovalPath(cx, cy, rx, ry float64) string {
	const kappa = 0.5522847498
	ox := rx * kappa
	oy := ry * kappa
	return fmt.Sprintf("M %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f Z",
		cx+rx, cy,
		cx+rx, cy+oy, cx+ox, cy+ry, cx, cy+ry,
		cx-ox, cy+ry, cx-rx, cy+oy, cx-rx, cy,
		cx-rx, cy-oy, cx-ox, cy-ry, cx, cy-ry,
		cx+ox, cy-ry, cx+rx, cy-oy, cx+rx, cy)
}

// figure8Path generates a smooth closed figure-8 (lemniscate) with C1 continuity.
func figure8Path(cx, cy, rx, ry float64) string {
	const n = 12
	dt := 2.0 * math.Pi / float64(n)

	type pt struct{ x, y, dx, dy float64 }
	pts := make([]pt, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) * dt
		x := cx + rx*math.Sin(t)
		y := cy + (ry/2.0)*math.Sin(2.0*t)
		dx := rx * math.Cos(t) * dt
		dy := ry * math.Cos(2.0*t) * dt
		pts[i] = pt{x, y, dx, dy}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("M %.1f,%.1f", pts[0].x, pts[0].y))

	for i := 0; i < n; i++ {
		p0 := pts[i]
		p1 := pts[i+1]
		sb.WriteString(fmt.Sprintf(" C %.1f,%.1f %.1f,%.1f %.1f,%.1f",
			p0.x+p0.dx/3.0, p0.y+p0.dy/3.0, p1.x-p1.dx/3.0, p1.y-p1.dy/3.0, p1.x, p1.y))
	}
	sb.WriteString(" Z")
	return sb.String()
}

// 12 Diverse, Natural Trajectory Paths:
// Kept away from corner plants: Top-Left (x<160, y<140) and Bottom-Right (x>830, y>260).
var (
	// 1. Wide Center Oval ("0" shape): grand sweeping perimeter loop
	pathWideOval = ovalPath(500, 210, 310, 130)

	// 2. Center-Right Figure-8: graceful looping lemniscate
	pathCenterFig8 = figure8Path(550, 215, 230, 135)

	// 3. Meandering Zig-Zag: serpentine S-curve across middle and upper pond
	pathZigZag = "M 220,180 C 310,120 400,280 500,190 C 600,100 690,270 780,180 C 700,260 590,290 500,260 C 410,230 310,250 220,180 Z"

	// 4. Straight Horizontal Patrol with turning loops at each end
	pathStraightPatrol = "M 240,230 C 370,230 630,230 760,230 C 800,230 810,270 770,275 C 640,285 360,285 230,275 C 190,270 200,230 240,230 Z"

	// 5. Left Deep Oval ("0" shape): slow circling in open mid-left water
	pathLeftOval = ovalPath(340, 225, 150, 95)

	// 6. Right Deep Oval ("0" shape): circling in open mid-right water
	pathRightOval = ovalPath(660, 195, 150, 95)

	// 7. Upper Gliding Straight Path: slow drift across upper water column
	pathUpperGlide = "M 270,120 C 420,105 580,105 730,120 C 770,125 765,160 720,155 C 570,140 430,140 280,155 C 235,160 230,125 270,120 Z"

	// 8. Bottom-Left Figure-8: compact exploration in lower-left territory
	pathLowerFig8 = figure8Path(380, 275, 170, 90)

	// 9. Diagonal Wander: angled criss-cross drift through sun rays
	pathDiagonalWander = "M 300,150 C 440,190 560,240 700,280 C 750,295 760,250 710,225 C 570,175 450,125 310,95 C 260,85 250,135 300,150 Z"

	// 10. Central Soft Diamond / Rounded Square: mid-pond exploratory loop
	pathCentralDiamond = "M 500,125 C 620,140 680,180 660,210 C 640,255 580,290 500,295 C 420,290 360,255 340,210 C 320,180 380,140 500,125 Z"

	// 11. Subtle Hover / Station-Holding: nearly static action with gentle breathing drift
	pathStaticHover = "M 480,205 C 488,202 494,208 490,214 C 486,220 478,218 476,212 C 474,206 480,202 480,205 Z"

	// 12. Gentle Undulating Wave: flowing ripple trajectory through lower-center pond
	pathWaveMeander = "M 260,250 C 350,305 450,195 550,265 C 640,315 720,215 750,240 C 760,270 700,310 610,315 C 490,320 370,305 260,250 Z"
)

// DefaultFishSpecs defines 12 koi (2 of every variety) with diverse realistic swimming behaviors:
// Oval ("0"), Figure-8, Straight Patrol, Zig-Zag, Diagonal, Hover, and Wave.
var DefaultFishSpecs = []FishSpec{
	// ─── 1. Kohaku-1 (White/Red): Grand Wide Oval ("0" shape) ───
	{
		Body:      coat.BodyWhite,
		Name:      "white-1",
		Variety:   "Kohaku",
		Scale:     0.45,
		Duration:  "38s",
		Begin:     "0s",
		PathID:    "pathWideOval",
		PathD:     pathWideOval,
		WagAngle:  "8",
		WagDur:    "1.1s",
		KeyTimes:  "0; 0.25; 0.40; 0.70; 0.85; 1",
		KeyPoints: "0; 0.30; 0.38; 0.75; 0.82; 1",
	},

	// ─── 2. Yamabuki-1 (Gold): Center Figure-8 ───
	{
		Body:      coat.BodyGold,
		Name:      "gold-1",
		Variety:   "Yamabuki",
		Scale:     0.46,
		Duration:  "32s",
		Begin:     "-7s",
		PathID:    "pathCenterFig8",
		PathD:     pathCenterFig8,
		WagAngle:  "10",
		WagDur:    "0.9s",
		KeyTimes:  "0; 0.35; 0.50; 0.85; 1",
		KeyPoints: "0; 0.40; 0.48; 0.90; 1",
	},

	// ─── 3. Asagi-1 (Blue/Orange): Meandering Zig-Zag ───
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-1",
		Variety:   "Asagi",
		Scale:     0.43,
		Duration:  "35s",
		Begin:     "-12s",
		PathID:    "pathZigZag",
		PathD:     pathZigZag,
		WagAngle:  "9",
		WagDur:    "1.0s",
		KeyTimes:  "0; 0.30; 0.55; 0.80; 1",
		KeyPoints: "0; 0.35; 0.50; 0.85; 1",
	},

	// ─── 4. Karasu-1 (Black): Straight Horizontal Patrol ───
	{
		Body:      coat.BodyBlack,
		Name:      "black-1",
		Variety:   "Karasu",
		Scale:     0.42,
		Duration:  "42s",
		Begin:     "-18s",
		PathID:    "pathStraightPatrol",
		PathD:     pathStraightPatrol,
		WagAngle:  "6",
		WagDur:    "1.3s",
		KeyTimes:  "0; 0.40; 0.55; 0.90; 1",
		KeyPoints: "0; 0.46; 0.50; 0.96; 1", // coasts & pauses at turn
	},

	// ─── 5. Chagoi-1 (Tea): Left Deep Oval ("0" shape) ───
	{
		Body:      coat.BodyTea,
		Name:      "tea-1",
		Variety:   "Chagoi",
		Scale:     0.44,
		Duration:  "28s",
		Begin:     "-5s",
		PathID:    "pathLeftOval",
		PathD:     pathLeftOval,
		WagAngle:  "8",
		WagDur:    "1.15s",
		KeyTimes:  "0; 0.45; 0.60; 1",
		KeyPoints: "0; 0.50; 0.55; 1",
	},

	// ─── 6. Platinum-1 (Silver): Upper Gliding Straight Path ───
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-1",
		Variety:   "Ogon",
		Scale:     0.43,
		Duration:  "36s",
		Begin:     "-22s",
		PathID:    "pathUpperGlide",
		PathD:     pathUpperGlide,
		WagAngle:  "7",
		WagDur:    "1.25s",
		KeyTimes:  "0; 0.38; 0.52; 0.88; 1",
		KeyPoints: "0; 0.45; 0.50; 0.95; 1",
	},

	// ─── 7. Kohaku-2 (White/Red): Right Deep Oval ("0" shape) ───
	{
		Body:      coat.BodyWhite,
		Name:      "white-2",
		Variety:   "Kohaku",
		Scale:     0.43,
		Duration:  "30s",
		Begin:     "-15s",
		PathID:    "pathRightOval",
		PathD:     pathRightOval,
		WagAngle:  "9",
		WagDur:    "1.0s",
		KeyTimes:  "0; 0.35; 0.55; 0.85; 1",
		KeyPoints: "0; 0.40; 0.50; 0.90; 1",
	},

	// ─── 8. Yamabuki-2 (Gold): Bottom-Left Figure-8 ───
	{
		Body:      coat.BodyGold,
		Name:      "gold-2",
		Variety:   "Yamabuki",
		Scale:     0.42,
		Duration:  "26s",
		Begin:     "-9s",
		PathID:    "pathLowerFig8",
		PathD:     pathLowerFig8,
		WagAngle:  "10",
		WagDur:    "0.95s",
		KeyTimes:  "0; 0.30; 0.50; 0.80; 1",
		KeyPoints: "0; 0.36; 0.48; 0.86; 1",
	},

	// ─── 9. Asagi-2 (Blue/Orange): Diagonal Wander Criss-Cross ───
	{
		Body:      coat.BodyAsagi,
		Name:      "asagi-2",
		Variety:   "Asagi",
		Scale:     0.41,
		Duration:  "34s",
		Begin:     "-20s",
		PathID:    "pathDiagonalWander",
		PathD:     pathDiagonalWander,
		WagAngle:  "8",
		WagDur:    "1.1s",
		KeyTimes:  "0; 0.42; 0.58; 1",
		KeyPoints: "0; 0.48; 0.52; 1",
	},

	// ─── 10. Karasu-2 (Black): Central Soft Diamond Loop ───
	{
		Body:      coat.BodyBlack,
		Name:      "black-2",
		Variety:   "Karasu",
		Scale:     0.44,
		Duration:  "40s",
		Begin:     "-27s",
		PathID:    "pathCentralDiamond",
		PathD:     pathCentralDiamond,
		WagAngle:  "7",
		WagDur:    "1.35s",
		KeyTimes:  "0; 0.28; 0.50; 0.78; 1",
		KeyPoints: "0; 0.32; 0.48; 0.82; 1",
	},

	// ─── 11. Tea-2 (Chagoi): Static Station-Holding (Gentle Breathing Hover) ───
	{
		Body:      coat.BodyTea,
		Name:      "tea-2",
		Variety:   "Chagoi",
		Scale:     0.45,
		Duration:  "18s",
		Begin:     "0s",
		PathID:    "pathStaticHover",
		PathD:     pathStaticHover,
		WagAngle:  "4", // subtle resting fin & tail pulse
		WagDur:    "2.2s",
		KeyTimes:  "0; 0.50; 1",
		KeyPoints: "0; 0.50; 1",
	},

	// ─── 12. Platinum-2 (Silver): Gentle Undulating Wave Meander ───
	{
		Body:      coat.BodyPlatinum,
		Name:      "platinum-2",
		Variety:   "Ogon",
		Scale:     0.42,
		Duration:  "33s",
		Begin:     "-11s",
		PathID:    "pathWaveMeander",
		PathD:     pathWaveMeander,
		WagAngle:  "8",
		WagDur:    "1.1s",
		KeyTimes:  "0; 0.35; 0.65; 1",
		KeyPoints: "0; 0.40; 0.60; 1",
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

  <!-- ═══ 12 DIVERSE MOTION PATHS (OVALS, FIG-8s, ZIG-ZAGS, PATROLS, HOVER) ═══ -->
  <path id="pathWideOval" d="%s" fill="none"/>
  <path id="pathCenterFig8" d="%s" fill="none"/>
  <path id="pathZigZag" d="%s" fill="none"/>
  <path id="pathStraightPatrol" d="%s" fill="none"/>
  <path id="pathLeftOval" d="%s" fill="none"/>
  <path id="pathRightOval" d="%s" fill="none"/>
  <path id="pathUpperGlide" d="%s" fill="none"/>
  <path id="pathLowerFig8" d="%s" fill="none"/>
  <path id="pathDiagonalWander" d="%s" fill="none"/>
  <path id="pathCentralDiamond" d="%s" fill="none"/>
  <path id="pathStaticHover" d="%s" fill="none"/>
  <path id="pathWaveMeander" d="%s" fill="none"/>

</defs>

<!-- ═══ LAYER 1: POND BED ═══ -->
<rect width="100%%" height="100%%" fill="url(#pondDeep)"/>
<rect width="100%%" height="100%%" fill="url(#surfaceShimmer)"/>

<!-- ═══ LAYER 2: WATER CAUSTICS (ANGLED SUN SHAFTS & FLOOR DAPPLES) ═══ -->
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
  <ellipse cx="300" cy="320" rx="24" ry="9" fill="#2dd4bf" opacity="0.11" transform="rotate(20, 300, 320)">
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

<!-- ═══ LAYER 3: POND GRAVEL BED ═══ -->
<g opacity="0.15" fill="#1e3a5f">
  <circle cx="80" cy="390" r="3"/><circle cx="150" cy="405" r="2"/><circle cx="220" cy="400" r="4"/>
  <circle cx="310" cy="410" r="2"/><circle cx="400" cy="395" r="3"/><circle cx="490" cy="408" r="2"/>
  <circle cx="560" cy="402" r="4"/><circle cx="650" cy="412" r="2"/><circle cx="730" cy="398" r="3"/>
  <circle cx="820" cy="406" r="2"/><circle cx="900" cy="394" r="4"/><circle cx="960" cy="410" r="2"/>
</g>

<!-- ═══ LAYER 4: 12 SWIMMING KOI FISH (SWIMMING IN WATER BENEATH SURFACE PLANTS) ═══ -->
`, PondWidth, PondHeight,
		pathWideOval, pathCenterFig8, pathZigZag, pathStraightPatrol,
		pathLeftOval, pathRightOval, pathUpperGlide, pathLowerFig8,
		pathDiagonalWander, pathCentralDiamond, pathStaticHover, pathWaveMeander))

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
<!-- Koi: %s (%s) — Diverse Realistic Trajectory -->
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
