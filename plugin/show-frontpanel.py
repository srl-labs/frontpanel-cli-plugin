import os
import subprocess
import sys
import base64
import tempfile

from srlinux.location import build_path
from srlinux.mgmt.cli import CliPlugin
from srlinux.syntax import Syntax

FONT_FILE = "/usr/share/fonts/frontpanel/Arial.ttf"
SVG_DIR = "/etc/opt/srlinux/frontpanel/images"

CHASSIS_SVG = {
    "7220 IXR-D2L": "d2l.svg",
    "7220 IXR-D3L": "d3l.svg",
    "7220 IXR-D5": "d5.svg",
}

def get_chassis_svg(chassis_type):
    filename = CHASSIS_SVG.get(chassis_type)
    if not filename:
        return None

    path = os.path.join(SVG_DIR, filename)
    if os.path.isfile(path):
        return path

    return None

def build_port_css(port_states):
    up = []
    unhealthy = []
    for if_name, state in port_states.items():
        if state == "admin-up-oper-up":
            up.append(if_name)
        elif state == "admin-up-oper-down":
            unhealthy.append(if_name)

    rules = []
    if up:
        selector = ", ".join(f'[id="fp-{n}"]' for n in sorted(up))
        rules.append(f"{selector} {{ fill: #21c96e; }}") # GREEN.
    if unhealthy:
        selector = ", ".join(f'[id="fp-{n}"]' for n in sorted(unhealthy))
        rules.append(f"{selector} {{ fill: #f58220; }}") # ORANGE.

    return "\n".join(rules)


def get_term_size():
    try:
        import fcntl
        import struct
        import termios

        buf = fcntl.ioctl(
            sys.stdout.fileno(),
            termios.TIOCGWINSZ,
            b"\x00" * 8,
        )
        _, cols, xpixel, _ = struct.unpack("HHHH", buf)
        return cols, xpixel or None
    except Exception:
        pass

    return 80, None


def svg_to_png(svg_path, width=None, css=None):
    cmd = ["resvg"]

    if width:
        cmd += ["-w", str(width)]

    if os.path.isfile(FONT_FILE):
        cmd += ["--use-font-file", FONT_FILE]

    css_file = None
    try:
        if css:
            css_file = tempfile.NamedTemporaryFile(
                mode="w", suffix=".css", delete=False
            )
            css_file.write(css)
            css_file.close()
            cmd += ["--stylesheet", css_file.name]

        cmd += [svg_path, "-c"]

        proc = subprocess.run(cmd, capture_output=True)
        if proc.returncode != 0:
            raise RuntimeError(
                f"resvg failed: {proc.stderr.decode(errors='replace')}"
            )
        return proc.stdout
    finally:
        if css_file:
            os.unlink(css_file.name)


def print_iterm_image(png_bytes, cols):
    data = base64.b64encode(png_bytes).decode("ascii")
    size = len(png_bytes)
    sys.stdout.write(
        f"\033]1337;File=inline=1;size={size};width={cols}:{data}\a\n"
    )
    sys.stdout.flush()


def print_kitty_image(png_bytes, cols):
    data = base64.b64encode(png_bytes).decode("ascii")
    CHUNK = 4096
    first = True
    while data:
        chunk = data[:CHUNK]
        data = data[CHUNK:]
        more = 1 if data else 0
        if first:
            col_param = f",c={cols}" if cols else ""
            sys.stdout.write(f"\033_Gf=100,a=T{col_param},m={more};{chunk}\033\\")
            first = False
        else:
            sys.stdout.write(f"\033_Gm={more};{chunk}\033\\")
    sys.stdout.write("\n")
    sys.stdout.flush()


def render_svg(chassis_type, protocol, port_states):
    svg_path = get_chassis_svg(chassis_type)
    if not svg_path:
        return f"No frontpanel image for {chassis_type}"

    try:
        css = build_port_css(port_states) if port_states else None

        cols, pixel_width = get_term_size()
        png_bytes = svg_to_png(svg_path, width=pixel_width, css=css)

        if protocol == "kitty":
            print_kitty_image(png_bytes, cols)
        else:
            print_iterm_image(png_bytes, cols)

        return None
    except Exception as e:
        return str(e)


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
        port_states = self._front_port_states(state)

        err = render_svg(chassis.type, protocol, port_states)
        if err:
            output.print(f"Failed to render front panel image: {err}")
