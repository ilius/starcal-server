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
email, password, newPassword = sys.argv[1:4]

r = requests.post(
	"http://%s:9001/auth/change-password/" % host,
	json={
		"email": email,
		"password": password,
		"newPassword": newPassword,
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
