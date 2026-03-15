package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d1.webp
var d1 []byte

func registerIXR_D1() {
	platformRegistry["7220 IXR-D1"] = platformDef{
		image: d1,
		layout: portLayout{
			topRowX: []int{
				150, 208, 267, 326, 385, 445, 516, 576, 635, 694, 754, 813, 895, 955, 1015, 1073,
				1133, 1192, 1264, 1323, 1383, 1441, 1501, 1560,
			},
			botRowX: []int{
				150, 208, 267, 326, 385, 445, 516, 576, 635, 694, 754, 813, 895, 955, 1015, 1073,
				1133, 1192, 1264, 1323, 1383, 1441, 1501, 1560,
			},
			topY:   46,
			botY:   107,
			width:  53,
			height: 46,
		},
		portRects: d1PortRectangles,
	}
}

func d1PortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// D1 numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2+4)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	rects = append(rects, d1RightSidePortRectangles()...)
	return rects
}

func d1RightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// D1 ports 49-52 are four QSFP28 cages in a horizontal row at the bottom-right.
	rect49 := inset(image.Rect(1636, 130, 1695, 152), 2, 2)
	rect50 := inset(image.Rect(1697, 130, 1756, 152), 2, 2)
	rect51 := inset(image.Rect(1772, 130, 1831, 152), 2, 2)
	rect52 := inset(image.Rect(1833, 130, 1892, 152), 2, 2)
	return []image.Rectangle{rect49, rect50, rect51, rect52}
}
