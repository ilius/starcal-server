#!/usr/bin/python3
"""
argv[1]: defaultGroupId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
defaultGroupId = sys.argv[1]

r = requests.put(
	"http://%s:9001/user/default-group/" % host,
	auth=HTTPDigestAuth(email, password),
	json={
		"defaultGroupId": defaultGroupId,
	},
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
