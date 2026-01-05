import subprocess

from srlinux.mgmt.cli import CliPlugin
from srlinux.syntax import Syntax
from srlinux.location import build_path

class Plugin(CliPlugin):
    def get_required_plugins(self):
        return [
            RequiredPlugin(module='srlinux', plugin='platform_reports'),
        ]

    def load(self, cli, **_kwargs):
        platform = cli.show_mode.root.get_command('platform')

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
        chassis_path = build_path('/platform/chassis/type')
        chassis_server_data = state.server_data_store.get_data(chassis_path, recursive=False)
        chassis = chassis_server_data.platform.get().chassis.get()

        output.print(subprocess.run(['/usr/local/bin/frontpanel', '-image', chassis.type], stdout=subprocess.PIPE).stdout.decode('utf-8'))        
        
        front_panel_path = build_path('/platform/front-panel')
        front_panel_server_data = state.server_data_store.get_data(front_panel_path, recursive=False)
        front_panel = front_panel_server_data.platform.get().front_panel.get()
        output.print(f"\n\nHigh resolution image: {front_panel.url}\n")

