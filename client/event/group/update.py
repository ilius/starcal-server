#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: groupTitle.
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

groupId, groupTitle = sys.argv[1:3]

# not passing "readAccessEmails" will remove it if it was set before
# not passing "addAccessEmails" will remove it if it was set before

r = requests.put(
	f"http://{host}:9001/event/groups/{groupId}/",
	headers={"Authorization": "bearer " + token},
	json={
		"title": groupTitle,
		# "title": "", # must give error
		# "title": None, # must give error
		# "title": [], # must give error
		# "ownerEmail": "abcde@gmail.com", # must give error
		# "groupId": "57e199d5e576da125d153b70", # must give error
		# "readAccessEmails": "test-1@gmail.com", # must give error
		# "readAccessEmails": 12345, # must give error
		# "readAccessEmails": None, # will unset the value
		"readAccessEmails": ["test-1@gmail.com"],
		"addAccessEmails": ["test-2@gmail.com"],
	},
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("non-json data")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
