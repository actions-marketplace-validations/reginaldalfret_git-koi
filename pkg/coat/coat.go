package coat

import (
	"math"
	"sort"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/markings"
)

type RGBA [4]byte

var (
	WhiteRamp = []RGBA{
		{90, 105, 136, 255},
		{139, 155, 180, 255},
		{192, 203, 220, 255},
		{255, 255, 255, 255},
	}
	RedRamp = []RGBA{
		{162, 38, 51, 255},
		{228, 59, 68, 255},
		{240, 96, 79, 255},
		{246, 117, 122, 255},
	}
	BlackRamp = []RGBA{
		{24, 20, 37, 255},
		{38, 43, 68, 255},
		{58, 68, 102, 255},
	}
	YellowRamp = []RGBA{
		{215, 118, 67, 255},
		{254, 174, 52, 255},
		{254, 231, 97, 255},
	}
	OrangeRamp = []RGBA{
		{190, 74, 47, 255},
		{247, 118, 34, 255},
		{254, 174, 52, 255},
	}
	BlueRamp = []RGBA{
		{58, 68, 102, 255},
		{18, 78, 137, 255},
		{0, 153, 219, 255},
	}
	BlueNetRamp = []RGBA{
		{38, 43, 68, 255},
	}
	PaleRedRamp = []RGBA{
		{194, 133, 105, 255},
		{232, 183, 150, 255},
		{234, 212, 170, 255},
	}
	PaleBlackRamp = []RGBA{
		{90, 105, 136, 255},
		{139, 155, 180, 255},
		{192, 203, 220, 255},
	}
	PaleYellowRamp = []RGBA{
		{232, 183, 150, 255},
		{234, 212, 170, 255},
		{255, 255, 255, 255},
	}
	PaleBlueRamp = []RGBA{
		{139, 155, 180, 255},
		{192, 203, 220, 255},
		{255, 255, 255, 255},
	}
	PaleNetRamp = []RGBA{
		{139, 155, 180, 255},
	}
	StreakGoldRamp = []RGBA{
		{247, 118, 34, 255},
		{254, 174, 52, 255},
		{254, 231, 97, 255},
	}
	GlintColor = RGBA{255, 255, 255, 255}
	AsagiRamp  = []RGBA{
		{38, 43, 68, 255},
		{58, 68, 102, 255},
		{90, 105, 136, 255},
		{139, 155, 180, 255},
	}
	AsagiNetRamp = []RGBA{
		{58, 68, 102, 255},
	}
	TeaRamp = []RGBA{
		{62, 39, 49, 255},
		{115, 62, 57, 255},
		{184, 111, 80, 255},
		{228, 166, 114, 255},
	}
	GoldRamp = []RGBA{
		{115, 62, 57, 255},
		{215, 118, 67, 255},
		{254, 174, 52, 255},
		{254, 231, 97, 255},
		{255, 255, 255, 255},
	}
	PlatinumRamp = []RGBA{
		{90, 105, 136, 255},
		{139, 155, 180, 255},
		{192, 203, 220, 255},
		{244, 247, 251, 255},
		{255, 255, 255, 255},
	}
	BrightYellowRamp = []RGBA{
		{215, 118, 67, 255},
		{254, 174, 52, 255},
		{254, 231, 97, 255},
		{255, 243, 176, 255},
	}
	SkyRamp = []RGBA{
		{18, 78, 137, 255},
		{0, 153, 219, 255},
		{44, 232, 245, 255},
	}
	RustRamp = []RGBA{
		{115, 62, 57, 255},
		{190, 74, 47, 255},
		{215, 118, 67, 255},
	}
	UmberRamp = []RGBA{
		{24, 20, 37, 255},
		{62, 39, 49, 255},
		{115, 62, 57, 255},
	}
	SandRamp = []RGBA{
		{184, 111, 80, 255},
		{228, 166, 114, 255},
		{234, 212, 170, 255},
	}
	ClayRamp = []RGBA{
		{190, 74, 47, 255},
		{215, 118, 67, 255},
		{228, 166, 114, 255},
	}
	SlateRamp = []RGBA{
		{38, 43, 68, 255},
		{58, 68, 102, 255},
		{90, 105, 136, 255},
	}
	BronzeRamp = []RGBA{
		{62, 39, 49, 255},
		{115, 62, 57, 255},
		{160, 106, 58, 255},
		{201, 154, 90, 255},
	}
	MetalRedRamp = []RGBA{
		{162, 38, 51, 255},
		{247, 118, 34, 255},
		{254, 174, 52, 255},
		{255, 255, 255, 255},
	}
	MetalOrangeRamp = []RGBA{
		{190, 74, 47, 255},
		{247, 118, 34, 255},
		{254, 174, 52, 255},
		{255, 255, 255, 255},
	}
	CopperRamp = []RGBA{
		{115, 62, 57, 255},
		{184, 111, 80, 255},
		{228, 166, 114, 255},
		{255, 255, 255, 255},
	}
	MetalBlueRamp = []RGBA{
		{18, 78, 137, 255},
		{0, 153, 219, 255},
		{139, 155, 180, 255},
		{255, 255, 255, 255},
	}
	SteelRamp = []RGBA{
		{38, 43, 68, 255},
		{58, 68, 102, 255},
		{90, 105, 136, 255},
		{192, 203, 220, 255},
	}
)

