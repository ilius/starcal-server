#!/usr/bin/python3


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

r = requests.get(
	f"http://{host}:9001/me/info/",
	headers={"Authorization": "bearer " + token},
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
