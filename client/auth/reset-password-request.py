#!/usr/bin/python3
"""argv[1]: email."""

import json
import os
import sys
from pprint import pprint

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = sys.argv[1]

r = requests.post(
	f"http://{host}:9001/auth/reset-password-request/",
	json={
		"email": email,
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