const (
	WhitePatchScore        = 0.5
	PatchCoverage          = 0.58
	PatchSpanStart         = 0.02
	PatchSpanEnd           = 0.86
	VarietyPull            = 0.25
	VarietyVoteWeight      = 0.16 * VarietyPull
	SpotSpanStart          = 0.12
	SpotSpanEnd            = 0.8
	SpotLateralLimit       = 0.8
	FinRootTintFraction    = 0.4 * VarietyPull
	StreakFinStartFraction = 0.55
	DullSaturation         = 0.3
	VividSaturationGain    = 0.3
	DullDarkening          = 0.15
	VividBrightening       = 0.12
)

type KoiBody int

const (
	BodyWhite KoiBody = iota
	BodyBlack
	BodyAsagi
	BodyTea
	BodyGold
	BodyPlatinum
)

func ParseKoiBody(name string) (KoiBody, bool) {
	switch name {
	case "white":
		return BodyWhite, true
	case "black":
		return BodyBlack, true
	case "asagi":
		return BodyAsagi, true
	case "tea":
		return BodyTea, true
	case "gold":
		return BodyGold, true
	case "platinum":
		return BodyPlatinum, true
	default:
		return BodyWhite, false
	}
}

func (kb KoiBody) String() string {
	switch kb {
	case BodyWhite:
		return "white"
	case BodyBlack:
		return "black"
	case BodyAsagi:
		return "asagi"
	case BodyTea:
		return "tea"
	case BodyGold:
		return "gold"
	case BodyPlatinum:
		return "platinum"
	default:
		return "white"
	}
}

func (kb KoiBody) SelectGroundRamp() []RGBA {
	switch kb {
	case BodyWhite:
		return WhiteRamp
	case BodyBlack:
		return BlackRamp
	case BodyAsagi:
		return AsagiRamp
	case BodyTea:
		return TeaRamp
	case BodyGold:
		return GoldRamp
	case BodyPlatinum:
		return PlatinumRamp
	default:
		return WhiteRamp
	}
}

func (kb KoiBody) SelectFinRamp() []RGBA {
	switch kb {
	case BodyWhite, BodyBlack:
		return WhiteRamp
	case BodyAsagi:
		return OrangeRamp
	default:
		return kb.SelectGroundRamp()
	}
}

func (kb KoiBody) SelectMouthRimColor() [4]byte {
	ground := kb.SelectGroundRamp()
	if kb == BodyBlack {
		return [4]byte(ground[len(ground)-1])
	}
	return [4]byte(ground[len(ground)-2])
}

func (kb KoiBody) SelectStreakFinRamp() []RGBA {
	switch kb {
	case BodyAsagi, BodyGold:
		return RedRamp
	default:
		return StreakGoldRamp
	}
}

func (kb KoiBody) CalculateMetallicSheen(alongFraction, lateralFraction float64) float64 {
	const backGlow = 0.1
	const bandAlongFraction = 0.35
	const bandHalfWidthFraction = 0.06
	const bandBrightness = 0.3

	if kb != BodyGold && kb != BodyPlatinum {
		return 0.0
	}
	band := 0.0
	if math.Abs(alongFraction-bandAlongFraction) < bandHalfWidthFraction {
		band = bandBrightness
	}
	return backGlow*(1.0-math.Abs(lateralFraction)) + band
}

