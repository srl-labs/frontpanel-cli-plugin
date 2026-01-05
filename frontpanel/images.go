package frontpanel

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"os"

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

func Print(chassisType string) {
	if chassisDef, ok := chassisImages[chassisType]; ok {
		f := bytes.NewReader(chassisDef.Image)
		img, _, err := image.Decode(f)
		if err != nil {
			return
		}

		kittyimg.Fprintln(os.Stdout, img)

	} else {
		fmt.Println("not supported")
	}
}
