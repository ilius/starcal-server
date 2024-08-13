#!/usr/bin/python3
"""argv[1]: eventId."""

import json
import os
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

nowDt = datetime.now()

params = {
	"timeZone": "CET",
	"calType": "gregorian",
	"summary": "yearly event patched",
	"description": "desc patched",
	"icon": "",
	"month": nowDt.month,
	"day": nowDt.day,
	"startYear": nowDt.year - 30,
	"startYearEnable": True,
}

r = requests.put(
	f"http://{host}:9001/event/yearly/{eventId}/",
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
