#!/bin/bash
echo "Installing mongodb-org 3.2" ; sudo apt-get install 'mongodb-org=3.2.*'
echo "Installing daemontools" ; sudo apt-get install daemontools
echo "Installing bsdiff and libarchive-zip-perl" ; sudo apt-get install bsdiff libarchive-zip-perl

echo "Building" ; STARCAL_HOST=localhost ./build.sh || exit 1
sudo service starcal stop
echo "Copying binary" ; sudo cp ./server-localhost /usr/local/sbin/starcal-server || exit 1
echo "Copying init.d script" ; sudo cp ./init.d/starcal /etc/init.d/starcal
echo "Copying systemd service file" ; sudo cp ./systemd/starcal.service /lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; sudo systemctl enable starcal
echo "Reloading systemctl" ; sudo systemctl daemon-reload
echo "Restarting service" ; sudo service starcal restart

#tail -f /var/log/starcal-server/current | tai64nlocal
