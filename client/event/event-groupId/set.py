#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3]: newGroupId
"""

import sys
import os
import requests

from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventType, eventId, newGroupId = sys.argv[1:4]

r = requests.put(
	"http://%s:9001/event/%s/%s/group/" % (
		host,
		eventType,
		eventId,
	),
	headers={"Authorization": "bearer " + token},
	json={
		"newGroupId": newGroupId,
	},
)
print(r)
try:
	data = r.json()
except:
	print("data is not json")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
