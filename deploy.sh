#!/bin/bash
if [ -z $starcal_host ] ; then
    echo 'Set (and export) environment varibale `starcal_host` before running this script'
fi

echo "Compressing binary" ; bzip2 -kf server || exit 1
echo "Copying compressed binary to host" ; scp server.bz2 root@$starcal_host:/tmp/starcal-server.bz2 || exit 1
echo "Uncompressing binary" ; ssh root@$starcal_host "bunzip2 -f /tmp/starcal-server.bz2" || exit 1

echo "Stopping service, updating binary and starting service again"
ssh root@$starcal_host "service starcal stop
cp /tmp/starcal-server /usr/local/sbin/starcal-server
service starcal start
rm /tmp/starcal-server*"

echo "Cleaning up" ; rm server.bz2

# if we don't stop the service (running deamon) before updating binatry
# we will get this error:
# cp: cannot create regular file ‘/usr/local/sbin/starcal-server’: Text file busy


