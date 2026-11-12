package render

import (
	"math"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/behavior"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/coat"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/motion"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/output"
)

const (
	FrameWidthPixels  = output.FrameWidthPixels
	FrameHeightPixels = output.FrameHeightPixels
	Pi                = math.Pi
	Tau               = 2.0 * math.Pi
)

var (
	OutlineColor = [4]byte{24, 20, 37, 255}
	DitherRanks  = [4][4]byte{
		{0, 8, 2, 10},
		{12, 4, 14, 6},
		{3, 11, 1, 9},
		{15, 7, 13, 5},
	}
)

type KoiAnatomy struct {
	BodyLengthPixels    float64
	BodyWidthPixels     float64
	HeadSlantFraction   float64
	NoseWidthFraction   float64
	TailTaperExponent   float64
	AdultFinFraction    float64
	FinLengthMultiplier float64
	EyeAlongFraction    float64
	EyeScale            float64
}

type CurvePoint struct {
	Position        [2]float64
	HalfWidthPixels float64
}

type CurveSample struct {
	AlongFraction   float64
	LateralPixels   float64
	DistancePixels  float64
	HalfWidthPixels float64
	NormalDirection [2]float64
}

type FinShape struct {
	LengthBodyWidths    float64
	HalfWidthBodyWidths float64
	WidthStartFraction  float64
	TaperExponent       float64
}

func (fs FinShape) Blend(other FinShape, otherFraction float64) FinShape {
	return FinShape{
		LengthBodyWidths:    interpolate(fs.LengthBodyWidths, other.LengthBodyWidths, otherFraction),
		HalfWidthBodyWidths: interpolate(fs.HalfWidthBodyWidths, other.HalfWidthBodyWidths, otherFraction),
		WidthStartFraction:  interpolate(fs.WidthStartFraction, other.WidthStartFraction, otherFraction),
		TaperExponent:       interpolate(fs.TaperExponent, other.TaperExponent, otherFraction),
	}
}

type FinKind int

const (
	FinTail FinKind = iota
	FinSide
	FinPelvic
)

type FinPair struct {
	Kind              FinKind
	RootAlongFraction float64
	RootSideFraction  float64
	YoungShape        FinShape
	AdultShape        FinShape
	DirectionOffsets  [2][]float64
}

var FinPairs = []FinPair{
	{
		Kind:              FinTail,
		RootAlongFraction: 1.0,
		RootSideFraction:  0.0,
		YoungShape: FinShape{
			LengthBodyWidths:    1.05,
			HalfWidthBodyWidths: 0.3,
			WidthStartFraction:  0.16,
			TaperExponent:       0.6,
		},
		AdultShape: FinShape{
			LengthBodyWidths:    1.62,
			HalfWidthBodyWidths: 0.3,
			WidthStartFraction:  0.125,
			TaperExponent:       0.69,
		},
		DirectionOffsets: [2][]float64{
			{-0.18035, -0.25648, -0.37367, -0.50421, -0.61344, -0.66997},
			{0.248, 0.30115, 0.29856, 0.24064, 0.14695, 0.05079},
		},
	},
	{
		Kind:              FinSide,
		RootAlongFraction: 0.22,
		RootSideFraction:  0.8,
		YoungShape: FinShape{
			LengthBodyWidths:    0.66,
			HalfWidthBodyWidths: 0.25,
			WidthStartFraction:  0.18,
			TaperExponent:       0.45,
		},
		AdultShape: FinShape{
			LengthBodyWidths:    1.13,
			HalfWidthBodyWidths: 0.24,
			WidthStartFraction:  0.09,
			TaperExponent:       0.54,
		},
		DirectionOffsets: [2][]float64{
			{0.87, 0.77239, 0.71583, 0.70044},
			{-0.65221, -0.58002, -0.5356, -0.48311},
		},
	},
	{
		Kind:              FinPelvic,
		RootAlongFraction: 0.47,
		RootSideFraction:  0.7,
		YoungShape: FinShape{
			LengthBodyWidths:    0.36,
			HalfWidthBodyWidths: 0.1,
			WidthStartFraction:  0.15,
			TaperExponent:       0.5,
		},
		AdultShape: FinShape{
			LengthBodyWidths:    0.57,
			HalfWidthBodyWidths: 0.11,
			WidthStartFraction:  0.11,
			TaperExponent:       0.59,
		},
		DirectionOffsets: [2][]float64{
			{0.39139, 0.3769, 0.39376},
			{-0.62997, -0.48164, -0.32809},
		},
	},
}

type framingMode int

const (
	framingCenterOnKoi framingMode = iota
	framingAt
)

type framing struct {
	mode   framingMode
	center [2]float64
}

// RenderKoi renders the koi in a centered, still pose.
func RenderKoi(dev *koi.KoiDevelopment, body coat.KoiBody) *output.KoiFrame {
	anatomy := calculateAnatomy(dev.TotalContributions)
	return renderPose(dev, body, anatomy, &motion.StillPose, framing{mode: framingCenterOnKoi})
}

