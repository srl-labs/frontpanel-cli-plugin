package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d3.webp
var d3 []byte

func registerD3() {
	platformRegistry["7220 IXR-D3"] = platformDef{
		image: d3,
		layout: portLayout{
			topRowX: []int{
				268, 353, 439, 524, 632, 717, 803, 888,
				996, 1082, 1167, 1253, 1360, 1446, 1531, 1616,
			},
			botRowX: []int{
				268, 353, 439, 524, 632, 717, 803, 888,
				996, 1082, 1167, 1253, 1360, 1446, 1531, 1616,
			},
			topY:   44,
			botY:   103,
			width:  85,
			height: 41,
		},
		portRects: d3PortRectangles,
	}
}

func d3PortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if bn := len(layout.botRowX); bn < n {
		n = bn
	}

	rects := make([]image.Rectangle, 0, 2+n*2)

	// Ports 1-2: stacked cages on the left, top then bottom.
	rects = append(rects, d3LeftPortRectangles()...)

	// Ports 3-34: 4 blocks × 4 cages, top then bottom at each column, left to right.
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	return rects
}

func d3LeftPortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// D3 ports 1/2 are the stacked cages on the left side.
	rect1 := inset(image.Rect(200, 44, 259, 85), 2, 2)
	rect2 := inset(image.Rect(200, 103, 259, 144), 2, 2)
	return []image.Rectangle{rect1, rect2}
}
