package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7250-ixr-x4-osfp.webp
var x4osfp []byte

func registerIXR_X4_OSFP() {
	platformRegistry["7250 IXR-X4-OSFP"] = platformDef{
		image:     x4osfp,
		portRects: x4osfpPortRectangles,
	}
}

func x4osfpPortRectangles(_ portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	inset := func(r image.Rectangle) image.Rectangle {
		return image.Rect(r.Min.X+padX, r.Min.Y+padY, r.Max.X-padX, r.Max.Y-padY)
	}

	// Port positions derived from the 7250 IXR-X4 OSFP visio stencil.
	raw := []image.Rectangle{
		image.Rect(203, 32, 316, 99),
		image.Rect(203, 104, 316, 172),
		image.Rect(317, 32, 429, 99),
		image.Rect(317, 104, 429, 172),
		image.Rect(445, 32, 557, 99),
		image.Rect(445, 104, 557, 172),
		image.Rect(558, 32, 671, 99),
		image.Rect(558, 104, 671, 172),
		image.Rect(686, 32, 798, 99),
		image.Rect(686, 104, 798, 172),
		image.Rect(800, 32, 912, 99),
		image.Rect(800, 104, 912, 172),
		image.Rect(928, 32, 1040, 99),
		image.Rect(928, 104, 1040, 172),
		image.Rect(1055, 32, 1168, 99),
		image.Rect(1055, 104, 1168, 172),
		image.Rect(1195, 32, 1308, 99),
		image.Rect(1195, 104, 1308, 172),
		image.Rect(1323, 32, 1435, 99),
		image.Rect(1323, 104, 1435, 172),
		image.Rect(1451, 32, 1563, 99),
		image.Rect(1451, 104, 1563, 172),
		image.Rect(1564, 32, 1677, 99),
		image.Rect(1564, 104, 1677, 172),
		image.Rect(1692, 32, 1804, 99),
		image.Rect(1692, 104, 1804, 172),
		image.Rect(1806, 32, 1918, 99),
		image.Rect(1806, 104, 1918, 172),
		image.Rect(1934, 32, 2046, 99),
		image.Rect(1934, 104, 2046, 172),
		image.Rect(2047, 32, 2160, 99),
		image.Rect(2047, 104, 2160, 172),
	}

	rects := make([]image.Rectangle, len(raw))
	for i, r := range raw {
		rects[i] = inset(r)
	}

	return rects
}
