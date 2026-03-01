package frontpanel

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
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
	printWithProtocol(chassisType, parseImageProtocol(protocol))
}

func printWithProtocol(chassisType string, protocol imageProtocol) {
	if chassisDef, ok := chassisImages[chassisType]; ok {
		f := bytes.NewReader(chassisDef.Image)
		img, _, err := image.Decode(f)
		if err != nil {
			return
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
