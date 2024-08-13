#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: sinceDateTime.
"""

import json
import os
import sys
from pprint import pprint

import requests
from dateutil.parser import parse as parseDatetime

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

groupId, sinceDateTimeInput = sys.argv[1:3]

if "T" in sinceDateTimeInput:
	sinceDateTimeStr = sinceDateTimeInput
else:
	sinceDateTimeStr = parseDatetime(sinceDateTimeInput).isoformat()

if "Z" not in sinceDateTimeStr and "+" not in sinceDateTimeStr:
	sinceDateTimeStr += "Z"


print("sinceDateTimeStr =", sinceDateTimeStr)

limit = 10

baseUrl = (
	f"http://{host}:9001/event/groups/{groupId}/modified-events/{sinceDateTimeStr}/"
)

url = f"{baseUrl}?limit={limit}"

r = requests.get(
	url,
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
