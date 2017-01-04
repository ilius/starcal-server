#!/usr/bin/python3
"""
argv[1]: groupTitle
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

groupTitle = sys.argv[1]

r = requests.post(
	"http://%s:9001/event/groups/" % host,
	headers={"Authorization": "bearer " + token},
	json={
		"title": groupTitle,
		#"ownerEmail": "abcde@gmail.com", # must give error
		#"groupId": "57e199d5e576da125d153b70", # must give error
		#"readAccessEmails": "test-1@gmail.com", # must give error
		#"readAccessEmails": 12345, # must give error
		#"readAccessEmails": None, # no error, no effect
		#"readAccessEmails": ["test-1@gmail.com"],
		#"addAccessEmails": ["test-2@gmail.com"],
	}
)
print(r)
try:
	data = r.json()
except:
	print("non-json data")
	print(r.text)
else:
	error = data.get("error", "")
	if error:
		print(error)
	else:
		pprint(data, width=80)
