package output

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"math"
	"strings"
)

const (
	StaticExportScale  = 2
	FrameWidthPixels   = 256
	FrameHeightPixels  = 256
	StaticWidthPixels  = FrameWidthPixels * StaticExportScale
	StaticHeightPixels = FrameHeightPixels * StaticExportScale
	FramesPerSecond    = 15

	GifColorLimit       = 255
	GifTransparentIndex = 0
)

// KoiFrame holds 256x256 RGBA pixel data (row-major).
type KoiFrame struct {
	Pixels [FrameWidthPixels * FrameHeightPixels][4]byte
}

func (kf *KoiFrame) AsRGBABytes() []byte {
	buf := make([]byte, len(kf.Pixels)*4)
	for i, p := range kf.Pixels {
		buf[i*4] = p[0]
		buf[i*4+1] = p[1]
		buf[i*4+2] = p[2]
		buf[i*4+3] = p[3]
	}
	return buf
}

// EncodePNG encodes a static PNG at 2x scale (512x512) with nearest-neighbor scaling.
func EncodePNG(frame *KoiFrame) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, StaticWidthPixels, StaticHeightPixels))
	for y := 0; y < StaticHeightPixels; y++ {
		ny := y / StaticExportScale
		for x := 0; x < StaticWidthPixels; x++ {
			nx := x / StaticExportScale
			p := frame.Pixels[ny*FrameWidthPixels+nx]
			img.SetRGBA(x, y, color.RGBA{R: p[0], G: p[1], B: p[2], A: p[3]})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeSVG encodes a self-contained SVG at 2x export size with crisp pixel edges.
func EncodeSVG(frame *KoiFrame) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\" shape-rendering=\"crispEdges\">\n",
		StaticWidthPixels, StaticHeightPixels, FrameWidthPixels, FrameHeightPixels))

	for y := 0; y < FrameHeightPixels; y++ {
		x := 0
		for x < FrameWidthPixels {
			p := frame.Pixels[y*FrameWidthPixels+x]
			// Find horizontal run of identical pixels
			run := 1
			for x+run < FrameWidthPixels && frame.Pixels[y*FrameWidthPixels+x+run] == p {
				run++
			}
			if p[3] != 0 {
				opacity := float64(p[3]) / 255.0
				sb.WriteString(fmt.Sprintf("  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"1\" fill=\"#%02x%02x%02x\" fill-opacity=\"%f\"/>\n",
					x, y, run, p[0], p[1], p[2], opacity))
			}
			x += run
		}
	}
	sb.WriteString("</svg>\n")
	return sb.String()
}

// EncodeGIF encodes a seamless animated GIF loop at 2x scale (512x512).
func EncodeGIF(frames []*KoiFrame) ([]byte, error) {
	return EncodeScaledGIF(frames, StaticExportScale)
}