// RenderSwimFrame renders one frame of a behavior's seamless loop.
func RenderSwimFrame(dev *koi.KoiDevelopment, body coat.KoiBody, b behavior.Behavior, frameIndex uint32) *output.KoiFrame {
	anatomy := calculateAnatomy(dev.TotalContributions)
	swimPose := motion.CalculateSwimPose(b, frameIndex)

	var poseCenter [2]float64
	switch swimPose.Course.Type {
	case motion.CourseLap:
		poseCenter = [2]float64{0.0, 0.0}
	case motion.CourseHover:
		hoverFit := fitHoverLoop(b, anatomy)
		poseCenter[0] = hoverFit.sweepCenter[0] - hoverFit.driftScales[0]*swimPose.Course.DriftPixels[0]
		poseCenter[1] = hoverFit.sweepCenter[1] - hoverFit.driftScales[1]*swimPose.Course.DriftPixels[1]
	}

	return renderPose(dev, body, anatomy, &swimPose, framing{mode: framingAt, center: poseCenter})
}

type hoverFit struct {
	sweepCenter [2]float64
	driftScales [2]float64
}

func fitHoverLoop(b behavior.Behavior, anatomy *KoiAnatomy) hoverFit {
	const fitMarginPixels = 2.5
	frameCount := motion.FrameCount(b)

	type boundEntry struct {
		minCorner   [2]float64
		maxCorner   [2]float64
		driftPixels [2]float64
	}

	entries := make([]boundEntry, frameCount)
	for i := uint32(0); i < frameCount; i++ {
		pose := motion.CalculateSwimPose(b, i)
		spine := buildSpine(anatomy, &pose)
		fins := buildFins(spine, anatomy, &pose)
		minC, maxC := measureCurveBounds(spine, fins)

		whiskers := calculateWhiskerPoints(spine, anatomy, pose.WhiskerSwayRadians)
		for _, wp := range whiskers {
			for ax := 0; ax < 2; ax++ {
				if wp[ax] < minC[ax] {
					minC[ax] = wp[ax]
				}
				if wp[ax]+1.0 > maxC[ax] {
					maxC[ax] = wp[ax] + 1.0
				}
			}
		}

		drift := [2]float64{0.0, 0.0}
		if pose.Course.Type == motion.CourseHover {
			drift = pose.Course.DriftPixels
		}

		entries[i] = boundEntry{
			minCorner:   minC,
			maxCorner:   maxC,
			driftPixels: drift,
		}
	}

	frameSizes := [2]float64{float64(FrameWidthPixels), float64(FrameHeightPixels)}
	res := hoverFit{driftScales: [2]float64{1.0, 1.0}}

	for ax := 0; ax < 2; ax++ {
		sweepMin := math.Inf(1)
		sweepMax := math.Inf(-1)
		for _, e := range entries {
			if e.minCorner[ax] < sweepMin {
				sweepMin = e.minCorner[ax]
			}
			if e.maxCorner[ax] > sweepMax {
				sweepMax = e.maxCorner[ax]
			}
		}

		sweepCenter := (sweepMin + sweepMax) / 2.0
		frameHalf := frameSizes[ax] / 2.0
		driftScale := 1.0

		for _, e := range entries {
			drift := e.driftPixels[ax]
			var room float64
			if drift > 0.0 {
				room = frameHalf - fitMarginPixels - (e.maxCorner[ax] - sweepCenter)
			} else {
				room = frameHalf - fitMarginPixels + (e.minCorner[ax] - sweepCenter)
			}
			if drift != 0.0 {
				candidate := room / math.Abs(drift)
				if candidate < driftScale {
					driftScale = candidate
				}
			}
		}

		res.sweepCenter[ax] = sweepCenter
		if driftScale < 0.0 {
			driftScale = 0.0
		} else if driftScale > 1.0 {
			driftScale = 1.0
		}
		res.driftScales[ax] = driftScale
	}

	return res
}

func renderPose(dev *koi.KoiDevelopment, body coat.KoiBody, anatomy *KoiAnatomy, pose *motion.SwimPose, fr framing) *output.KoiFrame {
	coatEngine := coat.NewKoiCoat(dev, body, func(along float64) float64 {
		return calculateBodyHalfWidth(along, anatomy)
	})

	spine := buildSpine(anatomy, pose)
	fins := buildFins(spine, anatomy, pose)
	minCorner, maxCorner := measureCurveBounds(spine, fins)

	var poseCenter [2]float64
	switch fr.mode {
	case framingCenterOnKoi:
		poseCenter[0] = (minCorner[0] + maxCorner[0]) / 2.0
		poseCenter[1] = (minCorner[1] + maxCorner[1]) / 2.0
	case framingAt:
		poseCenter = fr.center
	}

	frame := &output.KoiFrame{}

	for py := 0; py < FrameHeightPixels; py++ {
		for px := 0; px < FrameWidthPixels; px++ {
			samplePos := [2]float64{
				float64(px) + 0.5 - float64(FrameWidthPixels)/2.0 + poseCenter[0],
				float64(py) + 0.5 - float64(FrameHeightPixels)/2.0 + poseCenter[1],
			}

			inside := samplePos[0] >= minCorner[0] && samplePos[0] <= maxCorner[0] &&
				samplePos[1] >= minCorner[1] && samplePos[1] <= maxCorner[1]

			if inside {
				frame.Pixels[py*FrameWidthPixels+px] = sampleKoi(spine, fins, samplePos, [2]uint32{uint32(px), uint32(py)}, anatomy, coatEngine)
			}
		}
	}

	paintMouth(frame, spine, body, poseCenter, pose.MouthOpening)

	fishMask := make([]bool, FrameWidthPixels*FrameHeightPixels)
	for i, p := range frame.Pixels {
		fishMask[i] = p[3] != 0
	}

	whiskers := calculateWhiskerPoints(spine, anatomy, pose.WhiskerSwayRadians)
	paintWhiskers(frame, whiskers, poseCenter)
	paintOutline(frame)

	if pose.FeedingPhase != nil {
		paintFeeding(frame, spine, anatomy, poseCenter, *pose.FeedingPhase)
	}

	if pose.Depth > 0.0 {
		paintDepth(frame, fishMask, pose.Depth)
	}

	return frame
}

