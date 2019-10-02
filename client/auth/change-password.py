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

# email, password, newPassword = sys.argv[1:4]
if len(sys.argv) > 1:
	email = sys.argv[1]

if not email:
	print("set STARCAL_EMAIL, or pass email as argument")
	sys.exit(1)

password = os.getenv("STARCAL_PASSWORD")
newPassword = os.getenv("STARCAL_PASSWORD_NEW")

if not (password and newPassword):
	print("set STARCAL_PASSWORD and STARCAL_PASSWORD_NEW")
	sys.exit(1)

r = requests.post(
	f"http://{host}:9001/auth/change-password/",
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
