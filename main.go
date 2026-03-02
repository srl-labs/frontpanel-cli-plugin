package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/srl-labs/ndk-frontpanel/frontpanel"
)

var (
	version = "0.0.0"
	commit  = ""
)

func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")
	imageFlag := flag.String("image", "", "print the front panel image and exit")
	imageProtocolFlag := flag.String("image-protocol", "auto", "image protocol: auto|kitty|iterm")
	portLabelsFlag := flag.Bool("port-labels", false, "overlay port labels (1/1, 1/2, ...)")
	portStatesJSONFlag := flag.String("port-states-json", "",
		"JSON object of interface state by name, e.g. {\"ethernet-1/1\":\"up\"}")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version + "-" + commit)
		os.Exit(0)
	}

	if *imageFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: --image flag is required")
		os.Exit(1)
	}

	portStatesJSON := *portStatesJSONFlag
	if portStatesJSON == "" {
		portStatesJSON = os.Getenv("FRONTPANEL_PORT_STATES_JSON")
	}

	portLabels := *portLabelsFlag
	if !portLabels {
		portLabels = frontpanel.ParsePortLabelsValue(os.Getenv("FRONTPANEL_PORT_LABELS"))
	}

	frontpanel.PrintWithProtocolAndPortStatesAndLabels(
		*imageFlag,
		*imageProtocolFlag,
		frontpanel.ParsePortStatesJSON(portStatesJSON),
		portLabels,
	)
	os.Exit(0)
}
