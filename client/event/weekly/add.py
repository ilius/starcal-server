#!/usr/bin/python3
"""argv[1]: groupId, optional."""

import json
import os
import random
import sys
from datetime import datetime
from pprint import pprint

import requests

todayJd = datetime.now().toordinal() + 1721425
dayStartSeconds = random.randint(0, 24 * 3600 - 1)

params = {
	# "eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "weekly 1",
	"description": "desc 1",
	"icon": "",
	"startJd": todayJd - 365,
	"endJd": todayJd + 365,
	"cycleWeeks": random.randint(1, 4),
	"dayStartSeconds": dayStartSeconds,
	"dayEndSeconds": dayStartSeconds + 3600,
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
	f"http://{host}:9001/event/weekly/",
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