type Pigment int

const (
	PigmentBeni Pigment = iota
	PigmentSumi
	PigmentKi
	PigmentOrenji
	PigmentAi
)

func (kb KoiBody) SelectPigmentRamp(p Pigment) []RGBA {
	switch kb {
	case BodyBlack:
		switch p {
		case PigmentSumi:
			return WhiteRamp
		case PigmentKi:
			return BrightYellowRamp
		case PigmentAi:
			return SkyRamp
		}
	case BodyAsagi:
		switch p {
		case PigmentAi:
			return WhiteRamp
		case PigmentKi:
			return BrightYellowRamp
		}
	case BodyTea:
		switch p {
		case PigmentBeni:
			return RustRamp
		case PigmentSumi:
			return UmberRamp
		case PigmentKi:
			return SandRamp
		case PigmentOrenji:
			return ClayRamp
		case PigmentAi:
			return SlateRamp
		}
	case BodyGold:
		switch p {
		case PigmentBeni:
			return MetalRedRamp
		case PigmentSumi:
			return BronzeRamp
		case PigmentKi:
			return PlatinumRamp
		case PigmentOrenji:
			return CopperRamp
		case PigmentAi:
			return MetalBlueRamp
		}
	case BodyPlatinum:
		switch p {
		case PigmentBeni:
			return MetalOrangeRamp
		case PigmentSumi:
			return SteelRamp
		case PigmentKi:
			return GoldRamp
		case PigmentOrenji:
			return CopperRamp
		case PigmentAi:
			return MetalBlueRamp
		}
	}

	switch p {
	case PigmentBeni:
		return RedRamp
	case PigmentSumi:
		return BlackRamp
	case PigmentKi:
		return YellowRamp
	case PigmentOrenji:
		return OrangeRamp
	case PigmentAi:
		return BlueRamp
	default:
		return BlackRamp
	}
}

type CoatPaint struct {
	ShadeRamp       []RGBA
	BrightnessShift float64
	Saturation      float64
	GlintColor      *[4]byte
}

func plainPaint(ramp []RGBA) CoatPaint {
	return CoatPaint{
		ShadeRamp:       ramp,
		BrightnessShift: 0.0,
		Saturation:      1.0,
		GlintColor:      nil,
	}
}

type BodySample struct {
	AlongFraction   float64
	LateralFraction float64
	AreaWeight      float64
}

type PigmentShare struct {
	Pigment Pigment
	Share   float64
}

type KoiCoat struct {
	IdentitySeed  uint64
	KoiBody       KoiBody
	PigmentShares []PigmentShare
	PatchBiases   []float64
	SpotThreshold float64
	ColorFraction float64
	StreakMarks   markings.StreakMarks
}

func NewKoiCoat(dev *koi.KoiDevelopment, body KoiBody, calcHalfWidth func(float64) float64) *KoiCoat {
	seed := dev.IdentitySeed
	shares := calculatePigmentShares(dev)
	samples := sampleBody(calcHalfWidth)

	colorFrac := koi.CalculateGrowthFraction(dev.TotalContributions) / koi.CalculateGrowthFraction(koi.FullColorContributions)
	if colorFrac > 1.0 {
		colorFrac = 1.0
	}

	return &KoiCoat{
		IdentitySeed:  seed,
		KoiBody:       body,
		PatchBiases:   fitPatchBiases(shares, seed, samples),
		SpotThreshold: findSpotThreshold(shares, seed, samples),
		PigmentShares: shares,
		ColorFraction: colorFrac,
		StreakMarks:   markings.CalculateStreakMarks(dev),
	}
}

