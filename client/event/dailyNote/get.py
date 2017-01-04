#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests

from pprint import pprint
from datetime import datetime, timedelta

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

r = requests.get(
	"http://%s:9001/event/dailyNote/%s/" % (host, eventId),
	headers={"Authorization": "bearer " + token},
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
		print("Date:", datetime.fromordinal(data["jd"] - 1721425))
