package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7250-ixr-x4.webp
var x4 []byte

func registerIXR_X4() {
	platformRegistry["7250 IXR-X4"] = platformDef{
		image:     x4,
		portRects: x4PortRectangles,
	}
}

func x4PortRectangles(_ portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	inset := func(r image.Rectangle) image.Rectangle {
		return image.Rect(r.Min.X+padX, r.Min.Y+padY, r.Max.X-padX, r.Max.Y-padY)
	}

	// Port positions derived from the 7250 IXR-X4 QSFP-DD visio stencil.
	raw := []image.Rectangle{
		image.Rect(297, 38, 393, 85),
		image.Rect(297, 122, 393, 169),
		image.Rect(413, 38, 508, 85),
		image.Rect(413, 122, 508, 169),
		image.Rect(527, 38, 623, 85),
		image.Rect(527, 122, 623, 169),
		image.Rect(651, 38, 747, 85),
		image.Rect(651, 122, 747, 169),
		image.Rect(766, 38, 862, 85),
		image.Rect(766, 122, 862, 169),
		image.Rect(881, 38, 977, 85),
		image.Rect(881, 122, 977, 169),
		image.Rect(1000, 38, 1095, 85),
		image.Rect(1000, 122, 1095, 169),
		image.Rect(1115, 38, 1211, 85),
		image.Rect(1115, 122, 1211, 169),
		image.Rect(1230, 38, 1325, 85),
		image.Rect(1230, 122, 1325, 169),
		image.Rect(1345, 38, 1440, 85),
		image.Rect(1345, 122, 1440, 169),
		image.Rect(1473, 38, 1569, 85),
		image.Rect(1473, 122, 1569, 169),
		image.Rect(1588, 38, 1684, 85),
		image.Rect(1588, 122, 1684, 169),
		image.Rect(1703, 38, 1799, 85),
		image.Rect(1703, 122, 1799, 169),
		image.Rect(1827, 38, 1923, 85),
		image.Rect(1827, 122, 1923, 169),
		image.Rect(1942, 38, 2038, 85),
		image.Rect(1942, 122, 2038, 169),
		image.Rect(2057, 38, 2153, 85),
		image.Rect(2057, 122, 2153, 169),
	}

	rects := make([]image.Rectangle, len(raw))
	for i, r := range raw {
		rects[i] = inset(r)
	}

	return rects
}
