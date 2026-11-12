package motion

import (
	"math"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/behavior"
)

const (
	FramesPerSecond = 15

	TailBendPerWaveStrength = 0.45
	TailBendLagRadians      = 1.7
	TailSplitRadians        = 0.1
	TailSplitLagRadians     = 2.6
	SideFinTrailFrames      = 3
	PelvicShareOfSideFin    = 0.5
	WhiskerSwayRadians      = 0.15
	WhiskerSwayLagRadians   = 0.6

	RestFrameCount       uint32  = 90
	RestDepth            float64 = 0.22
	FeedingCyclesPerLoop float64 = 3.0
	BreathsPerLoop       float64 = 5.0
	BreathOpening        float64 = 0.7
	GulpsPerLoop         float64 = 3.0
	DeepRestFrameCount   uint32  = 120
	DeepRestDepth        float64 = 0.5
)

const (
	Pi  = math.Pi
	Tau = 2.0 * math.Pi
)

type CourseType int

const (
	CourseLap CourseType = iota
	CourseHover
)

type Course struct {
	Type                 CourseType
	AngleRadians         float64
	HeadingOffsetRadians float64
	DriftPixels          [2]float64
}

type SwimPose struct {
	Course                  Course
	Depth                   float64
	WavePhaseRadians        float64
	WaveStrength            float64
	TailLobeBendsRadians    [2]float64
	TailSpreadRadians       float64
	SideFinOpeningsRadians  [2]float64
	SideFinBendsRadians     [2]float64
	PelvicFinOpeningRadians float64
	WhiskerSwayRadians      float64
	FeedingPhase            *float64
	MouthOpening            float64
}

var StillPose = SwimPose{
	Course: Course{
		Type:                 CourseHover,
		HeadingOffsetRadians: 0.0,
		DriftPixels:          [2]float64{0.0, 0.0},
	},
	Depth:                   0.0,
	WavePhaseRadians:        0.8,
	WaveStrength:            1.0,
	TailLobeBendsRadians:    [2]float64{0.0, 0.0},
	TailSpreadRadians:       0.0,
	SideFinOpeningsRadians:  [2]float64{0.0, 0.0},
	SideFinBendsRadians:     [2]float64{0.0, 0.0},
	PelvicFinOpeningRadians: 0.0,
	WhiskerSwayRadians:      0.0,
	FeedingPhase:            nil,
	MouthOpening:            0.0,
}

func FrameCount(b behavior.Behavior) uint32 {
	switch b {
	case behavior.Cruise:
		return cruiseStyle.LoopFrameCount
	case behavior.Dart:
		return dartStyle.LoopFrameCount
	case behavior.Rest, behavior.SurfaceFeeding:
		return RestFrameCount
	case behavior.DeepRest:
		return DeepRestFrameCount
	default:
		return RestFrameCount
	}
}

type sideFinStrokeType int

const (
	sideFinBrake sideFinStrokeType = iota
	sideFinTuckAndSpread
)

type sideFinStrokes struct {
	strokeType           sideFinStrokeType
	strokeOpeningRadians float64
	brakeOpeningRadians  float64
	tuckRadians          float64
	spreadRadians        float64
}

type lapStyle struct {
	LoopFrameCount          uint32
	LapCount                float64
	StrokeFrameCount        float64
	StrokeShare             float64
	SpeedLagFrames          float64
	GlideSpeed              float64
	StrokeSpeedGain         float64
	WaveFrameCount          float64
	GlideWaveStrength       float64
	StrokeWaveGain          float64
	GlideTailSpreadRadians  float64
	StrokeTailSpreadRadians float64
	SideFinStrokes          sideFinStrokes
	InsideFlareRadians      float64
	ScullSwingRadians       float64
	ScullFrameCount         float64
	ScullPhaseRadians       float64
}

