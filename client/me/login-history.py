#!/usr/bin/python3
"""
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


limit = 0
if len(sys.argv) == 2:
	limit = int(sys.argv[1])

r = requests.get(
	"http://%s:9001/me/login-history/" % host,
	headers={"Authorization": "bearer " + token},
	json={"limit": limit}
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
