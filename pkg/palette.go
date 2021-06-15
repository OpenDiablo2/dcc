package pkg

import (
	"image/color"
	"math"
)

const numColorsInPalette = 256

func DefaultPalette() *color.Palette {
	p := make(color.Palette, numColorsInPalette)

	for idx := range p {
		c := color.RGBA{}
		c.R, c.G, c.B, c.A = uint8(idx), uint8(idx), uint8(idx), uint8(math.MaxUint8)

		p[idx] = c
	}

	return &p
}
