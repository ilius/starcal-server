#!/usr/bin/python3
"""
argv[1]: eventId
argv[2...]: inviteEmails
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
inviteEmails = sys.argv[2:]

if not inviteEmails:
	print("no emails given to invite")
	sys.exit(1)

r = requests.post(
	"http://%s:9001/event/task/%s/invite/" % (host, eventId),
	headers={"Authorization": "bearer " + token},
	json={
		"inviteEmails": inviteEmails,
	}
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