func calculateAnatomy(total uint64) *KoiAnatomy {
	const minBodyLengthPixels = 37.5
	const adultBodyLengthPixels = 137.5
	const maxBodyLengthPixels = 167.5
	const youngBodyWidthFraction = 0.17
	const adultBodyWidthFraction = 0.2
	const matureBodyWidthGain = 0.045
	const youngHeadSlantFraction = 0.14
	const adultHeadSlantFraction = 0.187
	const youngNoseWidthFraction = 0.75
	const adultNoseWidthFraction = 0.517
	const youngTailTaperExponent = 1.25
	const adultTailTaperExponent = 1.8
	const youngEyeAlongFraction = 0.08
	const adultEyeAlongFraction = 0.06
	const youngEyeScale = 2.1
	const youngFinLengthMultiplier = 0.85
	const maxFinLengthMultiplier = 1.1
	const youngFinHoldFraction = 0.25

	growthFrac := koi.CalculateGrowthFraction(total)
	adultGrowthFrac := koi.CalculateGrowthFraction(koi.AdultContributions)
	youngGrowthFrac := math.Min(growthFrac/adultGrowthFrac, 1.0)
	matureGrowthFrac := math.Max((growthFrac-adultGrowthFrac)/(1.0-adultGrowthFrac), 0.0)
	adultShapeFrac := math.Min(growthFrac/koi.CalculateGrowthFraction(koi.YoungShapeContributions), 1.0)

	bodyLength := minBodyLengthPixels +
		(adultBodyLengthPixels-minBodyLengthPixels)*youngGrowthFrac +
		(maxBodyLengthPixels-adultBodyLengthPixels)*matureGrowthFrac

	bodyWidthFrac := interpolate(youngBodyWidthFraction, adultBodyWidthFraction, adultShapeFrac) +
		matureBodyWidthGain*matureGrowthFrac

	finHold := (adultShapeFrac - youngFinHoldFraction) / (1.0 - youngFinHoldFraction)

	return &KoiAnatomy{
		BodyLengthPixels:    bodyLength,
		BodyWidthPixels:     bodyLength * bodyWidthFrac,
		HeadSlantFraction:   interpolate(youngHeadSlantFraction, adultHeadSlantFraction, adultShapeFrac),
		NoseWidthFraction:   interpolate(youngNoseWidthFraction, adultNoseWidthFraction, adultShapeFrac),
		TailTaperExponent:   interpolate(youngTailTaperExponent, adultTailTaperExponent, adultShapeFrac),
		AdultFinFraction:    smoothStep(finHold),
		FinLengthMultiplier: interpolate(youngFinLengthMultiplier, 1.0, smoothStep(adultShapeFrac)) * interpolate(1.0, maxFinLengthMultiplier, matureGrowthFrac),
		EyeAlongFraction:    interpolate(youngEyeAlongFraction, adultEyeAlongFraction, adultShapeFrac),
		EyeScale:            interpolate(youngEyeScale, 1.0, adultShapeFrac),
	}
}

func interpolate(from, to, fraction float64) float64 {
	return from + (to-from)*fraction
}

func smoothStep(fraction float64) float64 {
	f := fraction
	if f < 0.0 {
		f = 0.0
	} else if f > 1.0 {
		f = 1.0
	}
	return f * f * (3.0 - 2.0*f)
}

func buildSpine(anatomy *KoiAnatomy, pose *motion.SwimPose) []CurvePoint {
	const headingRadians = -0.08
	const bendRadians = 0.35
	const waveAmplitudeFraction = 0.06
	const waveCycles = 0.9
	const minHalfWidthPixels = 1.3

	bodyLen := anatomy.BodyLengthPixels
	pointCount := int(math.Round(bodyLen))
	if pointCount < 2 {
		pointCount = 2
	}

	calcPointHalfWidth := func(idx int) float64 {
		return math.Max(calculateBodyHalfWidth(float64(idx)/float64(pointCount), anatomy), minHalfWidthPixels)
	}

	var centerPoints []CurvePoint

	switch pose.Course.Type {
	case motion.CourseHover:
		centerPoints = make([]CurvePoint, pointCount)
		centerPoints[0] = CurvePoint{
			Position:        [2]float64{0.0, 0.0},
			HalfWidthPixels: calcPointHalfWidth(0),
		}
		for i := 1; i < pointCount; i++ {
			backwardAngle := headingRadians + pose.Course.HeadingOffsetRadians + Pi + bendRadians*float64(i)/float64(pointCount)
			prev := centerPoints[i-1].Position
			centerPoints[i] = CurvePoint{
				Position: [2]float64{
					prev[0] + math.Cos(backwardAngle),
					prev[1] + math.Sin(backwardAngle),
				},
				HalfWidthPixels: calcPointHalfWidth(i),
			}
		}
	case motion.CourseLap:
		centerPoints = make([]CurvePoint, pointCount)
		radius := calculateLapRadius(bodyLen)
		for i := 0; i < pointCount; i++ {
			angle := pose.Course.AngleRadians - float64(i)/radius
			centerPoints[i] = CurvePoint{
				Position: [2]float64{
					radius * math.Cos(angle),
					radius * math.Sin(angle),
				},
				HalfWidthPixels: calcPointHalfWidth(i),
			}
		}
	}

	bodySpine := make([]CurvePoint, pointCount)
	for i, cp := range centerPoints {
		along := float64(i) / float64(pointCount)
		fwd := calculateForwardDirection(centerPoints, i)
		const headWaveFraction = 0.1
		waveEnv := headWaveFraction + (1.0-headWaveFraction)*along*along
		waveOffset := bodyLen * waveAmplitudeFraction * pose.WaveStrength * waveEnv * math.Sin(along*Tau*waveCycles-pose.WavePhaseRadians)

		bodySpine[i] = CurvePoint{
			Position: [2]float64{
				cp.Position[0] - fwd[1]*waveOffset,
				cp.Position[1] + fwd[0]*waveOffset,
			},
			HalfWidthPixels: cp.HalfWidthPixels,
		}
	}

	return bodySpine
}

