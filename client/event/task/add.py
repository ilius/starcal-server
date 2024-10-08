#!/usr/bin/python3
"""argv[1]: groupId, optional."""

import json
import os
import sys
import time
from pprint import pprint
from time import gmtime, strftime

import requests

timeFormat = "%Y-%m-%dT%H:%M:%SZ"

nowEpoch = int(time.time())
# nowDt = datetime.now()

params = {
	# "eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "task 1",
	"description": "desc 1",
	"icon": "task.png",
	"startTime": strftime(timeFormat, gmtime(nowEpoch)),
	"endTime": strftime(timeFormat, gmtime(nowEpoch - 3600)),
	"durationUnit": 0,
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)


try:
	params["groupId"] = sys.argv[1]
except IndexError:
	pass

r = requests.post(
	f"http://{host}:9001/event/task/",
	headers={"Authorization": "bearer " + token},
	json=params,
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("non-json data")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
