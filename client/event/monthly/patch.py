#!/usr/bin/python3
"""argv[1]: eventId."""

import json
import os
import random
import sys
from datetime import datetime
from pprint import pprint

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

todayJd = datetime.now().toordinal() + 1721425
dayStartSeconds = random.randint(0, 24 * 3600 - 1)

params = {
	"timeZone": "CET",
	"timeZoneEnable": True,
	"calType": "jalali",
	"summary": "monthly event patched",
	"description": "patched desc",
	"icon": "patched icon",
	# "foo": "", # must give error
	"startJd": todayJd - 300,
	"endJd": todayJd + 2 * 365,
	"day": random.randint(1, 29),
	"dayStartSeconds": dayStartSeconds,
	"dayEndSeconds": dayStartSeconds + 3600,
}

r = requests.patch(
	f"http://{host}:9001/event/monthly/{eventId}/",
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