func buildFins(spine []CurvePoint, anatomy *KoiAnatomy, pose *motion.SwimPose) [][]CurvePoint {
	var fins [][]CurvePoint
	for _, fp := range FinPairs {
		rootIdx := int(math.Round(fp.RootAlongFraction * float64(len(spine)-1)))
		rootPoint := spine[rootIdx]
		fwd := calculateForwardDirection(spine, rootIdx)
		backwardAngle := math.Atan2(-fwd[1], -fwd[0])

		sides := [2]float64{-1.0, 1.0}
		for sideIdx, bodySide := range sides {
			offsets := fp.DirectionOffsets[sideIdx]
			outwardSign := 1.0
			if offsets[0] < 0 {
				outwardSign = -1.0
			}

			var rootTurn, bend float64
			switch fp.Kind {
			case FinTail:
				rootTurn = outwardSign * pose.TailSpreadRadians
				bend = pose.TailLobeBendsRadians[sideIdx]
			case FinSide:
				rootTurn = outwardSign * pose.SideFinOpeningsRadians[sideIdx]
				bend = outwardSign * pose.SideFinBendsRadians[sideIdx]
			case FinPelvic:
				rootTurn = outwardSign * pose.PelvicFinOpeningRadians
				bend = 0.0
			}

			rootOffset := rootPoint.HalfWidthPixels * fp.RootSideFraction * bodySide
			rootPos := [2]float64{
				rootPoint.Position[0] - fwd[1]*rootOffset,
				rootPoint.Position[1] + fwd[0]*rootOffset,
			}

			fin := buildFin(fp, rootPos, backwardAngle+rootTurn, anatomy, offsets, bend)
			fins = append(fins, fin)
		}
	}
	return fins
}

func buildFin(fp FinPair, rootPos [2]float64, backwardAngle float64, anatomy *KoiAnatomy, offsets []float64, bend float64) []CurvePoint {
	shape := fp.YoungShape.Blend(fp.AdultShape, anatomy.AdultFinFraction)
	finLen := anatomy.BodyWidthPixels * shape.LengthBodyWidths * anatomy.FinLengthMultiplier
	segCount := int(math.Ceil(finLen))
	if segCount < 1 {
		segCount = 1
	}
	segLen := finLen / float64(segCount)

	pos := rootPos
	points := make([]CurvePoint, segCount+1)

	for i := 0; i <= segCount; i++ {
		along := float64(i) / float64(segCount)
		widthProg := shape.WidthStartFraction + (1.0-shape.WidthStartFraction)*along

		halfWidth := anatomy.BodyWidthPixels * shape.HalfWidthBodyWidths * math.Pow(math.Max(math.Sin(Pi*widthProg), 0.0), shape.TaperExponent)
		points[i] = CurvePoint{
			Position:        pos,
			HalfWidthPixels: halfWidth,
		}

		dirPos := along*float64(len(offsets)) - 0.5
		if dirPos < 0.0 {
			dirPos = 0.0
		} else if dirPos > float64(len(offsets)-1) {
			dirPos = float64(len(offsets) - 1)
		}

		dirIdx := int(math.Floor(dirPos))
		nextDirIdx := dirIdx + 1
		if nextDirIdx >= len(offsets) {
			nextDirIdx = len(offsets) - 1
		}

		startAngle := offsets[dirIdx]
		endAngle := offsets[nextDirIdx]
		dirFract := dirPos - float64(dirIdx)

		segAngle := backwardAngle + startAngle + (endAngle-startAngle)*dirFract + bend*along
		pos[0] += segLen * math.Cos(segAngle)
		pos[1] += segLen * math.Sin(segAngle)
	}

	return points
}

func calculateMouthPosition(spine []CurvePoint) [2]float64 {
	const mouthForwardFraction = 0.75
	snout := spine[0]
	fwd := calculateForwardDirection(spine, 0)
	return [2]float64{
		snout.Position[0] + fwd[0]*snout.HalfWidthPixels*mouthForwardFraction,
		snout.Position[1] + fwd[1]*snout.HalfWidthPixels*mouthForwardFraction,
	}
}

