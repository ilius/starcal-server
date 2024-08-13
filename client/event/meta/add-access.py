#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3]: toAddEmail.
"""

import json
import os
import sys
from pprint import pprint

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventType, eventId, toAddEmail = sys.argv[1:4]

r = requests.post(
	f"http://{host}:9001/event/{eventType}/{eventId}/access/",
	headers={"Authorization": "bearer " + token},
	json={
		"toAddEmail": toAddEmail,
	},
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("data is not json")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
