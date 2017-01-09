#!/usr/bin/python3
"""
argv[1]: email
"""

import sys
import os
import requests
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = sys.argv[1]

r = requests.post(
	"http://%s:9001/auth/reset-password-request/" % host,
	json={
		"email": email,
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
