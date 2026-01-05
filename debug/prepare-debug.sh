#!/bin/bash

set -e

# ensure sockatsock is not present
rm -f /tmp/socatsock

## Relevant Environment Variables
# DEBUG_PORT -> Port that the dlv debugger listens on (default = 7000)
# ENTRY_ID -> The cpm-filter ipv4-filter entry number to use for the dlv debugger allow rule

# install socat
echo "Installing socat"
ip netns exec srbase-mgmt bash -c "DEBIAN_FRONTEND=noninteractive ; apt-get update && sudo apt-get install -y socat" &> /debug/apt.log

# start socat based tcp redirection from srbase-mgmt (owner of the mgmt-ip) to srbase namespace (default location of app-manager started apps)
DPORT="${DEBUG_PORT:-7000}"
echo "Running socat on port ${DPORT}"
# Note: tcp6-listen will listen for v4 and v6 implicitly
ip netns exec srbase-mgmt socat -d -d -d -lf /tmp/mgmt-socat.log TCP6-LISTEN:${DPORT},reuseaddr,fork UNIX-CLIENT:/tmp/socatsock &
ip netns exec srbase socat -d -d -d -lf /tmp/base-socat.log UNIX-LISTEN:/tmp/socatsock,reuseaddr,fork TCP:127.0.0.1:${DPORT} &


# Add a CPM-Filter rule to allow the dlv debugger to connect
EID="${ENTRY_ID:-103}"
echo "Deploying cpm-filter ipv(4,6) entry ${EID}"
sr_cli -ec <<EOF
set / acl cpm-filter ipv4-filter entry ${EID} description "dlv debugger"
set / acl cpm-filter ipv4-filter entry ${EID} match protocol tcp
set / acl cpm-filter ipv4-filter entry ${EID} match destination-port value ${DPORT}
set / acl cpm-filter ipv4-filter entry ${EID} action accept
set / acl cpm-filter ipv6-filter entry ${EID} description "dlv debugger"
set / acl cpm-filter ipv6-filter entry ${EID} match next-header tcp
set / acl cpm-filter ipv6-filter entry ${EID} match destination-port value ${DPORT}
set / acl cpm-filter ipv6-filter entry ${EID} action accept
EOF

echo "Done"