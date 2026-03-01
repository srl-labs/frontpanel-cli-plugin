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

	_ "github.com/HugoSmits86/nativewebp"

	"github.com/dolmen-go/kittyimg"
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
	portUpColor   = color.RGBA{R: 33, G: 201, B: 110, A: 255}
	portDownColor = color.RGBA{R: 227, G: 68, B: 68, A: 255}
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
		topRowX: []int{73, 96, 128, 151, 182, 206, 237, 261, 292, 316, 347, 371, 402, 425, 456, 480},
		botRowX: []int{73, 96, 128, 151, 182, 206, 237, 261, 292, 316, 347, 371, 402, 425, 456, 480},
		topY:    17,
		botY:    42,
		width:   22,
		height:  10,
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
	PrintWithProtocolAndPortStates(chassisType, protocol, nil)
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

func PrintWithProtocolAndPortStates(chassisType string, protocol string, portStates map[string]string) {
	printWithProtocol(chassisType, parseImageProtocol(protocol), portStates)
}

func printWithProtocol(chassisType string, protocol imageProtocol, portStates map[string]string) {
	if chassisDef, ok := chassisImages[chassisType]; ok {
		f := bytes.NewReader(chassisDef.Image)
		img, _, err := image.Decode(f)
		if err != nil {
			return
		}

		img = applyPortStateOverlay(chassisType, img, portStates)

		selected := resolveImageProtocol(protocol)
		if selected == imageProtocolITerm {
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

func resolveImageProtocol(protocol imageProtocol) imageProtocol {
	if protocol == imageProtocolKitty || protocol == imageProtocolITerm {
		return protocol
	}

	// Environment override for deployments where CLI flags are not easy to set.
	override := parseImageProtocol(os.Getenv("FRONTPANEL_IMAGE_PROTOCOL"))
	if override == imageProtocolKitty || override == imageProtocolITerm {
		return override
	}

	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "kitty") || strings.Contains(term, "xterm-ghostty") || os.Getenv("KITTY_WINDOW_ID") != "" {
		return imageProtocolKitty
	}

	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	if termProgram == "iterm.app" || termProgram == "vscode" || termProgram == "wezterm" || os.Getenv("VSCODE_PID") != "" {
		return imageProtocolITerm
	}

	return imageProtocolKitty
}

func applyPortStateOverlay(chassisType string, base image.Image, portStates map[string]string) image.Image {
	if len(portStates) == 0 {
		return base
	}

	layout, ok := chassisPortLayouts[chassisType]
	if !ok {
		return base
	}

	rects := layout.portRects()
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
		clr := portDownColor
		if isUpState(state) {
			clr = portUpColor
		}

		drawPortOverlay(dst, rect, clr)
	}

	return dst
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

func drawPortOverlay(dst *image.RGBA, rect image.Rectangle, border color.RGBA) {
	r := rect.Intersect(dst.Bounds())
	if r.Empty() {
		return
	}

	fill := color.RGBA{R: border.R, G: border.G, B: border.B, A: 44}
	draw.Draw(dst, r, &image.Uniform{C: fill}, image.Point{}, draw.Over)

	border.A = 230
	for x := r.Min.X; x < r.Max.X; x++ {
		dst.SetRGBA(x, r.Min.Y, border)
		dst.SetRGBA(x, r.Max.Y-1, border)
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		dst.SetRGBA(r.Min.X, y, border)
		dst.SetRGBA(r.Max.X-1, y, border)
	}

	ledHeight := 3
	if r.Dy() <= 6 {
		ledHeight = 1
	}
	led := image.Rect(r.Min.X+1, r.Max.Y-ledHeight-1, r.Max.X-1, r.Max.Y-1).Intersect(dst.Bounds())
	if !led.Empty() {
		draw.Draw(dst, led, &image.Uniform{C: border}, image.Point{}, draw.Over)
	}
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

func isUpState(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "up", "enable", "enabled", "oper-up":
		return true
	default:
		return false
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
