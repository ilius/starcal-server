#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
    echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
    exit 1
fi

./build.sh

PROD_BIN=/usr/local/sbin/starcal-server

echo "Compressing binary" ; bzip2 -kf server-$STARCAL_HOST || exit 1
echo "Copying compressed binary to host" ; scp server-$STARCAL_HOST.bz2 root@$STARCAL_HOST:$PROD_BIN-new.bz2 || exit 1
echo "Uncompressing binary" ; ssh root@$STARCAL_HOST "bunzip2 -f $PROD_BIN-new.bz2" || exit 1

echo "Stopping service, updating binary and starting service again"
ssh root@$STARCAL_HOST "ls -l $PROD_BIN-new || exit 1
service starcal stop
mv -f $PROD_BIN-new $PROD_BIN
service starcal start
service starcal status"

echo "Cleaning up" ; rm server-$STARCAL_HOST.bz2

# if we don't stop the service (running deamon) before updating binatry
# we will get this error:
# cp: cannot create regular file ‘/usr/local/sbin/starcal-server’: Text file busy

