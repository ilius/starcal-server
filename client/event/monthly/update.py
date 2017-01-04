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
dayStartSeconds = random.randint(0, 24*3600-1)

params = {
	"timeZone": "Asia/Tehran",
	"calType": "jalali",
	"summary": "monthly event 1",
	"description": "desc 1",
	"icon": "",

	"startJd": todayJd - 365,
	"endJd": todayJd + 2*365,
	"day": random.randint(1, 29),
	"dayStartSeconds": dayStartSeconds,
	"dayEndSeconds": dayStartSeconds + 3600,
}

r = requests.put(
	"http://%s:9001/event/monthly/%s/" % (host, eventId),
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
