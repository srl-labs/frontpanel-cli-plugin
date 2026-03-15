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
| 7730 SXR-1x-44S |

## Supported terminals

Depending on your terminal capabilities, the plugin will use either kitty graphics protocol or iTerm inline images (OSC 1337) to render the front panel image. If your terminal supports neither protocol, the plugin will print a URL to a high-resolution image of the front panel instead.

| Terminal | Graphics protocol | Notes |
| --- | --- | --- |
| Kitty | Kitty graphics protocol | |
| iTerm2 | iTerm inline images (OSC 1337) | |
| VS Code Integrated Terminal | iTerm inline images (OSC 1337) | Requires `"terminal.integrated.enableImages": true` setting. On MacOS with narrow terminal windows images may appear blurry. |
| Ghostty | Kitty graphics protocol | |
| WezTerm | Kitty graphics protocol | |

Terminals with no image support: MacOS Terminal, PuTTY.

**Note**: VS Code Integrated Terminal does not support kitty graphics protocol, but supports iTerm inline images when `"terminal.integrated.enableImages"` setting is enabled.

![vscode-setting](https://gitlab.com/rdodin/pics/-/wikis/uploads/b1198e1d659adee7e5fb3f4e3cffac79/image.png)
