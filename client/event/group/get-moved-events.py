#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: sinceDateTime.
"""

import json
import os
import sys
from datetime import datetime
from pprint import pprint

import requests
from dateutil.parser import parse as parseDatetime

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

groupId, sinceDateTimeInput = sys.argv[1:3]

# TODO: use dateutil.parser.parse
try:
	sinceDateTime = datetime.strptime(sinceDateTimeInput, "%Y-%m-%dT%H:%M:%SZ")  # noqa: DTZ007
except json.decoder.JSONDecodeError:
	sinceDateTime = parseDatetime(sinceDateTimeInput)

sinceDateTimeStr = sinceDateTime.isoformat()
if "Z" not in sinceDateTimeStr:
	sinceDateTimeStr += "Z"
print("sinceDateTimeStr =", sinceDateTimeStr)

limit = 100


baseUrl = f"http://{host}:9001/event/groups/{groupId}/moved-events/{sinceDateTimeStr}/"

r = requests.get(
	f"{baseUrl}?limit={limit}",
	headers={"Authorization": "bearer " + token},
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("data is not json")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
		# for event in data["movedEvents"]:
		# print(event["time"])
