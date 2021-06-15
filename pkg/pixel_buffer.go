package pkg

type PixelBuffer []PixelBufferEntry

func (pb PixelBuffer) init() {
	for idx := range pb {
		pb[idx].Frame = none
		pb[idx].FrameCellIndex = none
	}
}

func newPixelBuffer(cellsWide, cellsHigh int) PixelBuffer {
	pb := make(PixelBuffer, cellsWide * cellsHigh)

	pb.init()

	return pb
}

type PixelBufferEntry struct {
	Value          [4]byte
	Frame          int
	FrameCellIndex int
}
