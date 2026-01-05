front-panel:
  path: /usr/local/bin
  launch-command: {{if ne (env.Getenv "NDK_DEBUG") "" }}{{ "/debug/dlv --listen=:7000"}}{{ if ne (env.Getenv "NDK_DEBUG") "" }} {{ "--continue --accept-multiclient" }}{{ end }} {{ "--headless=true --log=true --api-version=2 exec"}} {{ end }}frontpanel
  version-command: frontpanel --version
  failure-action: wait=10
  config-delivery-format: json
  yang-modules:
    names:
      - frontpanel
    source-directories:
      - /opt/frontpanel/yang