func (kc *KoiCoat) SelectBodyPaint(along, lateral, bodyLen, halfWidth, dither float64) CoatPaint {
	const asagiHeadFraction = 0.11
	isOnNet := func() bool {
		return isOnScaleNet(along, lateral, bodyLen, halfWidth)
	}

	p, hasPigment := kc.selectBodyPigment(along, lateral)
	if !hasPigment {
		var ramp []RGBA
		if kc.KoiBody == BodyAsagi && along <= asagiHeadFraction {
			ramp = WhiteRamp
		} else if kc.KoiBody == BodyAsagi && isOnNet() {
			ramp = AsagiNetRamp
		} else {
			ramp = kc.KoiBody.SelectGroundRamp()
		}
		paint := plainPaint(ramp)
		paint.BrightnessShift = kc.KoiBody.CalculateMetallicSheen(along, lateral)
		return paint
	}

	pMap := kc.KoiBody.SelectPigmentRamp(p)
	var basePigment *Pigment
	if len(kc.PigmentShares) > 0 {
		basePigment = &kc.PigmentShares[0].Pigment
	}

	var shadeRamp []RGBA
	if p == PigmentAi && sameRamp(pMap, BlueRamp) && isOnNet() {
		shadeRamp = BlueNetRamp
	} else if (p == PigmentKi || p == PigmentOrenji) && basePigment != nil && *basePigment == p && isOnNet() {
		shadeRamp = pMap[:1]
	} else {
		shadeRamp = pMap
	}

	isGlinting := kc.isGlinting(along, lateral, bodyLen, halfWidth)
	var glint *[4]byte
	if isGlinting {
		g := [4]byte(GlintColor)
		glint = &g
	}

	return CoatPaint{
		ShadeRamp:       kc.developRamp(p, shadeRamp, dither),
		BrightnessShift: kc.calculateVibrancyShift(p),
		Saturation:      kc.calculateVibrancySaturation(),
		GlintColor:      glint,
	}
}

func (kc *KoiCoat) SelectFinPaint(along, lateral float64, isOnRim bool, dither float64) CoatPaint {
	if dither < kc.StreakMarks.StreakFinStrength && (isOnRim || along > StreakFinStartFraction) {
		return plainPaint(kc.KoiBody.SelectStreakFinRamp())
	}
	plainFin := kc.KoiBody.SelectFinRamp()
	if len(kc.PigmentShares) > 2 {
		tertiary := kc.PigmentShares[2].Pigment
		if along < FinRootTintFraction && math.Abs(lateral) < 0.85 {
			tintRamp := kc.developRamp(tertiary, kc.KoiBody.SelectPigmentRamp(tertiary), dither)
			if sameRamp(tintRamp, plainFin) {
				return plainPaint(tintRamp)
			}
			p := plainPaint(tintRamp)
			p.Saturation = kc.calculateVibrancySaturation()
			return p
		}
	}
	return plainPaint(plainFin)
}

func (kc *KoiCoat) calculateVibrancyShift(p Pigment) float64 {
	vib := kc.StreakMarks.ColorVibrancy
	if vib < 1.0 {
		return -DullDarkening * (1.0 - vib)
	} else if p == PigmentSumi {
		return 0.0
	}
	return VividBrightening * (vib - 1.0)
}

func (kc *KoiCoat) calculateVibrancySaturation() float64 {
	vib := kc.StreakMarks.ColorVibrancy
	if vib < 1.0 {
		return DullSaturation + (1.0-DullSaturation)*vib
	}
	return 1.0 + VividSaturationGain*(vib-1.0)
}

func (kc *KoiCoat) isGlinting(along, lateral, bodyLen, halfWidth float64) bool {
	const glintSpanStart = 0.14
	const glintSpanEnd = 0.74
	const glintLateralLimit = 0.86
	const fullGlintShare = 0.36

	glimmer := kc.StreakMarks.GlimmerStrength
	if glimmer <= 0.0 || along < glintSpanStart || along >= glintSpanEnd || math.Abs(lateral) >= glintLateralLimit {
		return false
	}

	scaleRow := (lateral + 1.0) * halfWidth / 2.0
	scaleCol := along*bodyLen/2.4 + math.Mod(math.Floor(scaleRow), 2.0)*0.5

	hash := hashLatticePoint(kc.IdentitySeed^0x6119, int64(math.Floor(scaleCol)), int64(math.Floor(scaleRow)))
	rowFract := scaleRow - math.Floor(scaleRow)
	colFract := scaleCol - math.Floor(scaleCol)

	return hash < glimmer*fullGlintShare && rowFract < 0.5 && colFract < 0.42
}

