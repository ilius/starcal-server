#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: count
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

groupId, count = sys.argv[1:3]
count = int(count)

r = requests.get(
	"http://%s:9001/event/groups/%s/last-created-events/%s/" % (
		host,
		groupId,
		count,
	),
	headers={"Authorization": "bearer " + token},
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
