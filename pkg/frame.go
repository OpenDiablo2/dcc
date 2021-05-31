package pkg

import (
	"github.com/gravestench/bitstream"
	"image"
	"image/color"
	"log"
)

var _ image.PalettedImage = &Frame{}

type Frame struct {
	direction             *Direction
	Box                   image.Rectangle
	Cells                 []Cell
	PixelData             []byte
	Width                 int
	Height                int
	XOffset               int
	YOffset               int
	NumberOfOptionalBytes int
	NumberOfCodedBytes    int
	HorizontalCellCount   int
	VerticalCellCount     int
	FrameIsBottomUp       bool
	valid                 bool
}

func (f *Frame) decodeFrameHeader(stream *bitstream.BitStream) (err error) {
	_, err = stream.Next(f.direction.Variable0Bits).Bits().AsUInt32()
	if err != nil {
		return err
	}

	if width, err := stream.Next(f.direction.WidthBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		f.Width = int(width)
	}

	if height, err := stream.Next(f.direction.HeightBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		f.Height = int(height)
	}

	f.XOffset, err = stream.Next(f.direction.XOffsetBits).Bits().AsInt()
	if err != nil {
		return err
	}

	f.YOffset, err = stream.Next(f.direction.YOffsetBits).Bits().AsInt()
	if err != nil {
		return err
	}

	if numOptionBytes, err := stream.Next(f.direction.OptionalDataBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		f.NumberOfOptionalBytes = int(numOptionBytes)
	}

	if codedBytes, err := stream.Next(f.direction.CodedBytesBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		f.NumberOfCodedBytes = int(codedBytes)
	}

	if f.FrameIsBottomUp, err = stream.Next(1).Bits().AsBool(); err != nil {
		return err
	}

	if f.FrameIsBottomUp {
		log.Panic("Bottom up frames are not implemented.")
	} else {
		min := image.Point{f.XOffset, f.YOffset - f.Height + 1}
		max := image.Point{min.X + f.Width, min.Y + f.Height}
		f.Box = image.Rectangle{min, max}
	}

	f.valid = true

	return nil
}

func (f *Frame) recalculateCells() {
	// nolint:gomnd // constant
	var w = 4 - ((f.Box.Min.X - f.direction.Box.Min.X) % 4) // Width of the first column (in pixels)

	if (f.Width - w) <= 1 {
		f.HorizontalCellCount = 1
	} else {
		tmp := f.Width - w - 1
		f.HorizontalCellCount = 2 + (tmp / 4) //nolint:gomnd // magic math

		// nolint:gomnd // constant
		if (tmp % 4) == 0 {
			f.HorizontalCellCount--
		}
	}

	// Height of the first column (in pixels)
	h := 4 - ((f.Box.Min.Y - f.direction.Box.Min.Y) % 4) //nolint:gomnd // data decode

	if (f.Height - h) <= 1 {
		f.VerticalCellCount = 1
	} else {
		tmp := f.Height - h - 1
		f.VerticalCellCount = 2 + (tmp / 4) //nolint:gomnd // data decode

		// nolint:gomnd // constant
		if (tmp % 4) == 0 {
			f.VerticalCellCount--
		}
	}
	// Calculate the cell widths and heights
	cellWidths := make([]int, f.HorizontalCellCount)
	if f.HorizontalCellCount == 1 {
		cellWidths[0] = f.Width
	} else {
		cellWidths[0] = w
		for i := 1; i < (f.HorizontalCellCount - 1); i++ {
			cellWidths[i] = 4
		}

		// nolint:gomnd // constants
		cellWidths[f.HorizontalCellCount-1] = f.Width - w - (4 * (f.HorizontalCellCount - 2))
	}

	cellHeights := make([]int, f.VerticalCellCount)
	if f.VerticalCellCount == 1 {
		cellHeights[0] = f.Height
	} else {
		cellHeights[0] = h
		for i := 1; i < (f.VerticalCellCount - 1); i++ {
			cellHeights[i] = 4
		}

		// nolint:gomnd // constants
		cellHeights[f.VerticalCellCount-1] = f.Height - h - (4 * (f.VerticalCellCount - 2))
	}

	f.Cells = make([]Cell, f.HorizontalCellCount*f.VerticalCellCount)
	offsetY := f.Box.Min.Y - f.direction.Box.Min.Y

	for y := 0; y < f.VerticalCellCount; y++ {
		offsetX := f.Box.Min.X - f.direction.Box.Min.X

		for x := 0; x < f.HorizontalCellCount; x++ {
			f.Cells[x+(y*f.HorizontalCellCount)] = Cell{
				XOffset: offsetX,
				YOffset: offsetY,
				Width:   cellWidths[x],
				Height:  cellHeights[y],
			}

			offsetX += cellWidths[x]
		}

		offsetY += cellHeights[y]
	}
}

func (f *Frame) ColorIndexAt(x, y int) uint8 {
	panic("implement me")
}

func (f *Frame) ColorModel() color.Model {
	panic("implement me")
}

func (f *Frame) Bounds() image.Rectangle {
	return f.Box
}

func (f *Frame) At(x, y int) color.Color {
	panic("implement me")
}
