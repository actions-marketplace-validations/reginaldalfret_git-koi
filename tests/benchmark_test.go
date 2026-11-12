package tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/aquarium"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/coat"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/output"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/render"
)

func loadDevFixture(b *testing.B) *koi.KoiDevelopment {
	b.Helper()
	data, err := os.ReadFile("fixtures/account-with-history.json")
	if err != nil {
		b.Fatalf("reading fixture: %v", err)
	}
	var snapshot history.AccountSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		b.Fatalf("unmarshaling snapshot: %v", err)
	}
	dev, err := koi.DeriveKoiDevelopment(&snapshot)
	if err != nil {
		b.Fatalf("deriving dev: %v", err)
	}
	return dev
}

func BenchmarkRenderSingleKoi(b *testing.B) {
	dev := loadDevFixture(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame := render.RenderKoi(dev, coat.BodyWhite)
		if frame == nil {
			b.Fatal("nil frame")
		}
	}
}

func BenchmarkEncodeGIF(b *testing.B) {
	dev := loadDevFixture(b)
	frames := make([]*output.KoiFrame, 15)
	for i := range frames {
		frames[i] = render.RenderKoi(dev, coat.BodyWhite)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := output.EncodeScaledGIF(frames, 1)
		if err != nil || len(data) == 0 {
			b.Fatal("encode error")
		}
	}
}

func BenchmarkRenderAquariumGifs(b *testing.B) {
	dev := loadDevFixture(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gifs, err := aquarium.RenderAquariumGifs(dev)
		if err != nil || len(gifs) != 6 {
			b.Fatalf("RenderAquariumGifs failed: %v", err)
		}
	}
}

func BenchmarkGenerateAquariumSVG(b *testing.B) {
	dev := loadDevFixture(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svg, err := aquarium.GenerateAquariumSVG(dev, nil)
		if err != nil || len(svg) == 0 {
			b.Fatalf("GenerateAquariumSVG failed: %v", err)
		}
	}
}
