package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7250-ixr-x1b.webp
var x1b []byte

func registerIXR_X1B() {
	platformRegistry["7250 IXR-X1B"] = platformDef{
		image: x1b,
		layout: portLayout{
			topRowX: []int{291, 395, 499, 603, 707, 811, 915, 1019, 1122, 1227, 1330, 1434, 1538, 1642, 1746, 1850, 1954, 2058},
			botRowX: []int{291, 395, 499, 603, 707, 811, 915, 1019, 1122, 1227, 1330, 1434, 1538, 1642, 1746, 1850, 1954, 2058},
			topY:    50,
			botY:    123,
			width:   93,
			height:  44,
		},
		portRects: x1bPortRectangles,
	}
}

func x1bPortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// X1B numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	return rects
}
