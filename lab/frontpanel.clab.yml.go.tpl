name: frontpanel
prefix: ""

topology:
  nodes:
    frontpanel:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:25.10
      exec:
        - touch /tmp/.ndk-dev-mode
        {{- if ne (env.Getenv "NDK_DEBUG") "" }}
        - /debug/prepare-debug.sh
        {{- end }}
      binds:
        - ../build:/tmp/build # mount app binary
        - ../plugin:/tmp/plugin # mount show plugin script
        - ../frontpanel.yml:/tmp/frontpanel.yml # agent config file to appmgr directory
        - ../yang:/opt/frontpanel/yang # yang modules
        - ../logs/srl:/var/log/srlinux # expose srlinux logs
        - ../logs/frontpanel/:/var/log/frontpanel # expose greeter log file
        {{- if ne (env.Getenv "NDK_DEBUG") "" }}
        - ../debug/:/debug/
        {{- end }}