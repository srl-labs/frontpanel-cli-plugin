package frontpanel

import "image"

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
