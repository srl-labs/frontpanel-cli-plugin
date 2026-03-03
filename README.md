# Show SR Linux front panel in your terminal

[![Discord][discord-svg]][discord-url] [![Codespaces][codespaces-svg]][codespaces-url]  
![w212][w212][Learn more](https://containerlab.dev/manual/codespaces)

[discord-svg]: https://gitlab.com/rdodin/pics/-/wikis/uploads/b822984bc95d77ba92d50109c66c7afe/join-discord-btn.svg
[discord-url]: https://discord.gg/tZvgjQ6PZf
[codespaces-svg]: https://gitlab.com/rdodin/pics/-/wikis/uploads/80546a8c7cda8bb14aa799d26f55bd83/run-codespaces-btn.svg
[codespaces-url]: https://codespaces.new/srl-labs/frontpanel-cli-plugin?quickstart=1&devcontainer_path=.devcontainer%2Fdevcontainer.json
[w212]: https://gitlab.com/rdodin/pics/-/wikis/uploads/718a32dfa2b375cb07bcac50ae32964a/w212h1.svg

This repository provides a simple SR Linux CLI extension that shows a terminal-rendered image of the device front panel using terminal image protocols ([kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/) and iTerm inline images / OSC 1337).

![A screenshot displaying the CLI plugin in action - an image of the front panel is embedded as part of the CLI output](screenshot.png)

The repository includes a Containerlab topology so you can try the plugin end-to-end.

## Requirements

This repository is best tried out on Linux - either on a VM or bare metal works fine.

You will need to have a working installation of [Golang](https://go.dev/doc/install) and [Docker](https://docs.docker.com/engine/install/) to build this application.  
The rest of the tooling used during build and packaging are pulled from public container repository images.

To try out the application, you will also need to have [Containerlab](https://containerlab.dev/install/) installed (0.68.0 or newer).  
Additionally, _you must use a terminal application that supports either kitty graphics protocol or iTerm inline images (OSC 1337) to be able to see the embedded images in the CLI output._

A short list of terminals with kitty graphics protocol support:

**Mac:**

- Ghostty
- KiTTY
- iTerm2

**Linux:**

- Ghostty
- KiTTY
- Konsole

**Cross-platform:**

- WezTerm

**Note**: VS Code Integrated Terminal does not support kitty graphics protocol, but supports iTerm inline images when `"terminal.integrated.enableImages"` setting is enabled.

## Quick start

The helper script `run.sh` can build the binary and deploy the lab:

- `./run.sh deploy-all`

This will:

- format source files
- build `./build/frontpanel`
- deploy the topology from `frontpanel.clab.yml`

After deploy, run `show platform front-panel` on the SR Linux node.

## Build and package

- Build only: `./run.sh build-app`
- Package as `.deb`: `./run.sh package`

Package artifacts are written to `./build/` (for example `frontpanel_*.deb`).

## Usage

The front panel plugin is made of 2 components:

- The binary `frontpanel`  
This binary renders front panel images and prints them to the terminal.
You can run it directly with `frontpanel -image "7220 IXR-D2L"`.

The image protocol can be selected with:

- `-image-protocol auto|kitty|iterm` (default `auto`)
- `FRONTPANEL_IMAGE_PROTOCOL=kitty|iterm` (env override)
- `-port-labels` or `FRONTPANEL_PORT_LABELS=1` to overlay port numbers (`1/1`, `1/2`, ...)
- `-port-states-json '{"ethernet-1/1":"admin-up-oper-up","ethernet-1/2":"admin-up-oper-down","ethernet-1/3":"admin-down"}'` or `FRONTPANEL_PORT_STATES_JSON` to color front ports (bright green for admin+oper up, orange for admin up+oper down, no color for admin down)

Examples:

- `frontpanel -image "7220 IXR-D2L" -image-protocol iterm`
- `FRONTPANEL_IMAGE_PROTOCOL=iterm frontpanel -image "7220 IXR-D2L"`

- The Python CLI plugin `show-frontpanel.py`  
This plugin adds `show platform front-panel`, resolves the chassis type from SR Linux state, calls `frontpanel -image ...`, and prints a high-resolution image URL.

The plugin auto-selects `kitty` when `TERM` indicates kitty or ghostty; otherwise it uses `iterm` (OSC 1337), which works better over SSH in terminals like VS Code.
It also reads front panel interface state (`/interface[name=ethernet-*]`) and forwards it to the renderer using admin/oper-aware states.
You can override this with `FRONTPANEL_IMAGE_PROTOCOL=kitty|iterm|auto`.
Port labels are enabled by default in the plugin (`FRONTPANEL_PORT_LABELS=1` unless explicitly overridden).

## Cleanup

To destroy the test lab, run `./run.sh destroy-lab`.
