#!/bin/bash
. ./funcs.sh
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	exit 1
fi

echo "Installing mongodb-org 3.2" ; scal_ssh apt-get install 'mongodb-org=3.2.*'
echo "Installing daemontools" ; scal_ssh apt-get install daemontools
echo "Installing bsdiff and libarchive-zip-perl" ; scal_ssh apt-get install bsdiff libarchive-zip-perl
echo "Copying init.d script" ; scal_scp_to_host ./init.d/starcal /etc/init.d/starcal
echo "Copying systemd service file" ; scal_scp_to_host ./systemd/starcal.service /lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; scal_ssh systemctl enable starcal
echo "Reloading systemctl" ; scal_ssh systemctl daemon-reload

echo
echo "Running ./deploy.sh"
./deploy.sh
