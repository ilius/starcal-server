#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password.
"""

import json
import os
import sys
from pprint import pprint

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")

email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")

if not email:
	email, password = sys.argv[1:3]


r = requests.post(
	f"http://{host}:9001/auth/login/",
	json={
		"email": email,
		"password": password,
	},
)
print(r)
try:
	data = r.json()
except json.decoder.JSONDecodeError:
	print("non-json data")
	print(r.text)
else:
	token = data.get("token")
	if token:
		print(f"export STARCAL_TOKEN='{token}'")
	else:
		pprint(data, width=80)
