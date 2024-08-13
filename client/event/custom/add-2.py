#!/usr/bin/python3


import json
import os
import sys
from pprint import pprint

import requests

params = {
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "custom event 2",
	"description": "",
	"icon": "",
	"rules": [
		{"type": "start", "value": "1395/02/30 23:55:55"},
		{"type": "end", "value": "1396/05/20 00:00:00"},
		{"type": "weekDay", "value": "2 4 6"},
		{"type": "ex_dates", "value": "1395/05/01 1395/06/01 1395/07/01"},
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
except json.decoder.JSONDecodeError:
	print("non-json data")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
