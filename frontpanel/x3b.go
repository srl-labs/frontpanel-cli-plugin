package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7250-ixr-x3b.webp
var x3b []byte

func registerIXR_X3B() {
	platformRegistry["7250 IXR-X3B"] = platformDef{
		image: x3b,
		layout: portLayout{
			topRowX: []int{291, 395, 499, 603, 707, 811, 914, 1018, 1122, 1226, 1330, 1434, 1538, 1642, 1746, 1850, 1954, 2058},
			botRowX: []int{291, 395, 499, 603, 707, 811, 914, 1018, 1122, 1226, 1330, 1434, 1538, 1642, 1746, 1850, 1954, 2058},
			topY:    50,
			botY:    123,
			width:   93,
			height:  44,
		},
		portRects: x3bPortRectangles,
	}
}

func x3bPortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// X3B numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	return rects
}