func paintMouth(frame *output.KoiFrame, spine []CurvePoint, body coat.KoiBody, poseCenter [2]float64, mouthOpening float64) {
	const visibleOpening = 0.1
	const closedRadiusPixels = 0.8
	const openRadiusPerSnout = 0.5
	const minOpenRadiusPixels = 1.8
	const rimWidthPixels = 0.9

	if mouthOpening <= visibleOpening {
		return
	}

	openRadius := math.Max(spine[0].HalfWidthPixels*openRadiusPerSnout, minOpenRadiusPixels)
	radius := interpolate(closedRadiusPixels, openRadius, math.Min(mouthOpening, 1.0))
	mouthWorld := calculateMouthPosition(spine)

	mouth := [2]float64{
		mouthWorld[0] - poseCenter[0] + float64(FrameWidthPixels)/2.0,
		mouthWorld[1] - poseCenter[1] + float64(FrameHeightPixels)/2.0,
	}

	rimColor := body.SelectMouthRimColor()

	firstY := int(math.Max(math.Floor(mouth[1]-radius), 0.0))
	lastY := int(math.Min(math.Ceil(mouth[1]+radius), float64(FrameHeightPixels-1)))
	firstX := int(math.Max(math.Floor(mouth[0]-radius), 0.0))
	lastX := int(math.Min(math.Ceil(mouth[0]+radius), float64(FrameWidthPixels-1)))

	for py := firstY; py <= lastY; py++ {
		for px := firstX; px <= lastX; px++ {
			dist := math.Hypot(float64(px)+0.5-mouth[0], float64(py)+0.5-mouth[1])
			if dist <= radius {
				if dist > radius-rimWidthPixels {
					frame.Pixels[py*FrameWidthPixels+px] = rimColor
				} else {
					frame.Pixels[py*FrameWidthPixels+px] = OutlineColor
				}
			}
		}
	}
}

func calculateWhiskerPoints(spine []CurvePoint, anatomy *KoiAnatomy, sway float64) [][2]float64 {
	const spreadRadians = 0.85
	const rootSideFraction = 0.8
	const rootForwardFraction = 0.55

	whiskerLen := calculateWhiskerLength(anatomy)
	snoutPos := spine[0].Position
	snoutRadius := spine[0].HalfWidthPixels
	fwd := calculateForwardDirection(spine, 0)
	heading := math.Atan2(fwd[1], fwd[0])

	var points [][2]float64
	sides := []float64{-1.0, 1.0}

	for _, bodySide := range sides {
		sideOffset := snoutRadius * rootSideFraction * bodySide
		fwdOffset := snoutRadius * rootForwardFraction
		root := [2]float64{
			snoutPos[0] - fwd[1]*sideOffset + fwd[0]*fwdOffset,
			snoutPos[1] + fwd[0]*sideOffset + fwd[1]*fwdOffset,
		}
		angle := heading + sway + bodySide*spreadRadians
		for step := 1; step <= int(whiskerLen); step++ {
			points = append(points, [2]float64{
				root[0] + math.Cos(angle)*float64(step),
				root[1] + math.Sin(angle)*float64(step),
			})
		}
	}
	return points
}

func calculateWhiskerLength(anatomy *KoiAnatomy) uint32 {
	const bodyPixelsPerWhiskerPixel = 34.0
	val := math.Round(anatomy.BodyLengthPixels / bodyPixelsPerWhiskerPixel)
	if val < 3.0 {
		val = 3.0
	} else if val > 5.0 {
		val = 5.0
	}
	return uint32(val)
}

func paintWhiskers(frame *output.KoiFrame, points [][2]float64, poseCenter [2]float64) {
	for _, wp := range points {
		px := math.Floor(wp[0] - poseCenter[0] + float64(FrameWidthPixels)/2.0)
		py := math.Floor(wp[1] - poseCenter[1] + float64(FrameHeightPixels)/2.0)

		if px >= 0 && px < float64(FrameWidthPixels) && py >= 0 && py < float64(FrameHeightPixels) {
			idx := int(py)*FrameWidthPixels + int(px)
			if frame.Pixels[idx][3] == 0 {
				frame.Pixels[idx] = [4]byte(coat.WhiteRamp[1])
			}
		}
	}
}

func paintOutline(frame *output.KoiFrame) {
	orig := frame.Pixels
	for idx, p := range orig {
		if p[3] != 0 {
			continue
		}
		x := idx % FrameWidthPixels
		y := idx / FrameWidthPixels

		touches := (x > 0 && orig[idx-1][3] != 0) ||
			(x+1 < FrameWidthPixels && orig[idx+1][3] != 0) ||
			(y > 0 && orig[idx-FrameWidthPixels][3] != 0) ||
			(y+1 < FrameHeightPixels && orig[idx+FrameWidthPixels][3] != 0)

		if touches {
			frame.Pixels[idx] = OutlineColor
		}
	}
}

