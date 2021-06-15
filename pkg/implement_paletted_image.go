package pkg

import (
	"image"
	"image/color"
)

var _ image.PalettedImage = &Frame{}

func (f *Frame) ColorIndexAt(x, y int) uint8 {
	delta := image.Point{}
	delta = delta.Sub(f.Box.Min)

	pixelIndex := (delta.X + x) + ((delta.Y + y) * f.Width)

	if pixelIndex >= len(f.PixelData) {
		pixelIndex = 0
	}

	return f.PixelData[pixelIndex]
	// return f.direction.PaletteEntries[f.PixelData[pixelIndex]]
}

func (f *Frame) ColorModel() color.Model {
	return color.RGBAModel
}

func (f *Frame) Bounds() image.Rectangle {
	return f.Box
}

func (f *Frame) At(x, y int) color.Color {
	p := *f.direction.dcc.palette
	c := p[f.ColorIndexAt(x, y)]

	return c
}
