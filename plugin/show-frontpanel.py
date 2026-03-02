import argparse
import json
import os
import shutil
import subprocess
import sys

from srlinux.location import build_path
from srlinux.mgmt.cli import CliPlugin, CommandNodeWithArguments, RequiredPlugin
from srlinux.mgmt.cli.cli_loader import CliLoader
from srlinux.mgmt.cli.cli_output import CliOutput
from srlinux.mgmt.cli.cli_state import CliState
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

    def _front_port_states(self, state):
        states = {}

        try:
            ifaces_path = build_path("/interface[name=ethernet-*]")
            ifaces_data = state.server_data_store.get_data(ifaces_path, recursive=True)
            interfaces = ifaces_data.interface.items()
        except Exception:
            return states

        for interface in interfaces:
            if_name = getattr(interface, "name", "")
            if not if_name:
                continue

            admin_state = str(getattr(interface, "admin_state", "")).strip().lower()
            oper_state = str(getattr(interface, "oper_state", "")).strip().lower()

            if admin_state in ("disable", "disabled", "down"):
                states[if_name] = "admin-down"
                continue

            if oper_state in ("up", "oper-up"):
                states[if_name] = "admin-up-oper-up"
                continue

            states[if_name] = "admin-up-oper-down"

        return states

    def get_required_plugins(self):
        return [
            RequiredPlugin(module="srlinux", plugin="platform_reports"),
        ]

    def load(self, cli: CliLoader, arguments: argparse.Namespace):
        platform = cli.show_mode.root.get_command("platform")

        platform.add_command(
            syntax=self._syntax(),
            callback=self._print,
        )

    def _syntax(self):
        return Syntax(
            name="front-panel",
            short_help="Show image of front panel",
            help="Show image of front panel",
        )

    def _print(
        self,
        state: CliState,
        output: CliOutput,
        arguments: CommandNodeWithArguments,
        **_kwargs,
    ):
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

        env = os.environ.copy()
        env.setdefault("FRONTPANEL_PORT_LABELS", "1")
        term_size = shutil.get_terminal_size((0, 0))
        if term_size.columns > 0:
            env.setdefault("COLUMNS", str(term_size.columns))
        if term_size.lines > 0:
            env.setdefault("LINES", str(term_size.lines))

        front_port_states = self._front_port_states(state)
        if front_port_states:
            env["FRONTPANEL_PORT_STATES_JSON"] = json.dumps(
                front_port_states, separators=(",", ":")
            )

        proc = subprocess.run(
            cmd, stdout=sys.stdout, stderr=subprocess.PIPE, text=True, env=env
        )
        if proc.returncode != 0:
            output.print(
                f"Failed to render front panel image (protocol={protocol}): {proc.stderr.strip()}"
            )

        output.print(
            f"\n\n⚡ High resolution image: https://go.srlinux.dev/img-{chassis.type.replace(' ', '-').lower()}\n"
        )
