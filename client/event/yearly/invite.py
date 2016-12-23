#!/usr/bin/python3
"""
argv[1]: eventId
argv[2...]: inviteEmails
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
eventId = sys.argv[1]
inviteEmails = sys.argv[2:]

if not inviteEmails:
	print("no emails given to invite")
	sys.exit(1)

r = requests.post(
	"http://%s:9001/event/yearly/%s/invite/" % (host, eventId),
	auth=HTTPDigestAuth(email, password),
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
