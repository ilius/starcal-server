#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: sinceDateTime
"""

import sys
import os
import requests

from pprint import pprint
from datetime import datetime
from dateutil.parser import parse as parseDatetime

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

groupId, sinceDateTimeInput = sys.argv[1:3]

try:
	sinceDateTime = datetime.strptime(sinceDateTimeInput, "%Y-%m-%dT%H:%M:%SZ")
except:
	sinceDateTime = parseDatetime(sinceDateTimeInput)

sinceDateTimeStr = sinceDateTime.isoformat()
if not "Z" in sinceDateTimeStr:
	sinceDateTimeStr += "Z"
print("sinceDateTimeStr =", sinceDateTimeStr)

limit = 100


baseUrl = "http://%s:9001/event/groups/%s/moved-events/%s/" % (
	host,
	groupId,
	sinceDateTimeStr,
)

r = requests.get(
	baseUrl + "?limit=%d"%limit,
	headers={"Authorization": "bearer " + token},
)
print(r)
try:
	data = r.json()
except:
	print("data is not json")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
		# for event in data["movedEvents"]:
		#	print(event["time"])
