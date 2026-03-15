package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7215-ixs-a1.webp
var a1 []byte

func registerIXS_A1() {
	platformRegistry["7215 IXS-A1"] = platformDef{
		image: a1,
		layout: portLayout{
			topRowX: []int{
				178, 239, 299, 359, 419, 479, 540, 600, 682, 742, 802, 863, 922, 983, 1043, 1103,
				1178, 1238, 1299, 1359, 1419, 1479, 1539, 1600,
			},
			botRowX: []int{
				178, 239, 299, 359, 419, 479, 540, 600, 682, 742, 802, 863, 922, 983, 1043, 1103,
				1178, 1238, 1299, 1359, 1419, 1479, 1539, 1600,
			},
			topY:   42,
			botY:   106,
			width:  53,
			height: 48,
		},
		portRects: a1PortRectangles,
	}
}

func a1PortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	rectFor := func(x int, y int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+layout.height-padY)
	}

	n := len(layout.topRowX)
	if len(layout.botRowX) < n {
		n = len(layout.botRowX)
	}

	// A1 numbering: top then bottom at each column, left to right.
	rects := make([]image.Rectangle, 0, n*2+4)
	for i := 0; i < n; i++ {
		rects = append(rects,
			rectFor(layout.topRowX[i], layout.topY),
			rectFor(layout.botRowX[i], layout.botY),
		)
	}

	rects = append(rects, a1RightSidePortRectangles()...)
	return rects
}

func a1RightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	// A1 ports 49-52 are SFP28 cages in a 2x2 grid (top-to-bottom, left-to-right).
	rect49 := inset(image.Rect(1670, 45, 1729, 84), 2, 2)
	rect50 := inset(image.Rect(1670, 113, 1729, 152), 2, 2)
	rect51 := inset(image.Rect(1731, 45, 1790, 84), 2, 2)
	rect52 := inset(image.Rect(1731, 113, 1790, 152), 2, 2)
	return []image.Rectangle{rect49, rect50, rect51, rect52}
}
