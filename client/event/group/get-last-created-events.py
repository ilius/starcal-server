#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: count.
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

groupId = sys.argv[1]
try:
	maxCount = sys.argv[2]
except IndexError:
	maxCount = None
else:
	maxCount = int(maxCount)

url = f"http://{host}:9001/event/groups/{groupId}/last-created-events/"
if maxCount is not None:
	url += f"?maxCount={maxCount}"

print(url)

r = requests.get(
	url,
	headers={"Authorization": "bearer " + token},
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