func paintFeeding(frame *output.KoiFrame, spine []CurvePoint, anatomy *KoiAnatomy, poseCenter [2]float64, feedingPhase float64) {
	const ringCount = 3
	const ringWidthPerBodyLen = 0.3
	const ringHeightPerWidth = 0.4
	const ringAngleStep = 0.035
	const pelletCount = 5
	const pelletNearestPixels = 4.0
	const pelletReachPixels = 14.0
	const pelletWanderPixels = 9.0
	const pelletEatenLife = 0.92
	const edgeMarginPixels = 2.0

	pelletColors := [2][4]byte{
		{228, 166, 114, 255},
		{184, 111, 80, 255},
	}

	mouthWorld := calculateMouthPosition(spine)
	mouth := [2]float64{
		mouthWorld[0] - poseCenter[0] + float64(FrameWidthPixels)/2.0,
		mouthWorld[1] - poseCenter[1] + float64(FrameHeightPixels)/2.0,
	}

	room := [2]float64{
		math.Max(math.Min(mouth[0]-edgeMarginPixels, float64(FrameWidthPixels)-1.0-edgeMarginPixels-mouth[0]), 0.0),
		math.Max(math.Min(mouth[1]-edgeMarginPixels, float64(FrameHeightPixels)-1.0-edgeMarginPixels-mouth[1]), 0.0),
	}

	fwd := calculateForwardDirection(spine, 0)

	paintWater := func(point [2]float64, c [4]byte, ditherLimit float64) {
		px := math.Floor(point[0])
		py := math.Floor(point[1])
		if px < 0 || px >= float64(FrameWidthPixels) || py < 0 || py >= float64(FrameHeightPixels) {
			return
		}
		idx := int(py)*FrameWidthPixels + int(px)
		if frame.Pixels[idx][3] == 0 && calculateDitherThreshold([2]uint32{uint32(px), uint32(py)}) < ditherLimit {
			frame.Pixels[idx] = c
		}
	}

	largestRingWidth := anatomy.BodyLengthPixels * ringWidthPerBodyLen * 0.85
	ringScale := math.Min(1.0, math.Min(room[0]/largestRingWidth, room[1]/(largestRingWidth*ringHeightPerWidth)))

	for ringIdx := 0; ringIdx < ringCount; ringIdx++ {
		ringPhase := math.Mod(feedingPhase+float64(ringIdx)/float64(ringCount), 1.0)
		ringWidth := ringScale * anatomy.BodyLengthPixels * ringWidthPerBodyLen * (0.25 + 0.6*ringPhase)
		colorIdx := 1
		if ringIdx%2 == 1 {
			colorIdx = 2
		}
		ringColor := [4]byte(coat.WhiteRamp[colorIdx])

		for ringAngle := 0.0; ringAngle < Tau; ringAngle += ringAngleStep {
			paintWater([2]float64{
				mouth[0] + math.Cos(ringAngle)*ringWidth,
				mouth[1] + math.Sin(ringAngle)*ringWidth*ringHeightPerWidth,
			}, ringColor, (1.0-ringPhase)*0.75)
		}
	}

	farthestPellet := pelletNearestPixels + pelletReachPixels + pelletWanderPixels
	pelletScale := math.Min(1.0, math.Max(math.Min(room[0], room[1])-1.0, 0.0)/farthestPellet)

	for pIdx := 0; pIdx < pelletCount; pIdx++ {
		life := math.Mod(feedingPhase+float64(pIdx)/float64(pelletCount), 1.0)
		if life >= pelletEatenLife {
			continue
		}
		reach := pelletNearestPixels + (1.0-life)*pelletReachPixels
		wander := (1.0 - life) * pelletWanderPixels
		pelletPt := [2]float64{
			mouth[0] + pelletScale*(fwd[0]*reach+math.Sin(float64(pIdx)*2.1)*wander),
			mouth[1] + pelletScale*(fwd[1]*reach+math.Cos(float64(pIdx)*1.7)*wander),
		}
		paintWater(pelletPt, pelletColors[0], 1.0)
		paintWater([2]float64{pelletPt[0] + 1.0, pelletPt[1]}, pelletColors[1], 1.0)
	}
}

func paintDepth(frame *output.KoiFrame, fishMask []bool, depth float64) {
	deepWater := [3]float64{18.0, 78.0, 137.0}
	midWater := [3]float64{0.0, 153.0, 219.0}
	shallowWater := [3]float64{190.0, 225.0, 245.0}
	outlineWater := [3]float64{24.0, 48.0, 92.0}

	const fishDropoutPerDepth = 0.15
	const outlineDropoutPerDepth = 0.45
	const tintPerDepth = 1.15

	tint := math.Min(depth*tintPerDepth, 1.0)

	for idx, p := range frame.Pixels {
		if p[3] == 0 {
			continue
		}
		isFish := fishMask[idx]
		dropoutRate := outlineDropoutPerDepth
		if isFish {
			dropoutRate = fishDropoutPerDepth
		}

		px := uint32(idx % FrameWidthPixels)
		py := uint32(idx / FrameWidthPixels)

		if calculateDitherThreshold([2]uint32{px, py}) < depth*dropoutRate {
			frame.Pixels[idx] = [4]byte{0, 0, 0, 0}
			continue
		}

		orig := [3]float64{float64(p[0]), float64(p[1]), float64(p[2])}
		var water [3]float64
		if isFish {
			brightness := (0.3*orig[0] + 0.59*orig[1] + 0.11*orig[2]) / 255.0
			var fromC, toC [3]float64
			var blend float64
			if brightness < 0.5 {
				fromC = deepWater
				toC = midWater
				blend = brightness * 2.0
			} else {
				fromC = midWater
				toC = shallowWater
				blend = brightness*2.0 - 1.0
			}
			for c := 0; c < 3; c++ {
				water[c] = interpolate(fromC[c], toC[c], blend)
			}
		} else {
			water = outlineWater
		}

		for c := 0; c < 3; c++ {
			frame.Pixels[idx][c] = uint8(math.Round(interpolate(orig[c], water[c], tint)))
		}
	}
}

