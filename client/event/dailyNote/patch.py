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

todayJd = datetime.now().toordinal() + 1721425

params = {
	"timeZone": "CET",
	"calType": "jalali",
	"summary": "note patched",
	"description": "desc patched",
	"icon": "note-patched.png",
	"jd": todayJd - 2,
}

r = requests.patch(
	f"http://{host}:9001/event/dailyNote/{eventId}/",
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