var cruiseStyle = lapStyle{
	LoopFrameCount:          120,
	LapCount:                1.0,
	StrokeFrameCount:        60.0,
	StrokeShare:             0.55,
	SpeedLagFrames:          6.0,
	GlideSpeed:              0.75,
	StrokeSpeedGain:         0.5,
	WaveFrameCount:          8.0,
	GlideWaveStrength:       0.3,
	StrokeWaveGain:          0.9,
	GlideTailSpreadRadians:  0.05,
	StrokeTailSpreadRadians: 0.12,
	SideFinStrokes: sideFinStrokes{
		strokeType:           sideFinBrake,
		strokeOpeningRadians: 0.08,
		brakeOpeningRadians:  0.3,
	},
	InsideFlareRadians: 0.12,
	ScullSwingRadians:  0.1,
	ScullFrameCount:    20.0,
	ScullPhaseRadians:  1.1,
}

var dartStyle = lapStyle{
	LoopFrameCount:          120,
	LapCount:                2.0,
	StrokeFrameCount:        30.0,
	StrokeShare:             0.35,
	SpeedLagFrames:          3.0,
	GlideSpeed:              0.35,
	StrokeSpeedGain:         2.2,
	WaveFrameCount:          6.0,
	GlideWaveStrength:       0.25,
	StrokeWaveGain:          1.4,
	GlideTailSpreadRadians:  0.04,
	StrokeTailSpreadRadians: 0.16,
	SideFinStrokes: sideFinStrokes{
		strokeType:    sideFinTuckAndSpread,
		tuckRadians:   0.12,
		spreadRadians: 0.4,
	},
	InsideFlareRadians: 0.1,
	ScullSwingRadians:  0.08,
	ScullFrameCount:    15.0,
	ScullPhaseRadians:  0.0,
}

type drive struct {
	course                Course
	depth                 float64
	wavePhaseRadians      float64
	waveStrength          float64
	tailSpreadRadians     float64
	strokeOpeningRadians  float64
	rightFinOffsetRadians float64
	feedingPhase          *float64
	mouthOpening          float64
}

func (d *drive) calculateSideFinOpenings() [2]float64 {
	return [2]float64{
		d.strokeOpeningRadians - d.rightFinOffsetRadians,
		d.strokeOpeningRadians + d.rightFinOffsetRadians,
	}
}

func CalculateSwimPose(b behavior.Behavior, frameIndex uint32) SwimPose {
	loopCount := FrameCount(b)
	loopFrame := frameIndex % loopCount
	d := calculateDrive(b, loopFrame)

	trailingFrame := (loopFrame + loopCount - SideFinTrailFrames) % loopCount
	trailingDrive := calculateDrive(b, trailingFrame)

	tailBend := TailBendPerWaveStrength * d.waveStrength * math.Sin(d.wavePhaseRadians-TailBendLagRadians)
	tailSplit := TailSplitRadians * math.Sin(d.wavePhaseRadians-TailSplitLagRadians)

	sideFinOpenings := d.calculateSideFinOpenings()
	trailingOpenings := trailingDrive.calculateSideFinOpenings()

	return SwimPose{
		Course:                  d.course,
		Depth:                   d.depth,
		WavePhaseRadians:        d.wavePhaseRadians,
		WaveStrength:            d.waveStrength,
		TailLobeBendsRadians:    [2]float64{tailBend - tailSplit, tailBend + tailSplit},
		TailSpreadRadians:       d.tailSpreadRadians,
		SideFinOpeningsRadians:  sideFinOpenings,
		SideFinBendsRadians:     [2]float64{trailingOpenings[0] - sideFinOpenings[0], trailingOpenings[1] - sideFinOpenings[1]},
		PelvicFinOpeningRadians: PelvicShareOfSideFin * d.strokeOpeningRadians,
		WhiskerSwayRadians:      WhiskerSwayRadians * math.Sin(d.wavePhaseRadians-WhiskerSwayLagRadians),
		FeedingPhase:            d.feedingPhase,
		MouthOpening:            d.mouthOpening,
	}
}