func calculateLapRadius(bodyLen float64) float64 {
	const frameReachPixels = 116.0
	const outerReachFraction = 0.33
	const maxLapRadiusPixels = 88.0
	const maxBodyShare = 0.6

	return math.Max(
		math.Min(frameReachPixels-outerReachFraction*bodyLen, maxLapRadiusPixels),
		bodyLen/(Tau*maxBodyShare),
	)
}

func measureCurveBounds(spine []CurvePoint, fins [][]CurvePoint) ([2]float64, [2]float64) {
	minC := [2]float64{math.Inf(1), math.Inf(1)}
	maxC := [2]float64{math.Inf(-1), math.Inf(-1)}

	checkPoint := func(cp CurvePoint) {
		for ax := 0; ax < 2; ax++ {
			if cp.Position[ax]-cp.HalfWidthPixels < minC[ax] {
				minC[ax] = cp.Position[ax] - cp.HalfWidthPixels
			}
			if cp.Position[ax]+cp.HalfWidthPixels > maxC[ax] {
				maxC[ax] = cp.Position[ax] + cp.HalfWidthPixels
			}
		}
	}

	for _, pt := range spine {
		checkPoint(pt)
	}
	for _, fin := range fins {
		for _, pt := range fin {
			checkPoint(pt)
		}
	}

	return minC, maxC
}

func sampleKoi(spine []CurvePoint, fins [][]CurvePoint, samplePos [2]float64, pixelPos [2]uint32, anatomy *KoiAnatomy, coatEngine *coat.KoiCoat) [4]byte {
	if bodySample, ok := sampleCurve(spine, samplePos); ok {
		return shadeBody(bodySample, pixelPos, anatomy, coatEngine)
	}
	for _, fin := range fins {
		if finSample, ok := sampleCurve(fin, samplePos); ok {
			return shadeFin(finSample, pixelPos, coatEngine)
		}
	}
	return [4]byte{0, 0, 0, 0}
}

func sampleCurve(points []CurvePoint, samplePos [2]float64) (CurveSample, bool) {
	var closestSample CurveSample
	closestDist := math.Inf(1)
	found := false

	for i := 0; i < len(points)-1; i++ {
		start := points[i]
		end := points[i+1]

		segX := end.Position[0] - start.Position[0]
		segY := end.Position[1] - start.Position[1]
		offX := samplePos[0] - start.Position[0]
		offY := samplePos[1] - start.Position[1]

		segLen := math.Hypot(segX, segY)
		segProg := (offX*segX + offY*segY) / (segLen * segLen)
		if segProg < 0.0 {
			segProg = 0.0
		} else if segProg > 1.0 {
			segProg = 1.0
		}

		dist := math.Hypot(offX-segProg*segX, offY-segProg*segY)
		if dist >= closestDist {
			continue
		}

		closestDist = dist
		found = true
		closestSample = CurveSample{
			AlongFraction:   (float64(i) + segProg) / float64(len(points)-1),
			LateralPixels:   (offX*segY - offY*segX) / segLen,
			DistancePixels:  dist,
			HalfWidthPixels: start.HalfWidthPixels + segProg*(end.HalfWidthPixels-start.HalfWidthPixels),
			NormalDirection: [2]float64{segY / segLen, -segX / segLen},
		}
	}

	if found && closestSample.DistancePixels <= closestSample.HalfWidthPixels {
		return closestSample, true
	}
	return CurveSample{}, false
}

func shadeBody(sample CurveSample, pixelPos [2]uint32, anatomy *KoiAnatomy, coatEngine *coat.KoiCoat) [4]byte {
	bodyLen := anatomy.BodyLengthPixels
	lateralFrac := sample.LateralPixels / sample.HalfWidthPixels
	dither := calculateDitherThreshold(pixelPos)

	paint := coatEngine.SelectBodyPaint(sample.AlongFraction, lateralFrac, bodyLen, sample.HalfWidthPixels, dither)

	const innerRimPixels = 0.8
	if sample.DistancePixels > sample.HalfWidthPixels-innerRimPixels {
		return [4]byte(paint.ShadeRamp[0])
	}

	const eyeLateralFrac = 0.55
	eyeHalfThickness := 0.8 * anatomy.EyeScale
	eyeHalfSpan := 0.25 * math.Sqrt(anatomy.EyeScale)

	if math.Abs((sample.AlongFraction-anatomy.EyeAlongFraction)*bodyLen) < eyeHalfThickness &&
		math.Abs(math.Abs(lateralFrac)-eyeLateralFrac) < eyeHalfSpan {
		return OutlineColor
	}

	if paint.GlintColor != nil {
		return *paint.GlintColor
	}

	surfaceBrightness := calculateBodyBrightness(sample, lateralFrac, anatomy) + paint.BrightnessShift
	shade := chooseShade(paint.ShadeRamp, surfaceBrightness, pixelPos)
	return adjustSaturation(shade, paint.Saturation)
}

