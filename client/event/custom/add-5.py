#!/usr/bin/python3


import json
import os
import sys
from pprint import pprint

import requests

params = {
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "custom event 5",
	"description": "",
	"icon": "",
	"rules": [
		{"type": "date", "value": "1390/02/30"},
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
