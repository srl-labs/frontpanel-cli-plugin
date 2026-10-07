# Show SR Linux front panel in your terminal

[![Discord][discord-svg]][discord-url] [![Codespaces][codespaces-svg]][codespaces-url]  
![w212][w212][Learn more](https://containerlab.dev/manual/codespaces)

[discord-svg]: https://gitlab.com/rdodin/pics/-/wikis/uploads/b822984bc95d77ba92d50109c66c7afe/join-discord-btn.svg
[discord-url]: https://discord.gg/tZvgjQ6PZf
[codespaces-svg]: https://gitlab.com/rdodin/pics/-/wikis/uploads/80546a8c7cda8bb14aa799d26f55bd83/run-codespaces-btn.svg
[codespaces-url]: https://codespaces.new/srl-labs/frontpanel-cli-plugin?quickstart=1&devcontainer_path=.devcontainer%2Fdevcontainer.json
[w212]: https://gitlab.com/rdodin/pics/-/wikis/uploads/718a32dfa2b375cb07bcac50ae32964a/w212h1.svg

This repository provides an [SR Linux CLI plugin](https://learn.srlinux.dev/cli/plugins/) that shows a terminal-rendered image of the device front panel using terminal image protocols ([kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/) and iTerm inline images / OSC 1337) with port states overlay.

![A screenshot displaying the CLI plugin in action - an image of the front panel is embedded as part of the CLI output](https://gitlab.com/rdodin/pics/-/wikis/uploads/1bb8b3236a7fa7954f0af2ba388496b1/image.png)

## Demo

Click on the preview image to view a demo on YouTube.

[![Watch the video](https://img.youtube.com/vi/-flkq5MpBOA/hqdefault.jpg)](https://www.youtube.com/watch?v=-flkq5MpBOA)

Leave your comments on [LinkedIn](https://www.linkedin.com/feed/update/urn:li:activity:7435277879200014337).

## Quick start

> If you want to see how the plugin works without having to build it yourself, you can try it out in a GitHub Codespace with the "Open in Codespaces" button at the top of this README.

To run the plugin locally with the provided Containerlab topology, ensure you have Go 1.24+ installed and run the below command to build the binary and deploy the lab:

```bash
./run.sh deploy-all
```

The lab contains two SR Linux nodes (7220 IXR-D2L and 7220 IXR-D3L) with different front panels, so you can try out the plugin on both by running `show platform front-panel` on each node.

SSH into one of the nodes and run `show platform front-panel` to see the front panel image rendered directly in your terminal. The plugin will auto-detect your terminal capabilities and use the best available image protocol.

On top of the frontpanel image, you will see the port labels (e.g. `1/1`, `1/2`, ...) and color-coded port states:

- **green** for admin up AND oper up
- **orange** for admin up AND oper down
- no color for admin down

Port states/color are based on the actual interface state in SR Linux.

## Build and install on an SR Linux node

Use these steps to test a branch on a physical or virtual SR Linux node before
creating a release. Go 1.24 or newer is required on the build host.

First, determine the node architecture:

```bash
BOX=admin@<node-address>
ssh "$BOX" uname -m
```

Use `amd64` for an `x86_64` node or `arm64` for an `aarch64` node, then build a
static Linux binary:

```bash
export GOARCH=amd64 # or arm64
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)

mkdir -p build
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" \
  go build -trimpath \
  -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT" \
  -o build/frontpanel .
```

Copy the binary and CLI plugin to the node:

```bash
scp build/frontpanel plugin/show-frontpanel.py "$BOX":/tmp/
ssh -t "$BOX" \
  'sudo install -m 0755 /tmp/frontpanel /usr/local/bin/frontpanel &&
   sudo install -o srlinux -g srlinux -m 0644 /tmp/show-frontpanel.py \
     /etc/opt/srlinux/cli/plugins/show-frontpanel.py'
```

Start a new SR Linux CLI session so the plugin is loaded, then verify it:

```text
show platform front-panel
```

Confirm that the displayed platform matches `/platform/chassis/type` and that
each `1/N` overlay corresponds to `ethernet-1/N` before releasing.

## Supported platforms

Added platforms are listed below. Request new platforms by opening an issue.

| Platform |
| --- |
| 7215 IXS-A1 |
| 7220 IXR-D1 |
| 7220 IXR-D2 |
| 7220 IXR-D2L |
| 7220 IXR-D3 |
| 7220 IXR-D3L |
| 7220 IXR-D5 |
| 7250 IXR-X1B |
| 7250 IXR-X3B |
| 7250 IXR-X4 |
| 7250 IXR-X4-OSFP |
| 7730 SXR-1x-44S |

## Supported terminals

Depending on your terminal capabilities, the plugin will use either kitty graphics protocol or iTerm inline images (OSC 1337) to render the front panel image.

| Terminal | Graphics protocol | Notes |
| --- | --- | --- |
| Kitty | Kitty graphics protocol | |
| iTerm2 | iTerm inline images (OSC 1337) | |
| VS Code Integrated Terminal | Kitty graphics protocol (>=1.110.1) and iTerm images (<1.110.1) | Requires `"terminal.integrated.enableImages": true` setting. On MacOS with narrow terminal windows images may appear blurry. |
| Ghostty | Kitty graphics protocol | |
| WezTerm | Kitty graphics protocol | |

Terminals with no image support: MacOS Terminal, PuTTY.

**Note**: VS Code Integrated Terminal [started supporting kitty graphics protocol in version 1.110](https://code.visualstudio.com/updates/v1_110#_terminal), prior to that it supported iTerm inline images when `"terminal.integrated.enableImages"` setting is enabled.

![vscode-setting](https://gitlab.com/rdodin/pics/-/wikis/uploads/b1198e1d659adee7e5fb3f4e3cffac79/image.png)