func calculateBodyBrightness(sample CurveSample, lateralFrac float64, anatomy *KoiAnatomy) float64 {
	const ambientBrightness = 0.24
	const diffuseStrength = 0.7
	lightDir := [3]float64{-0.32, -0.5, 0.8}

	normalZ := math.Sqrt(math.Max(1.0-lateralFrac*lateralFrac, 0.0))
	incident := lateralFrac*(sample.NormalDirection[0]*lightDir[0]+sample.NormalDirection[1]*lightDir[1]) + normalZ*lightDir[2]

	surfaceBrightness := ambientBrightness + diffuseStrength*math.Max(incident, 0.0)

	const dorsalSpanStart = 0.27
	const dorsalSpanEnd = 0.62
	const dorsalHalfWidthFrac = 0.1
	const dorsalShadow = 0.18

	if sample.AlongFraction >= dorsalSpanStart && sample.AlongFraction < dorsalSpanEnd &&
		math.Abs(lateralFrac) < dorsalHalfWidthFrac {
		surfaceBrightness -= dorsalShadow
	}

	gillAlong := anatomy.HeadSlantFraction * 0.72
	const gillInnerFrac = 0.3
	const gillShadow = 0.22

	if math.Abs((sample.AlongFraction-gillAlong)*anatomy.BodyLengthPixels) < 0.5 &&
		math.Abs(lateralFrac) > gillInnerFrac {
		surfaceBrightness -= gillShadow
	}

	return surfaceBrightness
}

func shadeFin(sample CurveSample, pixelPos [2]uint32, coatEngine *coat.KoiCoat) [4]byte {
	const finRimPixels = 0.6
	const finRaySpacingPixels = 1.8
	const finRootShadow = 0.12

	isOnRim := sample.DistancePixels > sample.HalfWidthPixels-finRimPixels
	dither := calculateDitherThreshold(pixelPos)
	paint := coatEngine.SelectFinPaint(sample.AlongFraction, sample.LateralPixels/sample.HalfWidthPixels, isOnRim, dither)

	if isOnRim {
		return adjustSaturation([4]byte(paint.ShadeRamp[1]), paint.Saturation)
	}

	rayPhase := math.Mod(math.Mod(sample.LateralPixels/finRaySpacingPixels, 1.0)+1.0, 1.0)
	surfaceBrightness := 0.94
	if rayPhase < 0.5 {
		surfaceBrightness = 0.68
	}

	rootShadow := finRootShadow * math.Pow(1.0-sample.AlongFraction, 2)
	shade := chooseShade(paint.ShadeRamp, surfaceBrightness-rootShadow, pixelPos)
	return adjustSaturation(shade, paint.Saturation)
}

func calculateBodyHalfWidth(along float64, anatomy *KoiAnatomy) float64 {
	const minShoulderFraction = 0.3
	const shoulderGapFraction = 0.04
	const taperEndFraction = 0.92
	const tailWidthFraction = 0.26

	var widthFrac float64
	if along < anatomy.HeadSlantFraction {
		noseW := anatomy.NoseWidthFraction
		widthFrac = noseW + (1.0-noseW)*math.Sin(along/anatomy.HeadSlantFraction*Pi/2.0)
	} else {
		shoulderFrac := math.Max(anatomy.HeadSlantFraction+shoulderGapFraction, minShoulderFraction)
		taperLen := taperEndFraction - shoulderFrac
		taperProg := math.Max(along-shoulderFrac, 0.0) / taperLen
		widthFrac = math.Max(1.0-math.Pow(taperProg, anatomy.TailTaperExponent), tailWidthFraction)
	}

	return anatomy.BodyWidthPixels / 2.0 * widthFrac
}

func calculateForwardDirection(spine []CurvePoint, pointIndex int) [2]float64 {
	const span = 2
	earlierIdx := pointIndex - span
	if earlierIdx < 0 {
		earlierIdx = 0
	}
	laterIdx := pointIndex + span
	if laterIdx >= len(spine) {
		laterIdx = len(spine) - 1
	}

	tx := spine[earlierIdx].Position[0] - spine[laterIdx].Position[0]
	ty := spine[earlierIdx].Position[1] - spine[laterIdx].Position[1]
	lenPx := math.Hypot(tx, ty)
	if lenPx == 0.0 {
		return [2]float64{1.0, 0.0}
	}
	return [2]float64{tx / lenPx, ty / lenPx}
}

func chooseShade(ramp []coat.RGBA, brightness float64, pixelPos [2]uint32) [4]byte {
	b := brightness
	if b < 0.0 {
		b = 0.0
	} else if b > 1.0 {
		b = 1.0
	}
	pos := b * float64(len(ramp)-1)
	baseIdx := int(math.Floor(pos))
	fract := pos - float64(baseIdx)

	idx := baseIdx
	if fract > calculateDitherThreshold(pixelPos) {
		idx++
	}
	if idx >= len(ramp) {
		idx = len(ramp) - 1
	}
	return [4]byte(ramp[idx])
}

func adjustSaturation(c [4]byte, saturation float64) [4]byte {
	r := float64(c[0])
	g := float64(c[1])
	b := float64(c[2])
	gray := 0.3*r + 0.59*g + 0.11*b

	sat := func(ch float64) uint8 {
		val := math.Round(gray + (ch-gray)*saturation)
		if val < 0.0 {
			val = 0.0
		} else if val > 255.0 {
			val = 255.0
		}
		return uint8(val)
	}

	return [4]byte{sat(r), sat(g), sat(b), c[3]}
}

func calculateDitherThreshold(pixelPos [2]uint32) float64 {
	dx := int(pixelPos[0] % 4)
	dy := int(pixelPos[1] % 4)
	return (float64(DitherRanks[dy][dx]) + 0.5) / 16.0
}
