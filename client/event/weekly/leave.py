#!/usr/bin/python3
"""argv[1]: eventId."""

import json
import os
import sys

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

eventId = sys.argv[1]

r = requests.get(
	f"http://{host}:9001/event/weekly/{eventId}/leave",
	headers={"Authorization": "bearer " + token},
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("non-json data: ", r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
