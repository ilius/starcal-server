#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint
from datetime import datetime, timedelta

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
eventId = sys.argv[1]

r = requests.get(
	"http://%s:9001/event/allDayTask/%s/" % (host, eventId),
	auth=HTTPDigestAuth(email, password),
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
		print("Start Date:", datetime.fromordinal(data["startJd"] - 1721425))
		print("End Date:", datetime.fromordinal(data["endJd"] - 1721425))
