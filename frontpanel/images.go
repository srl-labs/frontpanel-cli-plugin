package frontpanel

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	_ "github.com/HugoSmits86/nativewebp"

	"github.com/dolmen-go/kittyimg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed images/d2l.webp
var d2l []byte

//go:embed images/d3l.webp
var d3l []byte

//go:embed images/d5.webp
var d5 []byte

type ChassisDef struct {
	Image []byte
	URL   string
}

var chassisImages = map[string]ChassisDef{
	"7220 IXR-D2L": {
		Image: d2l,
		URL:   "https://go.srlinux.dev/fp-d2l",
	},
	"7220 IXR-D3L": {
		Image: d3l,
		URL:   "https://go.srlinux.dev/fp-d3l",
	},
	"7220 IXR-D5": {
		Image: d5,
		URL:   "https://go.srlinux.dev/fp-d5",
	},
}

type imageProtocol string

const (
	imageProtocolAuto  imageProtocol = "auto"
	imageProtocolKitty imageProtocol = "kitty"
	imageProtocolITerm imageProtocol = "iterm"
)

var lastNumberPattern = regexp.MustCompile(`\d+`)

var (
	portUpColor              = color.RGBA{R: 33, G: 201, B: 110, A: 255}
	portAdminUpOperDownColor = color.RGBA{R: 245, G: 130, B: 32, A: 255}
)

var (
	labelFontOnce sync.Once
	labelFont     *opentype.Font
	labelFontErr  error
	labelFaceMu   sync.Mutex
	labelFaces    = map[int]font.Face{}
)

type portLayout struct {
	topRowX []int
	botRowX []int
	topY    int
	botY    int
	width   int
	height  int
}

var chassisPortLayouts = map[string]portLayout{
	"7220 IXR-D2L": {
		topRowX: []int{172, 233, 312, 374, 453, 514, 593, 655, 734, 795, 875, 936, 1015, 1077, 1156, 1217},
		botRowX: []int{172, 233, 312, 374, 453, 514, 593, 655, 734, 795, 875, 936, 1015, 1077, 1156, 1217},
		topY:    13,
		botY:    73,
		width:   59,
		height:  44,
	},
	"7220 IXR-D3L": {
		topRowX: []int{123, 154, 184, 215, 251, 282, 312, 343, 379, 410, 440, 471, 507, 538, 568, 599},
		botRowX: []int{123, 154, 184, 215, 251, 282, 312, 343, 379, 410, 440, 471, 507, 538, 568, 599},
		topY:    37,
		botY:    57,
		width:   29,
		height:  13,
	},
	"7220 IXR-D5": {
		topRowX: []int{108, 139, 179, 210, 251, 282, 323, 354, 395, 426, 467, 498, 538, 569, 610, 641},
		botRowX: []int{108, 139, 179, 210, 251, 282, 323, 354, 395, 426, 467, 498, 538, 569, 610, 641},
		topY:    33,
		botY:    56,
		width:   30,
		height:  14,
	},
}

func parseImageProtocol(v string) imageProtocol {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", string(imageProtocolAuto):
		return imageProtocolAuto
	case string(imageProtocolKitty):
		return imageProtocolKitty
	case "iip", "osc1337", string(imageProtocolITerm):
		return imageProtocolITerm
	default:
		return imageProtocolAuto
	}
}

func Print(chassisType string) {
	PrintWithProtocol(chassisType, string(imageProtocolAuto))
}

func PrintWithProtocol(chassisType string, protocol string) {
	PrintWithProtocolAndPortStatesAndLabels(chassisType, protocol, nil, false)
}

func ParsePortStatesJSON(payload string) map[string]string {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil
	}

	portStates := map[string]string{}
	if err := json.Unmarshal([]byte(payload), &portStates); err != nil {
		return nil
	}

	return portStates
}

func ParsePortLabelsValue(payload string) bool {
	switch strings.ToLower(strings.TrimSpace(payload)) {
	case "1", "true", "yes", "y", "on", "enable", "enabled":
		return true
	default:
		return false
	}
}

func PrintWithProtocolAndPortStates(chassisType string, protocol string, portStates map[string]string) {
	PrintWithProtocolAndPortStatesAndLabels(chassisType, protocol, portStates, false)
}

