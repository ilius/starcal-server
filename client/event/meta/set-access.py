#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3...]: accessEmails
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

eventType, eventId = sys.argv[1:3]
accessEmails = sys.argv[3:]

r = requests.put(
	"http://%s:9001/event/%s/%s/access/" % (
		host,
		eventType,
		eventId,
	),
	headers={"Authorization": "bearer " + token},
	json={
		"isPublic": False,
		"accessEmails": accessEmails,
		"publicJoinOpen": False,  # used for public events, just pass False for non-public events
		"maxAttendees": 0,  # used for public events, 0 means Unlimited
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
