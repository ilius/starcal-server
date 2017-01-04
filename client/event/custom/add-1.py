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
	"summary": "custom event 1",
	"description": "",
	"icon": "",

	"rules": [
		{"type": "start", "value": "1390/02/30 23:55:55"},
		{"type": "end", "value": "1396/03/20 00:00:00"},
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


r = requests.post(
	"http://%s:9001/event/custom/" % host,
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