func PrintWithProtocolAndPortStatesAndLabels(chassisType string, protocol string, portStates map[string]string, portLabels bool) {
	printWithProtocol(chassisType, parseImageProtocol(protocol), portStates, portLabels)
}

func printWithProtocol(chassisType string, protocol imageProtocol, portStates map[string]string, portLabels bool) {
	if chassisDef, ok := chassisImages[chassisType]; ok {
		f := bytes.NewReader(chassisDef.Image)
		img, _, err := image.Decode(f)
		if err != nil {
			return
		}

		img = applyPortStateOverlay(chassisType, img, portStates)
		if portLabels {
			img = applyPortLabelOverlay(chassisType, img)
		}

		if protocol == imageProtocolITerm {
			if err := printITermImage(os.Stdout, img, chassisType); err != nil {
				_ = kittyimg.Fprintln(os.Stdout, img)
			}
			return
		}

		_ = kittyimg.Fprintln(os.Stdout, img)

	} else {
		fmt.Println("not supported")
	}
}

func applyPortStateOverlay(chassisType string, base image.Image, portStates map[string]string) image.Image {
	if len(portStates) == 0 {
		return base
	}

	layout, ok := chassisPortLayouts[chassisType]
	if !ok {
		return base
	}

	rects := portRectsForChassis(chassisType, layout)
	if len(rects) == 0 {
		return base
	}

	dst := image.NewRGBA(base.Bounds())
	draw.Draw(dst, dst.Bounds(), base, base.Bounds().Min, draw.Src)

	for ifName, state := range portStates {
		portIndex, ok := parseInterfaceIndex(ifName)
		if !ok || portIndex < 1 || portIndex > len(rects) {
			continue
		}

		rect := rects[portIndex-1]
		clr, ok := stateOverlayColor(state)
		if !ok {
			continue
		}

		drawPortOverlay(dst, rect, clr)
	}

	return dst
}

func applyPortLabelOverlay(chassisType string, base image.Image) image.Image {
	layout, ok := chassisPortLayouts[chassisType]
	if !ok {
		return base
	}

	rects := portRectsForChassis(chassisType, layout)
	if len(rects) == 0 {
		return base
	}

	dst := image.NewRGBA(base.Bounds())
	draw.Draw(dst, dst.Bounds(), base, base.Bounds().Min, draw.Src)

	for idx, rect := range rects {
		drawPortLabel(dst, rect, fmt.Sprintf("1/%d", idx+1))
	}

	return dst
}

func portRectsForChassis(chassisType string, layout portLayout) []image.Rectangle {
	if chassisType != "7220 IXR-D2L" {
		return layout.portRects()
	}
	return d2lPortRects(layout)
}

func d2lPortRects(layout portLayout) []image.Rectangle {
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

	rects = append(rects, d2lRightSidePortRects()...)
	return rects
}

