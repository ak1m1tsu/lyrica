// gen_icons converts build/appicon.png into ICO files needed by the app.
//
//   - build/tray_icon.ico      — 32×32 single-size ICO (system tray)
//   - build/tray_icon.png      — 32×32 PNG (macOS tray, Darwin build)
//   - build/windows/icon.ico   — 16+32+48+256 multi-size ICO (exe / taskbar / task manager)
//
// Run with: go run ./build/gen_icons
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// appBg is the app background color used to flatten transparent corners.
var appBg = color.NRGBA{R: 15, G: 17, B: 23, A: 255}

func main() {
	src := loadPNG("build/appicon.png")

	// Tray icon: single 32×32
	writeTrayICO("build/tray_icon.ico", src, 32)

	// macOS tray icon: 32×32 PNG (used by tray_icon_darwin.go)
	resized32 := flatten(resizeNearest(src, 32))
	writePNG("build/tray_icon.png", resized32)

	// App / exe icon: multi-size for Windows taskbar + task manager
	writeAppICO("build/windows/icon.ico", src, []int{16, 32, 48, 256})
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

// flatten composites img onto a solid appBg background, eliminating all
// transparent pixels so Windows ICO rendering never shows a white fringe.
func flatten(img image.Image) *image.NRGBA {
	b := img.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, image.NewUniform(appBg), image.Point{}, draw.Src)
	draw.Draw(dst, b, img, b.Min, draw.Over)
	return dst
}

func writeTrayICO(path string, src image.Image, size int) {
	flat := flatten(resizeNearest(src, size))
	ico := buildICO([]icoEntry{{w: size, h: size, data: toBMP(flat)}})
	writeFile(path, ico)
}

func writeAppICO(path string, src image.Image, sizes []int) {
	entries := make([]icoEntry, len(sizes))
	for i, sz := range sizes {
		flat := flatten(resizeNearest(src, sz))
		var data []byte
		if sz >= 256 {
			data = encodePNG(flat)
		} else {
			data = toBMP(flat)
		}
		entries[i] = icoEntry{w: sz, h: sz, data: data}
	}
	writeFile(path, buildICO(entries))
}

// resizeNearest downscales src to size×size using nearest-neighbour sampling.
func resizeNearest(src image.Image, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	sb := src.Bounds()
	sw := float64(sb.Dx())
	sh := float64(sb.Dy())
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			sx := int(float64(dx)/float64(size)*sw) + sb.Min.X
			sy := int(float64(dy)/float64(size)*sh) + sb.Min.Y
			dst.Set(dx, dy, src.At(sx, sy))
		}
	}
	return dst
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

type icoEntry struct {
	w, h int
	data []byte
}

func buildICO(entries []icoEntry) []byte {
	count := len(entries)
	header := []byte{0, 0, 1, 0, byte(count), byte(count >> 8)}

	dirSize := 6 + count*16
	dataOffset := uint32(dirSize)

	var dir, data []byte
	for _, e := range entries {
		w, h := e.w, e.h
		if w >= 256 {
			w, h = 0, 0
		}
		d := make([]byte, 16)
		d[0] = byte(w)
		d[1] = byte(h)
		d[2] = 0
		d[3] = 0
		binary.LittleEndian.PutUint16(d[4:], 1)
		binary.LittleEndian.PutUint16(d[6:], 32)
		binary.LittleEndian.PutUint32(d[8:], uint32(len(e.data)))
		binary.LittleEndian.PutUint32(d[12:], dataOffset)
		dir = append(dir, d...)
		data = append(data, e.data...)
		dataOffset += uint32(len(e.data))
	}

	var result []byte
	result = append(result, header...)
	result = append(result, dir...)
	result = append(result, data...)
	return result
}

// toBMP encodes img as a 32bpp BITMAPINFOHEADER DIB (bottom-up, BGRA).
func toBMP(img *image.NRGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	hdr := make([]byte, 40)
	binary.LittleEndian.PutUint32(hdr[0:], 40)
	binary.LittleEndian.PutUint32(hdr[4:], uint32(w))
	binary.LittleEndian.PutUint32(hdr[8:], uint32(h*2))
	binary.LittleEndian.PutUint16(hdr[12:], 1)
	binary.LittleEndian.PutUint16(hdr[14:], 32)
	binary.LittleEndian.PutUint32(hdr[16:], 0)

	pixels := make([]byte, w*h*4)
	for row := 0; row < h; row++ {
		srcRow := h - 1 - row
		for col := 0; col < w; col++ {
			c := img.NRGBAAt(col, srcRow)
			idx := (row*w + col) * 4
			pixels[idx+0] = c.B
			pixels[idx+1] = c.G
			pixels[idx+2] = c.R
			pixels[idx+3] = c.A
		}
	}

	maskRowBytes := ((w + 31) / 32) * 4
	mask := make([]byte, h*maskRowBytes)

	var result []byte
	result = append(result, hdr...)
	result = append(result, pixels...)
	result = append(result, mask...)
	return result
}

func writeFile(path string, data []byte) {
	if err := os.WriteFile(path, data, 0644); err != nil {
		panic(err)
	}
}
