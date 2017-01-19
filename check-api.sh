#!/bin/bash
. ./funcs.sh
if [ -z $STARCAL_HOST ] ; then
	echo 'Set (and export) environment varibale `STARCAL_HOST` before running this script'
	exit 1
fi

scal_check_api
