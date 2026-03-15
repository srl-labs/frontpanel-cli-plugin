package frontpanel

import (
	"bytes"
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
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type platformDef struct {
	image     []byte
	layout    portLayout
	portRects func(portLayout) []image.Rectangle
}

var platformRegistry = map[string]platformDef{}

func registerAllPlatforms() {
	registerIXS_A1()
	registerIXR_D1()
	registerIXR_D2()
	registerIXR_D2L()
	registerIXR_D3()
	registerIXR_D3L()
	registerIXR_D5()
	registerSXR_1X_44S()
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
	registerAllPlatforms()

	if def, ok := platformRegistry[chassisType]; ok {
		f := bytes.NewReader(def.image)
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

		_ = printKittyImage(os.Stdout, img)

	} else {
		fmt.Println("not supported")
	}
}

func printKittyImage(w io.Writer, img image.Image) error {
	cols, rows := terminalSize(w)
	if cols <= 0 || rows <= 0 {
		return kittyimg.Fprintln(w, img)
	}

	var raw bytes.Buffer
	if err := kittyimg.Fprintln(&raw, img); err != nil {
		return kittyimg.Fprintln(w, img)
	}

	data := raw.Bytes()
	semi := bytes.IndexByte(data, ';')
	if semi <= 0 {
		return kittyimg.Fprintln(w, img)
	}

	targetCols, _ := fitImageToCells(img, cols, rows)
	if targetCols <= 0 {
		return kittyimg.Fprintln(w, img)
	}

	header := append([]byte{}, data[:semi]...)
	if len(header) > 0 {
		header = append(header, ',')
		header = append(header, []byte(fmt.Sprintf("c=%d", targetCols))...)
	}
	header = append(header, data[semi:]...)
	_, err := w.Write(header)
	return err
}

func fitImageToCells(img image.Image, maxCols int, maxRows int) (int, int) {
	if maxCols <= 0 || maxRows <= 0 {
		return 0, 0
	}

	b := img.Bounds()
	imgW := b.Dx()
	imgH := b.Dy()
	if imgW <= 0 || imgH <= 0 {
		return 0, 0
	}

	targetCols := maxCols
	targetRows := (imgH*targetCols + imgW - 1) / imgW
	if targetRows > maxRows {
		targetRows = maxRows
		targetCols = (imgW*targetRows + imgH - 1) / imgH
	}

	if targetCols > maxCols {
		targetCols = maxCols
	}
	if targetRows > maxRows {
		targetRows = maxRows
	}
	if targetCols < 1 {
		targetCols = 1
	}
	if targetRows < 1 {
		targetRows = 1
	}

	return targetCols, targetRows
}

func fitImageToPixels(img image.Image, maxW int, maxH int) (int, int) {
	if maxW <= 0 || maxH <= 0 {
		return 0, 0
	}

	b := img.Bounds()
	imgW := b.Dx()
	imgH := b.Dy()
	if imgW <= 0 || imgH <= 0 {
		return 0, 0
	}

	targetW := maxW
	targetH := (imgH*targetW + imgW - 1) / imgW
	if targetH > maxH {
		targetH = maxH
		targetW = (imgW*targetH + imgH - 1) / imgH
	}

	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	return targetW, targetH
}

func terminalSize(w io.Writer) (int, int) {
	if cols, rows := terminalSizeFromWriter(w); cols > 0 && rows > 0 {
		return cols, rows
	}

	// Try common stdio streams in case output itself is not the terminal stream.
	for _, f := range []*os.File{os.Stdout, os.Stderr, os.Stdin} {
		if cols, rows := terminalSizeFromFile(f); cols > 0 && rows > 0 {
			return cols, rows
		}
	}

	if cols, rows := terminalSizeFromTty(); cols > 0 && rows > 0 {
		return cols, rows
	}

	if cols, rows := terminalSizeFromEnv(); cols > 0 && rows > 0 {
		return cols, rows
	}

	// Some execution environments (for example, CLI plugins or non-interactive runs)
	// do not expose terminal geometry; use a conservative default so images never
	// render at their intrinsic pixel size.
	return 80, 24
}

func terminalSizeFromWriter(w io.Writer) (int, int) {
	stdout, ok := w.(*os.File)
	if !ok {
		return 0, 0
	}
	return terminalSizeFromFile(stdout)
}

func terminalSizeFromFile(f *os.File) (int, int) {
	if f == nil {
		return 0, 0
	}

	cols, rows, err := term.GetSize(int(f.Fd()))
	if err != nil || cols <= 0 || rows <= 0 {
		return 0, 0
	}

	return cols, rows
}

func terminalSizeFromEnv() (int, int) {
	cols, _ := strconv.Atoi(os.Getenv("COLUMNS"))
	rows, _ := strconv.Atoi(os.Getenv("LINES"))
	return cols, rows
}

func terminalSizeFromTty() (int, int) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0
	}
	defer tty.Close()

	cols, rows, err := term.GetSize(int(tty.Fd()))
	if err != nil || cols <= 0 || rows <= 0 {
		return 0, 0
	}

	return cols, rows
}

