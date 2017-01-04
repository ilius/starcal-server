#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests

from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime, gmtime

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

timeFormat = "%Y-%m-%dT%H:%M:%SZ"

nowEpoch = int(time.time())
#nowDt = datetime.now()

params = {
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "task 1",
	"description": "desc 1",
	"icon": "task.png",

	"startTime": strftime(timeFormat, gmtime(nowEpoch)),
	"endTime": strftime(timeFormat, gmtime(nowEpoch - 3600)),
	"durationUnit": 0,
}

r = requests.put(
	"http://%s:9001/event/task/%s/" % (host, eventId),
	headers={"Authorization": "bearer " + token},
	json=params,
)
print(r)
try:
	data = r.json()
except:
	print("non-json data")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
