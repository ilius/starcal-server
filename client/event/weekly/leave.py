#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests


host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

r = requests.get(
	"http://%s:9001/event/weekly/%s/leave" % (host, eventId),
	headers={"Authorization": "bearer " + token},
)
print(r)
try:
	data = r.json()
except:
	print("non-json data: ", r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