func terminalPixelSize(w io.Writer) (int, int) {
	if pxW, pxH := terminalPixelSizeFromWriter(w); pxW > 0 && pxH > 0 {
		return pxW, pxH
	}

	for _, f := range []*os.File{os.Stdout, os.Stderr, os.Stdin} {
		if pxW, pxH := terminalPixelSizeFromFile(f); pxW > 0 && pxH > 0 {
			return pxW, pxH
		}
	}

	if pxW, pxH := terminalPixelSizeFromTty(); pxW > 0 && pxH > 0 {
		return pxW, pxH
	}

	return 0, 0
}

func terminalPixelSizeFromWriter(w io.Writer) (int, int) {
	stdout, ok := w.(*os.File)
	if !ok {
		return 0, 0
	}
	return terminalPixelSizeFromFile(stdout)
}

func terminalPixelSizeFromFile(f *os.File) (int, int) {
	if f == nil {
		return 0, 0
	}

	ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws == nil || ws.Xpixel <= 0 || ws.Ypixel <= 0 {
		return 0, 0
	}

	return int(ws.Xpixel), int(ws.Ypixel)
}

func terminalPixelSizeFromTty() (int, int) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0
	}
	defer tty.Close()

	return terminalPixelSizeFromFile(tty)
}

func itermTargetPixelBounds(w io.Writer, termCols int, termRows int, targetCols int, targetRows int) (int, int) {
	if targetCols <= 0 || targetRows <= 0 {
		return 0, 0
	}

	cellW := 8
	cellH := 16
	pxW, pxH := terminalPixelSize(w)
	if pxW > 0 && termCols > 0 {
		if cw := pxW / termCols; cw > 0 {
			cellW = cw
		}
	}
	if pxH > 0 && termRows > 0 {
		if ch := pxH / termRows; ch > 0 {
			cellH = ch
		}
	}

	return targetCols * cellW, targetRows * cellH
}

func scaleImageForITerm(w io.Writer, img image.Image, termCols int, termRows int, targetCols int, targetRows int) image.Image {
	maxW, maxH := itermTargetPixelBounds(w, termCols, termRows, targetCols, targetRows)
	targetW, targetH := fitImageToPixels(img, maxW, maxH)
	if targetW <= 0 || targetH <= 0 {
		return img
	}

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	if targetW >= srcW && targetH >= srcH {
		return img
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, srcBounds, xdraw.Src, nil)
	return dst
}

func applyPortStateOverlay(chassisType string, base image.Image, portStates map[string]string) image.Image {
	if len(portStates) == 0 {
		return base
	}

	def, ok := platformRegistry[chassisType]
	if !ok || def.portRects == nil {
		return base
	}

	rects := def.portRects(def.layout)
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
	def, ok := platformRegistry[chassisType]
	if !ok || def.portRects == nil {
		return base
	}

	rects := def.portRects(def.layout)
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
	cols, rows := terminalSize(w)
	targetCols, targetRows := fitImageToCells(img, cols, rows)
	if targetCols <= 0 || targetRows <= 0 {
		targetCols = cols
		targetRows = rows
	}
	if targetCols <= 0 {
		targetCols = 80
	}
	if targetRows <= 0 {
		targetRows = 24
	}

	itermImg := scaleImageForITerm(w, img, cols, rows, targetCols, targetRows)

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, itermImg); err != nil {
		return err
	}

	size := encoded.Len()
	name := base64.StdEncoding.EncodeToString([]byte(imageName + ".png"))
	data := base64.StdEncoding.EncodeToString(encoded.Bytes())

	_, err := fmt.Fprintf(
		w,
		"\x1b]1337;File=inline=1;size=%d;name=%s;width=%d;height=%d;preserveAspectRatio=1:%s\a\n",
		size,
		name,
		targetCols,
		targetRows,
		data,
	)
	return err
}