func (kc *KoiCoat) selectBodyPigment(along, lateral float64) (Pigment, bool) {
	varietyPigment, hasVariety := kc.selectVarietyPigment(along, lateral)
	isInPatchSpan := along >= PatchSpanStart && along < PatchSpanEnd

	var selectedPigment *Pigment
	var noneVariety float64
	if !hasVariety {
		noneVariety = 1.0
	}
	bestScore := (1.0-VarietyVoteWeight)*WhitePatchScore + VarietyVoteWeight*noneVariety

	for i, ps := range kc.PigmentShares {
		bias := 0.0
		if i < len(kc.PatchBiases) {
			bias = kc.PatchBiases[i]
		}
		patchScore := 0.0
		if isInPatchSpan {
			patchScore = samplePatchScore(kc.IdentitySeed, i, along, lateral) + bias
		}
		var isSameVariety float64
		if hasVariety && varietyPigment == ps.Pigment {
			isSameVariety = 1.0
		}
		blended := (1.0-VarietyVoteWeight)*patchScore + VarietyVoteWeight*isSameVariety
		if blended > bestScore {
			bestScore = blended
			p := ps.Pigment
			selectedPigment = &p
		}
	}

	if selectedPigment != nil {
		return *selectedPigment, true
	}
	return PigmentBeni, false
}

func (kc *KoiCoat) selectVarietyPigment(along, lateral float64) (Pigment, bool) {
	if len(kc.PigmentShares) == 0 {
		return PigmentBeni, false
	}
	base := kc.PigmentShares[0].Pigment
	var secondary *Pigment
	if len(kc.PigmentShares) > 1 {
		secondary = &kc.PigmentShares[1].Pigment
	}

	hasAsagiSides := base == PigmentAi && secondary != nil &&
		(*secondary == PigmentBeni || *secondary == PigmentOrenji || *secondary == PigmentKi)

	if secondary != nil && !hasAsagiSides && along >= SpotSpanStart && along < SpotSpanEnd &&
		math.Abs(lateral) < SpotLateralLimit &&
		sampleSpotNoise(kc.IdentitySeed, along, lateral) > kc.SpotThreshold {
		return *secondary, true
	}

	switch base {
	case PigmentBeni:
		if isInRedPlate(kc.IdentitySeed, along, lateral) {
			return PigmentBeni, true
		}
		return PigmentBeni, false
	case PigmentSumi:
		isLightning := along < 0.13 && math.Abs(lateral-0.25*math.Sin(along*55.0)) < 0.14
		if isLightning || sampleFractalNoise(kc.IdentitySeed, along*4.0+60.0, lateral*1.2) > 0.5 {
			return PigmentSumi, true
		}
		return PigmentSumi, false
	case PigmentAi:
		if along < 0.11 {
			return PigmentAi, false
		}
		if math.Abs(lateral) > 0.64 {
			if hasAsagiSides && secondary != nil {
				return *secondary, true
			}
			return PigmentAi, false
		}
		return base, true
	default:
		return base, true
	}
}

func (kc *KoiCoat) developRamp(p Pigment, shadeRamp []RGBA, dither float64) []RGBA {
	if dither <= calculatePigmentDevelopment(p, kc.ColorFraction) {
		return shadeRamp
	}
	if len(shadeRamp) == 1 {
		return PaleNetRamp
	}
	switch p {
	case PigmentBeni, PigmentOrenji:
		return PaleRedRamp
	case PigmentSumi:
		return PaleBlackRamp
	case PigmentKi:
		return PaleYellowRamp
	case PigmentAi:
		return PaleBlueRamp
	default:
		return PaleBlackRamp
	}
}

func sameRamp(a, b []RGBA) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func calculatePigmentShares(dev *koi.KoiDevelopment) []PigmentShare {
	totalShare := 0.0
	for _, s := range dev.LanguageShares {
		totalShare += s
	}
	if totalShare == 0.0 {
		return nil
	}

	summed := make(map[Pigment]float64)
	for lang, share := range dev.LanguageShares {
		color := dev.LanguageColors[lang]
		p := selectPigment(color)
		summed[p] += share / totalShare
	}

	var list []PigmentShare
	for p, s := range summed {
		list = append(list, PigmentShare{Pigment: p, Share: s})
	}

	sort.Slice(list, func(i, j int) bool {
		rI := math.Round(list[i].Share * 1e9)
		rJ := math.Round(list[j].Share * 1e9)
		if rI != rJ {
			return rI > rJ
		}
		return list[i].Pigment < list[j].Pigment
	})

	return list
}

