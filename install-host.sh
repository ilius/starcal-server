#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
    echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
    exit 1
fi

echo "Installing mongodb-org 3.2" ; ssh root@$STARCAL_HOST apt-get install 'mongodb-org=3.2.*'
echo "Installing daemontools" ; ssh root@$STARCAL_HOST apt-get install daemontools
echo "Copying init.d script" ; scp ./init.d/starcal root@$STARCAL_HOST:/etc/init.d/starcal
echo "Copying systemd service file" ; scp ./systemd/starcal.service root@$STARCAL_HOST:/lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; ssh root@$STARCAL_HOST systemctl enable starcal
echo "Reloading systemctl" ; ssh root@$STARCAL_HOST systemctl daemon-reload

echo
echo "Running ./deploy.sh"
./deploy.sh
