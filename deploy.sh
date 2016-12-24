#!/bin/bash
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	exit 1
fi

./build.sh || exit $?

BIN_CACHE="./bin_cache"
PROD_BIN=/usr/local/sbin/starcal-server
NEW_BIN=server-$STARCAL_HOST

NEW_HASH=`crc32 $NEW_BIN` || exit $?
mkdir -p $BIN_CACHE; cp $NEW_BIN $BIN_CACHE/$NEW_HASH

echo "Checking remote binary CRC hash"; PROD_HASH=`ssh root@$STARCAL_HOST crc32 $PROD_BIN`
if [ "$PROD_HASH" = "$NEW_HASH" ] ; then
	echo "Remote binary is already up to date"
	exit 0
fi

BIN_PATCH=
if [ -n "$PROD_HASH" ] && \
which bsdiff >/dev/null && \
ssh root@$STARCAL_HOST which bspatch >/dev/null
then
	echo "Preparing for binary patch"
	PROD_BIN_CACHED="$BIN_CACHE/$PROD_HASH"
	if [ -f "$PROD_BIN_CACHED" ] ; then
		echo "Creating binary patch"
		bsdiff "$PROD_BIN_CACHED" "$NEW_BIN" "$NEW_BIN.patch"
		bzip2 -f "$NEW_BIN.patch"
		BIN_PATCH="$NEW_BIN.patch.bz2"
	fi
fi

if [ -n "$BIN_PATCH" ] ; then
	ls -lh "$BIN_PATCH" "$NEW_BIN"
	echo "Copying binary patch to host"; scp "$BIN_PATCH" root@$STARCAL_HOST:$PROD_BIN.patch.bz2
	echo "Applying binary patch"; ssh root@$STARCAL_HOST "
bunzip2 -f $PROD_BIN.patch.bz2 || exit $?
bspatch $PROD_BIN $PROD_BIN-new $PROD_BIN.patch
chmod u+x $PROD_BIN-new
	" || exit $?
else
	echo "Compressing binary" ; bzip2 -kf $NEW_BIN || exit $?
	echo "Copying compressed binary to host"; scp $NEW_BIN.bz2 root@$STARCAL_HOST:$PROD_BIN-new.bz2 || exit $?
	echo "Uncompressing binary" ; ssh root@$STARCAL_HOST "bunzip2 -f $PROD_BIN-new.bz2" || exit $?
fi

echo "Stopping service, updating binary and starting service again"
ssh root@$STARCAL_HOST "ls -l $PROD_BIN-new || exit $?
service starcal stop
mv -f $PROD_BIN-new $PROD_BIN
service starcal start
sleep 1
service starcal status"

echo "Cleaning up" ; rm $NEW_BIN*.bz2

# if we don't stop the service (running deamon) before updating binatry
# we will get this error:
# cp: cannot create regular file ‘/usr/local/sbin/starcal-server’: Text file busy

for API_VERSION in 1 ; do
	API_PORT="900$API_VERSION"
	echo "Checking API version $API_VERSION on port $API_PORT"
	V="`curl -s http://$STARCAL_HOST:$API_PORT/util/api-version/`"
	if [ "$V" != "$API_VERSION" ] ; then
		echo "API version mismatch: '$V' != '$API_VERSION'"
	fi
done
