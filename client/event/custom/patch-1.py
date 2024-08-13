#!/usr/bin/python3
"""argv[1]: eventId."""

import json
import os
import sys
from pprint import pprint

import requests

params = {
	# "timeZone": "Asia/Tehran",
	# "calType": "jalali",
	"summary": "custom event 1 patched",
	# "description": "",
	# "icon": "",
	"rules": [  # replaces the whole `rules` if present
		{"type": "start", "value": "1390/02/29 23:55:55"},
		{"type": "end", "value": "1396/03/23 00:00:00"},
		{"type": "year", "value": "1380-1390 1393 1396"},  # includes 1390 too
		{"type": "month", "value": "1-5 11 12"},  # includes 5 too
		{"type": "day", "value": "1-7 30 31"},  # includes 7 too
		{"type": "weekDay", "value": "5"},
	],
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

r = requests.patch(
	f"http://{host}:9001/event/custom/{eventId}/",
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
