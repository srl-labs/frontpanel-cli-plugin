package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d5.webp
var d5 []byte

func registerIXR_D5() {
	platformRegistry["7220 IXR-D5"] = platformDef{
		image: d5,
		layout: portLayout{
			topRowX: []int{247, 334, 435, 521, 624, 710, 812, 898, 1001, 1087, 1189, 1275, 1377, 1464, 1566, 1652},
			botRowX: []int{247, 334, 435, 521, 624, 710, 812, 898, 1001, 1087, 1189, 1275, 1377, 1464, 1566, 1652},
			topY:    60,
			botY:    118,
			width:   85,
			height:  41,
		},
		portRects: d5PortRectangles,
	}
}

func d5PortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// D5 numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2+2)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	rects = append(rects, d5RightSidePortRectangles()...)
	return rects
}

func d5RightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// D5 ports 33/34 are the stacked QSFP-DD cages to the right of the main QSFP28 grid.
	rect33 := inset(image.Rect(1749, 61, 1808, 98), 2, 2)
	rect34 := inset(image.Rect(1749, 119, 1808, 157), 2, 2)
	return []image.Rectangle{rect33, rect34}
}
