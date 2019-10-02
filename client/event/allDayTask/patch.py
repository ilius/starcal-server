#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests

from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime, gmtime
import random

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
	"summary": "all-day task patched",
	"description": "desc patched",
	"icon": "task-patched.png",

	"startJd": todayJd-1,
	"endJd": todayJd + random.randint(1, 5),
	"durationEnable": False,
}

r = requests.patch(
	f"http://{host}:9001/event/allDayTask/{eventId}/",
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
