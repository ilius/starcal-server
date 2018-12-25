#!/bin/bash

function scal_ssh {
	if [ -z $STARCAL_SSH_PORT ] ; then
		STARCAL_SSH_PORT=22
	fi
	ssh -p $STARCAL_SSH_PORT root@$STARCAL_HOST "$@"
	return $?
}

function scal_scp_to_host {
	if [ -z $STARCAL_SSH_PORT ] ; then
		STARCAL_SSH_PORT=22
	fi
	scp -P $STARCAL_SSH_PORT "$1" "root@$STARCAL_HOST:$2"
	return $?
}

function scal_check_api {
	for API_VERSION in 1 ; do
		API_PORT="900$API_VERSION"
		echo "Checking API version $API_VERSION on port $API_PORT"
		V="`curl -s http://$STARCAL_HOST:$API_PORT/util/api-version/`"
		if [ "$V" != "$API_VERSION" ] ; then
			echo "API version mismatch: '$V' != '$API_VERSION'"
			return 1
		fi
		echo "OK: api v$API_VERSION"
	done
	return 0
}