func d2lRightSidePortRects() []image.Rectangle {
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

func (l portLayout) portRects() []image.Rectangle {
	rects := make([]image.Rectangle, 0, len(l.topRowX)+len(l.botRowX))
	for _, x := range l.topRowX {
		rects = append(rects, image.Rect(x, l.topY, x+l.width, l.topY+l.height))
	}
	for _, x := range l.botRowX {
		rects = append(rects, image.Rect(x, l.botY, x+l.width, l.botY+l.height))
	}
	return rects
}

func drawPortOverlay(dst *image.RGBA, rect image.Rectangle, fillBase color.RGBA) {
	r := rect.Intersect(dst.Bounds())
	if r.Empty() {
		return
	}

	// Use non-premultiplied alpha for correct blending in draw.Over.
	fill := color.NRGBA{R: fillBase.R, G: fillBase.G, B: fillBase.B, A: 98}
	draw.Draw(dst, r, &image.Uniform{C: fill}, image.Point{}, draw.Over)
}

func drawPortLabel(dst *image.RGBA, rect image.Rectangle, label string) {
	r := rect.Intersect(dst.Bounds())
	if r.Empty() || strings.TrimSpace(label) == "" {
		return
	}

	face := labelFaceForRect(r, label)
	if face == nil {
		return
	}

	ascent := face.Metrics().Ascent.Round()
	descent := face.Metrics().Descent.Round()
	textHeight := ascent + descent

	measure := &font.Drawer{Face: face}
	textWidth := measure.MeasureString(label).Round()

	x := r.Min.X + (r.Dx()-textWidth)/2
	baseline := r.Min.Y + (r.Dy()-textHeight)/2 + ascent

	shadow := &image.Uniform{C: color.NRGBA{A: 220}}
	text := &image.Uniform{C: color.NRGBA{R: 245, G: 245, B: 245, A: 255}}

	d := &font.Drawer{Dst: dst, Face: face}
	d.Src = shadow
	d.Dot = fixed.P(x+1, baseline+1)
	d.DrawString(label)
	d.Src = text
	d.Dot = fixed.P(x, baseline)
	d.DrawString(label)
}

func labelFaceForRect(r image.Rectangle, label string) font.Face {
	minPx := 7
	maxPx := r.Dy() - 1
	if maxPx > 18 {
		maxPx = 18
	}
	if maxPx < minPx {
		maxPx = minPx
	}

	maxW := r.Dx() - 2
	maxH := r.Dy() - 1
	if maxW < 1 || maxH < 1 {
		return nil
	}

	for px := maxPx; px >= minPx; px-- {
		face := labelFace(px)
		if face == nil {
			continue
		}

		d := &font.Drawer{Face: face}
		textW := d.MeasureString(label).Round()
		textH := face.Metrics().Ascent.Round() + face.Metrics().Descent.Round()
		if textW <= maxW && textH <= maxH {
			return face
		}
	}

	return labelFace(minPx)
}

func labelFace(px int) font.Face {
	labelFontOnce.Do(func() {
		labelFont, labelFontErr = opentype.Parse(goregular.TTF)
	})
	if labelFontErr != nil || labelFont == nil {
		return nil
	}

	labelFaceMu.Lock()
	defer labelFaceMu.Unlock()
	if face, ok := labelFaces[px]; ok {
		return face
	}

	face, err := opentype.NewFace(labelFont, &opentype.FaceOptions{
		Size:    float64(px),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil
	}

	labelFaces[px] = face
	return face
}

func parseInterfaceIndex(ifName string) (int, bool) {
	parts := strings.Split(ifName, "/")
	if len(parts) >= 2 {
		if idx, err := strconv.Atoi(parts[1]); err == nil && idx > 0 {
			return idx, true
		}
	}

	nums := lastNumberPattern.FindAllString(ifName, -1)
	if len(nums) == 0 {
		return 0, false
	}

	idx, err := strconv.Atoi(nums[len(nums)-1])
	if err != nil || idx <= 0 {
		return 0, false
	}

	return idx, true
}

func stateOverlayColor(state string) (color.RGBA, bool) {
	s := strings.ToLower(strings.TrimSpace(state))
	s = strings.ReplaceAll(s, "_", "-")
	switch s {
	case "", "disable", "disabled", "admin-down", "admin-disable":
		return color.RGBA{}, false
	case "admin-up-oper-up", "up", "oper-up":
		return portUpColor, true
	case "admin-up-oper-down", "down", "oper-down", "lower-layer-down", "dormant",
		"not-present", "unknown", "testing", "enable", "enabled", "admin-up":
		return portAdminUpOperDownColor, true
	default:
		if strings.Contains(s, "disable") || strings.Contains(s, "admin-down") {
			return color.RGBA{}, false
		}
		if strings.Contains(s, "oper-up") || s == "up" {
			return portUpColor, true
		}
		if strings.Contains(s, "down") || strings.Contains(s, "admin-up") || strings.Contains(s, "enable") {
			return portAdminUpOperDownColor, true
		}
		return color.RGBA{}, false
	}
}

func printITermImage(w io.Writer, img image.Image, imageName string) error {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		return err
	}

	size := encoded.Len()
	name := base64.StdEncoding.EncodeToString([]byte(imageName + ".png"))
	data := base64.StdEncoding.EncodeToString(encoded.Bytes())

	_, err := fmt.Fprintf(w, "\x1b]1337;File=inline=1;size=%d;name=%s:%s\a\n", size, name, data)
	return err
}
