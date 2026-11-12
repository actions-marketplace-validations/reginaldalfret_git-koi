package coat_test

import (
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/coat"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
)

func TestParseKoiBody(t *testing.T) {
	validBodies := map[string]coat.KoiBody{
		"white":    coat.BodyWhite,
		"black":    coat.BodyBlack,
		"asagi":    coat.BodyAsagi,
		"tea":      coat.BodyTea,
		"gold":     coat.BodyGold,
		"platinum": coat.BodyPlatinum,
	}

	for name, expected := range validBodies {
		b, ok := coat.ParseKoiBody(name)
		if !ok || b != expected {
			t.Errorf("expected %s -> %v, got %v (ok=%v)", name, expected, b, ok)
		}
		if b.String() != name {
			t.Errorf("expected %v.String() == %s, got %s", b, name, b.String())
		}
	}

	invalidBody, ok := coat.ParseKoiBody("unknown-color")
	if ok || invalidBody != coat.BodyWhite {
		t.Errorf("expected false and BodyWhite fallback for unknown body")
	}
}

func TestCoatRampsAndMetallicSheen(t *testing.T) {
	bodies := []coat.KoiBody{
		coat.BodyWhite,
		coat.BodyBlack,
		coat.BodyAsagi,
		coat.BodyTea,
		coat.BodyGold,
		coat.BodyPlatinum,
	}

	for _, b := range bodies {
		ground := b.SelectGroundRamp()
		if len(ground) == 0 {
			t.Errorf("body %s has empty ground ramp", b.String())
		}
		fin := b.SelectFinRamp()
		if len(fin) == 0 {
			t.Errorf("body %s has empty fin ramp", b.String())
		}
		streak := b.SelectStreakFinRamp()
		if len(streak) == 0 {
			t.Errorf("body %s has empty streak fin ramp", b.String())
		}
	}

	// Gold and Platinum have metallic sheen
	if s := coat.BodyGold.CalculateMetallicSheen(0.35, 0.0); s <= 0 {
		t.Errorf("expected positive metallic sheen for Gold, got %f", s)
	}
	if s := coat.BodyPlatinum.CalculateMetallicSheen(0.35, 0.0); s <= 0 {
		t.Errorf("expected positive metallic sheen for Platinum, got %f", s)
	}

	// White has zero metallic sheen
	if s := coat.BodyWhite.CalculateMetallicSheen(0.35, 0.0); s != 0 {
		t.Errorf("expected 0 metallic sheen for White, got %f", s)
	}
	_ = history.DateFormat
}
