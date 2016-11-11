#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
    echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
    exit 1
fi

./build.sh

echo "Compressing binary" ; bzip2 -kf server-$STARCAL_HOST || exit 1
echo "Copying compressed binary to host" ; scp server-$STARCAL_HOST.bz2 root@$STARCAL_HOST:/tmp/starcal-server.bz2 || exit 1
echo "Uncompressing binary" ; ssh root@$STARCAL_HOST "bunzip2 -f /tmp/starcal-server.bz2" || exit 1

echo "Stopping service, updating binary and starting service again"
ssh root@$STARCAL_HOST "service starcal stop
cp /tmp/starcal-server /usr/local/sbin/starcal-server
service starcal start
rm /tmp/starcal-server*"

echo "Cleaning up" ; rm server-$STARCAL_HOST.bz2

# if we don't stop the service (running deamon) before updating binatry
# we will get this error:
# cp: cannot create regular file ‘/usr/local/sbin/starcal-server’: Text file busy


