import os
import subprocess
import sys

from srlinux.location import build_path
from srlinux.mgmt.cli import CliPlugin
from srlinux.syntax import Syntax


class Plugin(CliPlugin):
    def _image_protocol(self):
        # Manual override when needed.
        override = os.environ.get("FRONTPANEL_IMAGE_PROTOCOL", "").strip().lower()
        if override in ("kitty", "iterm", "auto"):
            return override

        term = os.environ.get("TERM", "").lower()
        if (
            "kitty" in term
            or "xterm-ghostty" in term  # ghostty supports kitty protocol
            or os.environ.get("KITTY_WINDOW_ID")
        ):
            return "kitty"

        # Over SSH TERM is often xterm-256color; default to iTerm protocol.
        return "iterm"

    def get_required_plugins(self):
        return [
            RequiredPlugin(module="srlinux", plugin="platform_reports"),
        ]

    def load(self, cli, **_kwargs):
        platform = cli.show_mode.root.get_command("platform")

        frontpanel = platform.add_command(
            syntax=self._syntax(),
            callback=self._print,
        )

    def _syntax(self):
        return Syntax(
            name="front-panel",
            short_help="Show image of front panel",
            help="Show image of front panel",
        )

    def _print(self, state, arguments, output, **_kwargs):
        chassis_path = build_path("/platform/chassis/type")
        chassis_server_data = state.server_data_store.get_data(
            chassis_path, recursive=False
        )
        chassis = chassis_server_data.platform.get().chassis.get()

        protocol = self._image_protocol()
        cmd = [
            "/usr/local/bin/frontpanel",
            "-image",
            chassis.type,
            "-image-protocol",
            protocol,
        ]
        proc = subprocess.run(cmd, stdout=sys.stdout, stderr=subprocess.PIPE, text=True)
        if proc.returncode != 0:
            output.print(
                f"Failed to render front panel image (protocol={protocol}): {proc.stderr.strip()}"
            )

        sys.stdout.flush()

        front_panel_path = build_path("/platform/front-panel")
        front_panel_server_data = state.server_data_store.get_data(
            front_panel_path, recursive=False
        )
        front_panel = front_panel_server_data.platform.get().front_panel.get()
        output.print(f"\n\nHigh resolution image: {front_panel.url}\n")
