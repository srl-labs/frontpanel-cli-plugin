package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d3l.webp
var d3l []byte

func registerIXR_D3L() {
	platformRegistry["7220 IXR-D3L"] = platformDef{
		image: d3l,
		layout: portLayout{
			topRowX: []int{273, 359, 444, 530, 623, 709, 795, 881, 974, 1060, 1146, 1232, 1325, 1411, 1497, 1583},
			botRowX: []int{273, 359, 444, 530, 623, 709, 795, 881, 974, 1060, 1146, 1232, 1325, 1411, 1497, 1583},
			topY:    62,
			botY:    114,
			width:   84,
			height:  41,
		},
		portRects: d3lPortRectangles,
	}
}

func d3lPortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// D3L numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2+2)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	rects = append(rects, d3lRightSidePortRectangles()...)
	return rects
}

func d3lRightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// D3L ports 33/34 are the stacked SFP+ cages to the right of console/mgmt ports.
	rect33 := inset(image.Rect(1751, 63, 1809, 100), 2, 2)
	rect34 := inset(image.Rect(1751, 116, 1809, 153), 2, 2)
	return []image.Rectangle{rect33, rect34}
}
