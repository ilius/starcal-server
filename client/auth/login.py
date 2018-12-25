#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
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


r = requests.post(
	"http://%s:9001/auth/login/" % host,
	json={
		"email": email,
		"password": password,
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
	if token:
		print("export STARCAL_TOKEN='%s'" % token)
	else:
		pprint(data, width=80)
