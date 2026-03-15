package frontpanel

import (
	_ "embed"
	"image"
)

//go:embed images/7730-sxr-1x-44s.webp
var sxr1x44s []byte

func registerSXR_1X_44S() {
	platformRegistry["7730 SXR-1x-44S"] = platformDef{
		image: sxr1x44s,
		layout: portLayout{
			topY:   47,
			botY:   108,
			width:  58,
			height: 40,
		},
		portRects: sxr1x44sPortRectangles,
	}
}

func sxr1x44sPortRectangles(layout portLayout) []image.Rectangle {
	padX := 2
	padY := 2

	type col struct {
		x, w int
	}

	columns := []col{
		{269, 58},
		{329, 58},
		{390, 58},
		{450, 58},
		{510, 58},
		{571, 58},
		{672, 58},
		{733, 58},
		{793, 58},
		{854, 58},
		{930, 86},
		{1032, 58},
		{1092, 58},
		{1153, 58},
		{1213, 58},
		{1314, 58},
		{1374, 58},
		{1435, 58},
		{1495, 58},
		{1556, 58},
		{1616, 58},
		{1688, 85},
	}

	rects := make([]image.Rectangle, 0, len(columns)*2)
	for _, c := range columns {
		rects = append(rects,
			image.Rect(c.x+padX, layout.topY+padY, c.x+c.w-padX, layout.topY+layout.height-padY),
			image.Rect(c.x+padX, layout.botY+padY, c.x+c.w-padX, layout.botY+layout.height-padY),
		)
	}

	return rects
}