// EncodeScaledGIF encodes a seamless animated GIF loop at the given scale factor.
func EncodeScaledGIF(frames []*KoiFrame, scale int) ([]byte, error) {
	if len(frames) == 0 {
		return nil, errors.New("no frames to encode")
	}
	if scale < 1 {
		scale = 1
	}

	outGif := &gif.GIF{
		Image:     make([]*image.Paletted, len(frames)),
		Delay:     make([]int, len(frames)),
		Disposal:  make([]byte, len(frames)),
		LoopCount: 0,
		Config: image.Config{
			Width:  FrameWidthPixels * scale,
			Height: FrameHeightPixels * scale,
		},
	}

	for frameIdx, kf := range frames {
		boxLeft, boxTop, boxWidth, boxHeight := findDrawnBox(kf)

		type rgbKey [3]byte
		paletteMap := make(map[rgbKey]uint8)
		colorPalette := color.Palette{color.RGBA{0, 0, 0, 0}} // index 0 is transparent

		nativeIndices := make([]uint8, boxWidth*boxHeight)
		idx := 0
		for ny := boxTop; ny < boxTop+boxHeight; ny++ {
			for nx := boxLeft; nx < boxLeft+boxWidth; nx++ {
				p := kf.Pixels[ny*FrameWidthPixels+nx]
				var pIdx uint8
				switch p[3] {
				case 0:
					pIdx = GifTransparentIndex
				case 255:
					k := rgbKey{p[0], p[1], p[2]}
					if existing, exists := paletteMap[k]; exists {
						pIdx = existing
					} else {
						if len(paletteMap) >= GifColorLimit {
							return nil, fmt.Errorf("too many colors in frame %d", frameIdx)
						}
						pIdx = uint8(len(colorPalette))
						paletteMap[k] = pIdx
						colorPalette = append(colorPalette, color.RGBA{R: p[0], G: p[1], B: p[2], A: 255})
					}
				default:
					return nil, fmt.Errorf("partial transparency in frame %d", frameIdx)
				}
				nativeIndices[idx] = pIdx
				idx++
			}
		}

		exportW := int(boxWidth) * scale
		exportH := int(boxHeight) * scale
		palImg := image.NewPaletted(image.Rect(0, 0, exportW, exportH), colorPalette)

		// Set bounds offset
		palImg.Rect = image.Rect(
			int(boxLeft)*scale,
			int(boxTop)*scale,
			int(boxLeft+boxWidth)*scale,
			int(boxTop+boxHeight)*scale,
		)

		for row := 0; row < int(boxHeight); row++ {
			for col := 0; col < int(boxWidth); col++ {
				pIdx := nativeIndices[row*int(boxWidth)+col]
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := int(boxLeft)*scale + col*scale + dx
						py := int(boxTop)*scale + row*scale + dy
						palImg.SetColorIndex(px, py, pIdx)
					}
				}
			}
		}

		delay := int(calculateElapsedHundredths(frameIdx+1) - calculateElapsedHundredths(frameIdx))
		outGif.Image[frameIdx] = palImg
		outGif.Delay[frameIdx] = delay
		outGif.Disposal[frameIdx] = gif.DisposalBackground
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, outGif); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func findDrawnBox(frame *KoiFrame) (left, top, width, height uint32) {
	minX, minY := FrameWidthPixels, FrameHeightPixels
	maxX, maxY := 0, 0
	found := false

	for y := 0; y < FrameHeightPixels; y++ {
		for x := 0; x < FrameWidthPixels; x++ {
			if frame.Pixels[y*FrameWidthPixels+x][3] != 0 {
				found = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if !found {
		return 0, 0, 1, 1
	}
	return uint32(minX), uint32(minY), uint32(maxX - minX + 1), uint32(maxY - minY + 1)
}

func calculateElapsedHundredths(frameIndex int) uint16 {
	return uint16(math.Round(float64(frameIndex*100) / float64(FramesPerSecond)))
}

const previewHead = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>git-koi preview</title>
<style>
  body { margin: 0; font: 14px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; }
  h1, h2 { margin: 0; font-size: 14px; font-weight: 600; }
  header, section { display: grid; gap: 12px; padding: 16px; }
  .light { background: #ffffff; color: #1f2328; }
  .dark { background: #0d1117; color: #f0f6fc; }
  .images { display: flex; flex-wrap: wrap; gap: 16px; }
  figure { margin: 0; }
  img { display: block; max-width: 100%; height: auto; image-rendering: pixelated; }
</style>
</head>
<body>
<header class="light"><h1>`

const previewPanels = `</h1></header>
<section class="light">
<h2>GitHub light</h2>
<div class="images">
<figure><img src="koi.gif" width="512" height="512" alt="The koi swimming"><figcaption>koi.gif</figcaption></figure>
<figure><img src="koi.png" width="512" height="512" alt="The koi, still"><figcaption>koi.png</figcaption></figure>
</div>
</section>
<section class="dark">
<h2>GitHub dark</h2>
<div class="images">
<figure><img src="koi.gif" width="512" height="512" alt="The koi swimming"><figcaption>koi.gif</figcaption></figure>
<figure><img src="koi.png" width="512" height="512" alt="The koi, still"><figcaption>koi.png</figcaption></figure>
</div>
</section>
</body>
</html>
`

// EncodePreviewHTML creates a preview page displaying koi.gif and koi.png side by side.
func EncodePreviewHTML(caption string) string {
	escaped := html.EscapeString(caption)
	return previewHead + escaped + previewPanels
}
