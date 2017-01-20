#!/bin/bash
. ./funcs.sh
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	exit 1
fi

# Install package 'dpkg' to provide command 'start-stop-daemon'
echo "Installing dpkg" ; scal_ssh yum install dpkg

echo "Adding mongodb repository" ; scal_ssh "echo '[mongodb-org-3.2]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/\$releasever/mongodb-org/3.2/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://www.mongodb.org/static/pgp/server-3.2.asc' > /etc/yum.repos.d/mongodb-org.repo"

echo "Installing mongodb-server 3.2.8" ; scal_ssh yum install mongodb-server-3.2.8 || exit $?

# http://djbware.csi.hu/daemontools.html
echo "Downloading and installing daemontools"
scal_ssh yum install 'http://djbware.csi.hu/RPMS/daemontools-0.76-112memphis.i386.rpm' || exit $?

echo "Installing bsdiff and perl-Archive-Zip" ; scal_ssh yum install bsdiff perl-Archive-Zip || exit $?
echo "Copying init.d script" ; scal_scp_to_host ./init.d/starcal /etc/init.d/starcal || exit $?
echo "Copying systemd service file" ; scal_scp_to_host ./systemd/starcal.service /lib/systemd/system/ || exit $?
echo "Enabling systemd service" ; scal_ssh systemctl enable starcal || exit $?
echo "Reloading systemctl" ; scal_ssh systemctl daemon-reload || exit $?

echo
echo "Running ./deploy.sh"
./deploy.sh || exit $?