func calculateDrive(b behavior.Behavior, loopFrame uint32) drive {
	switch b {
	case behavior.Cruise:
		return cruiseStyle.calculateDrive(loopFrame)
	case behavior.Dart:
		return dartStyle.calculateDrive(loopFrame)
	case behavior.Rest:
		return calculateRestDrive(loopFrame, RestDepth)
	case behavior.SurfaceFeeding:
		d := calculateRestDrive(loopFrame, 0.0)
		phase := math.Mod(FeedingCyclesPerLoop*float64(loopFrame)/float64(RestFrameCount), 1.0)
		d.feedingPhase = &phase
		return d
	case behavior.DeepRest:
		return calculateDeepRestDrive(loopFrame)
	default:
		return calculateRestDrive(loopFrame, RestDepth)
	}
}

func calculateBeat(frame, cycleFrameCount, beatShare float64) float64 {
	cycleFraction := math.Mod(math.Mod(frame, cycleFrameCount)+cycleFrameCount, cycleFrameCount) / cycleFrameCount
	if cycleFraction < beatShare {
		return math.Sin(Pi * cycleFraction / beatShare)
	}
	return 0.0
}

func calculateRestDrive(loopFrame uint32, depth float64) drive {
	const strokeFrameCount = 45.0
	const strokeShare = 0.4
	const swingRadians = 0.1
	driftPixels := [2]float64{3.0, 2.0}
	const waveFrameCount = 15.0
	const glideWaveStrength = 0.15
	const strokeWaveGain = 0.5
	const tailSpreadRadians = 0.04
	const sideFinOpeningRadians = 0.25
	const sideFinBreathRadians = 0.08
	const scullSwingRadians = 0.12
	const scullFrameCount = 18.0

	frame := float64(loopFrame)
	loopTurn := Tau * frame / float64(RestFrameCount)
	stroke := calculateBeat(frame, strokeFrameCount, strokeShare)

	mouth := math.Sin(GulpsPerLoop * loopTurn)
	if mouth < 0.0 {
		mouth = 0.0
	}

	return drive{
		course: Course{
			Type:                 CourseHover,
			HeadingOffsetRadians: swingRadians * math.Sin(loopTurn),
			DriftPixels: [2]float64{
				-driftPixels[0] * math.Sin(loopTurn),
				-driftPixels[1] * math.Sin(2.0*loopTurn+1.0),
			},
		},
		depth:                 depth,
		wavePhaseRadians:      StillPose.WavePhaseRadians + frame*Tau/waveFrameCount,
		waveStrength:          glideWaveStrength + strokeWaveGain*stroke,
		tailSpreadRadians:     tailSpreadRadians,
		strokeOpeningRadians:  sideFinOpeningRadians + sideFinBreathRadians*math.Sin(3.0*loopTurn),
		rightFinOffsetRadians: scullSwingRadians * math.Sin(Tau*frame/scullFrameCount),
		feedingPhase:          nil,
		mouthOpening:          mouth,
	}
}

func calculateDeepRestDrive(loopFrame uint32) drive {
	const swingRadians = 0.05
	driftPixels := [2]float64{1.5, 1.0}
	const sinkPixels = 4.0
	const waveFrameCount = 30.0
	const waveStrength = 0.12
	const waveBreath = 0.1
	const tailSpreadRadians = 0.02
	const sideFinOpeningRadians = 0.2
	const flickRadians = 0.35
	const flickStartFrame = 70.0
	const flickFrameCount = 8.0
	const gulpStartFrame = 66.0
	const gulpFrameCount = 16.0

	frame := float64(loopFrame)
	loopTurn := Tau * frame / float64(DeepRestFrameCount)

	flickFraction := (frame - flickStartFrame) / flickFrameCount
	flick := 0.0
	if flickFraction >= 0.0 && flickFraction <= 1.0 {
		flick = math.Sin(Pi * flickFraction)
	}

	gulpFraction := (frame - gulpStartFrame) / gulpFrameCount
	mouth := 0.0
	if gulpFraction >= 0.0 && gulpFraction <= 1.0 {
		mouth = math.Sin(Pi * gulpFraction)
	}

	return drive{
		course: Course{
			Type:                 CourseHover,
			HeadingOffsetRadians: swingRadians * math.Sin(loopTurn),
			DriftPixels: [2]float64{
				-driftPixels[0] * math.Sin(loopTurn),
				sinkPixels - driftPixels[1]*math.Sin(2.0*loopTurn),
			},
		},
		depth:                 DeepRestDepth,
		wavePhaseRadians:      StillPose.WavePhaseRadians + frame*Tau/waveFrameCount,
		waveStrength:          waveStrength + waveBreath*math.Sin(2.0*loopTurn),
		tailSpreadRadians:     tailSpreadRadians,
		strokeOpeningRadians:  sideFinOpeningRadians,
		rightFinOffsetRadians: flickRadians * flick,
		feedingPhase:          nil,
		mouthOpening:          mouth,
	}
}

