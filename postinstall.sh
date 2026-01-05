#!/bin/bash
set +e

source /etc/profile.d/sr_app_env.sh

echo "Reload appmgr"
sr_cli -e "/ tools system app-management application app_mgr reload" > /dev/null 2>&1 || true
sleep 1