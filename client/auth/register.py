#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: fullName (optional)
"""

import sys
import os
import requests
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")

email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")

if not email:
	email, password = sys.argv[1:3]

try:
	fullName = sys.argv[3]
except IndexError:
	fullName = ""

r = requests.post(
	f"http://{host}:9001/auth/register/",
	json={
		"email": email,
		"password": password,
		"fullName": fullName,
	},
)
print(r)
try:
	data = r.json()
except:
	print("non-json data")
	print(r.text)
else:
	token = data.get("token")
	error = data.get("error")
	if token:
		print(f"export STARCAL_TOKEN='{token}'")
	elif error:
		print(error)
	else:
		pprint(data, width=80)