func selectPigment(color *history.LanguageColor) Pigment {
	const minChroma = 0.12
	const minSaturation = 0.2
	if color == nil {
		return PigmentSumi
	}
	r := float64(color[0]) / 255.0
	g := float64(color[1]) / 255.0
	b := float64(color[2]) / 255.0

	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	chroma := maxC - minC
	lightness := (maxC + minC) / 2.0
	denom := 1.0 - math.Abs(2.0*lightness-1.0)

	if chroma < minChroma || denom <= 1e-6 || chroma/denom < minSaturation {
		return PigmentSumi
	}

	var hue float64
	if maxC == r {
		hue = 60.0 * math.Mod(math.Mod((g-b)/chroma, 6.0)+6.0, 6.0)
	} else if maxC == g {
		hue = 60.0 * ((b-r)/chroma + 2.0)
	} else {
		hue = 60.0 * ((r-g)/chroma + 4.0)
	}

	if hue < 18.0 || hue >= 330.0 {
		return PigmentBeni
	} else if hue < 45.0 {
		return PigmentOrenji
	} else if hue < 160.0 {
		return PigmentKi
	}
	return PigmentAi
}

func calculatePigmentDevelopment(p Pigment, colorFrac float64) float64 {
	startFrac := 0.05
	spanFrac := 0.55
	if p == PigmentSumi {
		startFrac = 0.35
		spanFrac = 0.6
	}

	f := (colorFrac - startFrac) / spanFrac
	if f < 0.0 {
		f = 0.0
	} else if f > 1.0 {
		f = 1.0
	}
	return f * f * (3.0 - 2.0*f)
}

func sampleBody(calcHalfWidth func(float64) float64) []BodySample {
	const alongSteps = 170
	const alongStepFrac = 0.005
	const lateralSteps = 41
	const lateralStepFrac = 0.05

	samples := make([]BodySample, 0, alongSteps*lateralSteps)
	for a := 0; a < alongSteps; a++ {
		along := PatchSpanStart + float64(a)*alongStepFrac
		weight := calcHalfWidth(along)
		for l := 0; l < lateralSteps; l++ {
			samples = append(samples, BodySample{
				AlongFraction:   along,
				LateralFraction: -1.0 + float64(l)*lateralStepFrac,
				AreaWeight:      weight,
			})
		}
	}
	return samples
}

func fitPatchBiases(shares []PigmentShare, seed uint64, samples []BodySample) []float64 {
	const fitPasses = 120
	const fitRate = 0.6

	scores := make([][]float64, len(samples))
	totalWeight := 0.0
	for i, s := range samples {
		totalWeight += s.AreaWeight
		row := make([]float64, len(shares))
		for pIdx := range shares {
			row[pIdx] = samplePatchScore(seed, pIdx, s.AlongFraction, s.LateralFraction)
		}
		scores[i] = row
	}

	biases := make([]float64, len(shares))
	for pass := 0; pass < fitPasses; pass++ {
		covered := make([]float64, len(shares))
		for i, s := range samples {
			coverIdx := -1
			best := WhitePatchScore
			for pIdx := range shares {
				sc := scores[i][pIdx] + biases[pIdx]
				if sc > best {
					best = sc
					coverIdx = pIdx
				}
			}
			if coverIdx >= 0 {
				covered[coverIdx] += s.AreaWeight / totalWeight
			}
		}
		for pIdx, ps := range shares {
			biases[pIdx] += fitRate * (PatchCoverage*ps.Share - covered[pIdx])
		}
	}
	return biases
}

func findSpotThreshold(shares []PigmentShare, seed uint64, samples []BodySample) float64 {
	const spotsPerShare = 0.75
	const maxSpotCoverage = 0.32

	if len(shares) < 2 {
		return math.Inf(1)
	}
	coverage := shares[1].Share * spotsPerShare
	if coverage > maxSpotCoverage {
		coverage = maxSpotCoverage
	}

	var noises []float64
	for _, s := range samples {
		if s.AlongFraction >= SpotSpanStart && s.AlongFraction < SpotSpanEnd {
			noises = append(noises, sampleSpotNoise(seed, s.AlongFraction, s.LateralFraction))
		}
	}
	if len(noises) == 0 {
		return math.Inf(1)
	}
	sort.Float64s(noises)
	idx := int(math.Floor((1.0 - coverage) * float64(len(noises)-1)))
	return noises[idx]
}

