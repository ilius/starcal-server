#!/bin/bash
. ./funcs.sh
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	exit 1
fi

# Install package 'dpkg' to provide command 'start-stop-daemon'
echo "Installing dpkg" ; scal_ssh dnf install dpkg

echo "Installing mongodb-server 3.2.8" ; scal_ssh dnf install mongodb-server-3.2.8

# http://djbware.csi.hu/daemontools.html
echo "Downloading and installing daemontools"
scal_ssh dnf install 'http://djbware.csi.hu/RPMS/daemontools-0.76-112memphis.i386.rpm'

echo "Installing bsdiff and perl-Archive-Zip" ; scal_ssh dnf install bsdiff perl-Archive-Zip
echo "Copying init.d script" ; scal_scp_to_host ./init.d/starcal /etc/init.d/starcal
echo "Copying systemd service file" ; scal_scp_to_host ./systemd/starcal.service /lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; scal_ssh systemctl enable starcal
echo "Reloading systemctl" ; scal_ssh systemctl daemon-reload

echo
echo "Running ./deploy.sh"
./deploy.sh
