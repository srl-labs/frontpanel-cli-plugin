package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d2.webp
var d2 []byte

func registerIXR_D2() {
	platformRegistry["7220 IXR-D2"] = platformDef{
		image: d2,
		layout: portLayout{
			topRowX: []int{
				129, 188, 248, 308, 368, 427, 487, 547, 613, 672, 732, 792,
				851, 911, 971, 1030, 1096, 1156, 1216, 1275, 1335, 1395, 1455, 1514,
			},
			botRowX: []int{
				129, 188, 248, 308, 368, 427, 487, 547, 613, 672, 732, 792,
				851, 911, 971, 1030, 1096, 1156, 1216, 1275, 1335, 1395, 1455, 1514,
			},
			topY:   44,
			botY:   103,
			width:  59,
			height: 41,
		},
		portRects: d2PortRectangles,
	}
}

func d2PortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if bn := len(layout.botRowX); bn < n {
		n = bn
	}

	// D2 numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2+8)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	rects = append(rects, d2RightSidePortRectangles()...)
	return rects
}

func d2RightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// D2 ports 49-56: top then bottom at each column, left to right.
	topCages := []image.Rectangle{
		image.Rect(1580, 44, 1666, 85),
		image.Rect(1668, 44, 1753, 85),
		image.Rect(1755, 44, 1840, 85),
		image.Rect(1842, 44, 1927, 85),
	}
	bottomCages := []image.Rectangle{
		image.Rect(1580, 103, 1666, 144),
		image.Rect(1668, 103, 1753, 144),
		image.Rect(1755, 103, 1840, 144),
		image.Rect(1842, 103, 1927, 144),
	}

	rects := make([]image.Rectangle, 0, 8)
	for i := range topCages {
		rects = append(rects,
			inset(topCages[i], 2, 2),
			inset(bottomCages[i], 2, 2),
		)
	}

	return rects
}
