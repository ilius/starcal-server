#!/bin/bash

echo "Building" ; ./build.sh      || exit 1
sudo service starcal stop
echo "Copying binary" ; sudo cp ./server /usr/local/sbin/starcal-server || exit 1
echo "Copying init.d script" ; sudo cp ./init.d/starcal /etc/init.d/starcal
echo "Copying systemd service file" ; sudo cp ./systemd/starcal.service /lib/systemd/system/ || exit 1
echo "Enabling systemd service" ; sudo systemctl enable starcal
echo "Reloading systemctl" ; sudo systemctl daemon-reload
echo "Restarting service" ; sudo service starcal restart