func samplePatchScore(seed uint64, pIdx int, along, lateral float64) float64 {
	pSeed := seed ^ (uint64(pIdx+1) * 0x9e3779b97f4a7c15)
	return sampleFractalNoise(pSeed, along*3.0+float64(pIdx)*13.0, lateral*0.95+7.0)
}

func sampleSpotNoise(seed uint64, along, lateral float64) float64 {
	return sampleFractalNoise(seed^0x5b07, along*9.0+30.0, lateral*2.6)
}

func isInRedPlate(seed uint64, along, lateral float64) bool {
	plateShares := []float64{0.42, 0.33, 0.25}
	const platesStart = 0.035
	const platesEnd = 0.8
	const plateGap = 0.05

	platedFrac := platesEnd - platesStart - plateGap*float64(len(plateShares)-1)
	plateStart := platesStart

	for pIdx, share := range plateShares {
		plateLen := platedFrac * share
		plateCenter := plateStart + plateLen/2.0
		plateStart += plateLen + plateGap

		latIdx := int64(pIdx)
		lean := (hashLatticePoint(seed, latIdx, 1) - 0.5) * 0.55
		halfWidth := 0.62 + hashLatticePoint(seed, latIdx, 2)*0.24
		edgeWobble := (sampleFractalNoise(seed^0xa11ce, along*16.0+float64(pIdx)*31.0, lateral*2.4) - 0.5) * 0.9

		distAlong := math.Abs((along - plateCenter) / (plateLen / 2.0))
		distLateral := math.Abs((lateral - lean) / halfWidth)
		dist := math.Pow(distAlong, 4) + math.Pow(distLateral, 4)

		if dist < 1.0+edgeWobble {
			return true
		}
	}
	return false
}

func isOnScaleNet(along, lateral, bodyLen, halfWidth float64) bool {
	row := (lateral + 1.0) * halfWidth / 2.4
	col := along*bodyLen/2.8 + math.Mod(math.Floor(row), 2.0)*0.5

	rowFract := row - math.Floor(row)
	colFract := col - math.Floor(col)
	return math.Min(rowFract, colFract) < 0.17
}

func sampleFractalNoise(seed uint64, x, y float64) float64 {
	sum := 0.0
	totalWeight := 0.0
	for oct := 0; oct < 3; oct++ {
		freq := math.Pow(2.0, float64(oct))
		weight := 1.0 / freq
		sum += sampleSmoothNoise(seed, x*freq, y*freq) * weight
		totalWeight += weight
	}
	return sum / totalWeight
}

func sampleSmoothNoise(seed uint64, x, y float64) float64 {
	latX := int64(math.Floor(x))
	latY := int64(math.Floor(y))
	fracX := x - math.Floor(x)
	fracY := y - math.Floor(y)

	blendX := fracX * fracX * (3.0 - 2.0*fracX)
	blendY := fracY * fracY * (3.0 - 2.0*fracY)

	tl := hashLatticePoint(seed, latX, latY)
	tr := hashLatticePoint(seed, latX+1, latY)
	bl := hashLatticePoint(seed, latX, latY+1)
	br := hashLatticePoint(seed, latX+1, latY+1)

	top := tl + (tr-tl)*blendX
	bottom := bl + (br-bl)*blendX
	return top + (bottom-top)*blendY
}

func hashLatticePoint(seed uint64, x, y int64) float64 {
	hash := seed ^ (uint64(x) * 0x9e3779b97f4a7c15) ^ (uint64(y) * 0xbf58476d1ce4e5b9)
	hash = (hash ^ (hash >> 30)) * 0xbf58476d1ce4e5b9
	hash = (hash ^ (hash >> 27)) * 0x94d049bb133111eb
	hash ^= (hash >> 31)
	return float64(hash>>11) / float64(uint64(1)<<53)
}