func (s *lapStyle) calculateCycleFraction(frame float64) float64 {
	return math.Mod(math.Mod(frame, s.StrokeFrameCount)+s.StrokeFrameCount, s.StrokeFrameCount) / s.StrokeFrameCount
}

func (s *lapStyle) calculateStroke(frame float64) float64 {
	return calculateBeat(frame, s.StrokeFrameCount, s.StrokeShare)
}

func (s *lapStyle) calculateDrive(loopFrame uint32) drive {
	lapSpeeds := make([]float64, s.LoopFrameCount)
	totalSpeed := 0.0
	for i := uint32(0); i < s.LoopFrameCount; i++ {
		sp := s.GlideSpeed + s.StrokeSpeedGain*s.calculateStroke(float64(i)-s.SpeedLagFrames)
		lapSpeeds[i] = sp
		totalSpeed += sp
	}

	coveredSpeed := 0.0
	for i := uint32(0); i < loopFrame; i++ {
		coveredSpeed += lapSpeeds[i]
	}

	frame := float64(loopFrame)
	stroke := s.calculateStroke(frame)
	wavePhase := StillPose.WavePhaseRadians + frame*Tau/s.WaveFrameCount
	scull := s.ScullSwingRadians * math.Sin(Tau*frame/s.ScullFrameCount+s.ScullPhaseRadians)

	breath := math.Sin(Tau * BreathsPerLoop * frame / float64(s.LoopFrameCount))
	if breath < 0.0 {
		breath = 0.0
	}

	return drive{
		course: Course{
			Type:         CourseLap,
			AngleRadians: -Pi/2.0 + Tau*s.LapCount*coveredSpeed/totalSpeed,
		},
		depth:                 0.0,
		wavePhaseRadians:      wavePhase,
		waveStrength:          s.GlideWaveStrength + s.StrokeWaveGain*stroke,
		tailSpreadRadians:     s.GlideTailSpreadRadians + s.StrokeTailSpreadRadians*stroke*math.Abs(math.Sin(wavePhase)),
		strokeOpeningRadians:  s.calculateStrokeOpening(frame),
		rightFinOffsetRadians: s.InsideFlareRadians + scull,
		feedingPhase:          nil,
		mouthOpening:          BreathOpening * breath * breath,
	}
}

func (s *lapStyle) calculateStrokeOpening(frame float64) float64 {
	switch s.SideFinStrokes.strokeType {
	case sideFinBrake:
		return s.SideFinStrokes.strokeOpeningRadians + s.SideFinStrokes.brakeOpeningRadians*(1.0-s.calculateStroke(frame))
	case sideFinTuckAndSpread:
		cycleFraction := s.calculateCycleFraction(frame)
		if cycleFraction < s.StrokeShare {
			return -s.SideFinStrokes.tuckRadians * s.calculateStroke(frame)
		}
		glideFraction := (cycleFraction - s.StrokeShare) / (1.0 - s.StrokeShare)
		return s.SideFinStrokes.spreadRadians * math.Sin(Pi*glideFraction)
	default:
		return 0.0
	}
}
