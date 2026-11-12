package output_test

import (
	"bytes"
	"image/gif"
	"image/png"
	"strings"
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/output"
)

func createSampleFrame() *output.KoiFrame {
	frame := &output.KoiFrame{}
	// Draw a small 20x20 opaque rectangle in the middle
	for y := 100; y < 120; y++ {
		for x := 100; x < 120; x++ {
			frame.Pixels[y*output.FrameWidthPixels+x] = [4]byte{200, 50, 50, 255}
		}
	}
	return frame
}

func TestEncodePNG(t *testing.T) {
	frame := createSampleFrame()
	pngBytes, err := output.EncodePNG(frame)
	if err != nil {
		t.Fatalf("EncodePNG failed: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("failed to decode generated PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != output.StaticWidthPixels || bounds.Dy() != output.StaticHeightPixels {
		t.Errorf("expected PNG dimensions %dx%d, got %dx%d",
			output.StaticWidthPixels, output.StaticHeightPixels, bounds.Dx(), bounds.Dy())
	}
}

func TestEncodeGIF(t *testing.T) {
	frames := []*output.KoiFrame{
		createSampleFrame(),
		createSampleFrame(),
		createSampleFrame(),
	}

	gifBytes, err := output.EncodeGIF(frames)
	if err != nil {
		t.Fatalf("EncodeGIF failed: %v", err)
	}

	decoded, err := gif.DecodeAll(bytes.NewReader(gifBytes))
	if err != nil {
		t.Fatalf("failed to decode generated GIF: %v", err)
	}

	if len(decoded.Image) != len(frames) {
		t.Errorf("expected %d GIF frames, got %d", len(frames), len(decoded.Image))
	}

	if decoded.Config.Width != output.StaticWidthPixels || decoded.Config.Height != output.StaticHeightPixels {
		t.Errorf("expected GIF canvas %dx%d, got %dx%d",
			output.StaticWidthPixels, output.StaticHeightPixels, decoded.Config.Width, decoded.Config.Height)
	}

	if decoded.LoopCount != 0 {
		t.Errorf("expected loop count 0 (infinite), got %d", decoded.LoopCount)
	}

	for i, palImg := range decoded.Image {
		// Index 0 in palette must be transparent
		if len(palImg.Palette) == 0 {
			t.Errorf("frame %d has empty palette", i)
			continue
		}
		_, _, _, a := palImg.Palette[output.GifTransparentIndex].RGBA()
		if a != 0 {
			t.Errorf("frame %d palette index 0 must be transparent, got alpha=%d", i, a)
		}

		if decoded.Disposal[i] != gif.DisposalBackground {
			t.Errorf("frame %d expected disposal background (%d), got %d",
				i, gif.DisposalBackground, decoded.Disposal[i])
		}

		if decoded.Delay[i] <= 0 {
			t.Errorf("frame %d non-positive delay %d", i, decoded.Delay[i])
		}
	}
}

func TestEncodePreviewHTML(t *testing.T) {
	caption := `Rest for 2026-11-10, "white" & 'gold' <body & test>`
	html := output.EncodePreviewHTML(caption)

	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("expected HTML5 doctype")
	}
	if strings.Contains(html, "<body & test>") {
		t.Errorf("caption was not properly escaped in HTML output")
	}
	if !strings.Contains(html, "&amp;") || !strings.Contains(html, "&#34;") {
		t.Errorf("expected escaped entities in HTML: %s", html)
	}
}
