package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7220-ixr-d2l.webp
var d2l []byte

func registerD2L() {
	platformRegistry["7220 IXR-D2L"] = platformDef{
		image: d2l,
		layout: portLayout{
			topRowX: []int{172, 233, 312, 374, 453, 514, 593, 655, 734, 795, 875, 936, 1015, 1077, 1156, 1217},
			botRowX: []int{172, 233, 312, 374, 453, 514, 593, 655, 734, 795, 875, 936, 1015, 1077, 1156, 1217},
			topY:    13,
			botY:    73,
			width:   59,
			height:  44,
		},
		portRects: d2lPortRectangles,
	}
}

func d2lPortRectangles(layout portLayout) []image.Rectangle {
	if len(layout.topRowX) < 2 {
		return nil
	}

	padX := 2
	padY := 2
	topY := layout.topY
	midY := layout.botY
	botY := 133 // D2L third row cages
	botH := 42  // D2L third row cage height
	topH := layout.height
	midH := layout.height

	rectFor := func(x int, y int, h int) image.Rectangle {
		return image.Rect(x+padX, y+padY, x+layout.width-padX, y+h-padY)
	}

	rects := make([]image.Rectangle, 0, (len(layout.topRowX)/2)*6)
	for pair := 0; pair+1 < len(layout.topRowX); pair += 2 {
		xLeft := layout.topRowX[pair]
		xRight := layout.topRowX[pair+1]

		// D2L numbering order per 2-column block: 1 4 / 2 5 / 3 6
		rects = append(rects,
			rectFor(xLeft, topY, topH),
			rectFor(xLeft, midY, midH),
			rectFor(xLeft, botY, botH),
			rectFor(xRight, topY, topH),
			rectFor(xRight, midY, midH),
			rectFor(xRight, botY, botH),
		)
	}

	rects = append(rects, d2lRightSidePortRectangles()...)
	return rects
}

func d2lRightSidePortRectangles() []image.Rectangle {
	inset := func(r image.Rectangle, dx int, dy int) image.Rectangle {
		return image.Rect(r.Min.X+dx, r.Min.Y+dy, r.Max.X-dx, r.Max.Y-dy)
	}

	splitDualCage := func(r image.Rectangle) (image.Rectangle, image.Rectangle) {
		mid := r.Min.X + (r.Dx() / 2)
		left := image.Rect(r.Min.X, r.Min.Y, mid, r.Max.Y)
		right := image.Rect(mid, r.Min.Y, r.Max.X, r.Max.Y)
		// Keep labels out of the center divider between dual cages.
		if left.Dx() > 4 {
			left.Max.X--
		}
		if right.Dx() > 4 {
			right.Min.X++
		}
		return left, right
	}

	topDualCages := []image.Rectangle{
		image.Rect(1379, 77, 1550, 118),
		image.Rect(1568, 77, 1739, 118),
	}
	bottomDualCages := []image.Rectangle{
		image.Rect(1379, 135, 1550, 176),
		image.Rect(1568, 135, 1739, 176),
	}

	rects := make([]image.Rectangle, 0, 10)
	for i := range topDualCages {
		top := inset(topDualCages[i], 2, 2)
		bottom := inset(bottomDualCages[i], 2, 2)

		topLeft, topRight := splitDualCage(top)
		bottomLeft, bottomRight := splitDualCage(bottom)

		// D2L ports 49..56 are numbered odd on top, even below, left to right.
		rects = append(rects, topLeft, bottomLeft, topRight, bottomRight)
	}

	// D2L ports 57/58 are the stacked cages at the far right.
	rect57 := inset(image.Rect(1759, 13, 1819, 57), 2, 2)
	rect58 := inset(image.Rect(1759, 73, 1819, 117), 2, 2)
	rects = append(rects, rect57, rect58)

	return rects
}
