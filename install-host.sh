#!/bin/bash
if [ -z $starcal_host ] ; then
    echo 'Set (and export) environment varibale `starcal_host` before running this script'
    exit 1
fi

echo "Installing mongodb-org 3.2" ; ssh root@$starcal_host apt-get install 'mongodb-org=3.2.*'
echo "Copying init.d script" ; scp ./init.d/starcal root@$starcal_host:/etc/init.d/starcal
echo "Copying systemd service file" ; scp ./systemd/starcal.service root@$starcal_host:/lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; ssh root@$starcal_host systemctl enable starcal
echo "Reloading systemctl" ; ssh root@$starcal_host systemctl daemon-reload

echo
echo "Running ./deploy.sh"
./deploy.sh
