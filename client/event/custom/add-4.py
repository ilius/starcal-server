#!/usr/bin/python3
"""
"""

import sys
import os
import requests

from pprint import pprint

params = {
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "custom event 4",
	"description": "",
	"icon": "",

	"rules": [
		{"type": "start", "value": "1396/02/30 23:55:55"},
		{"type": "duration", "value": "90 d"},
		{"type": "weekDay", "value": "1"},
		{"type": "weekNumMode", "value": "even"},
		{"type": "dayTimeRange", "value": "12:30:00 14:30:00"},
	],
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)


r = requests.post(
	f"http://{host}:9001/event/custom/",
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
