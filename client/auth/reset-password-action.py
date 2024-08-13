#!/usr/bin/python3
"""
argv[1]: email
argv[2]: resetPasswordToken
argv[3]: newPassword.
"""

import json
import os
import sys
from pprint import pprint

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email, resetPasswordToken, newPassword = sys.argv[1:4]

try:
	fullName = sys.argv[3]
except IndexError:
	fullName = ""

r = requests.post(
	f"http://{host}:9001/auth/reset-password-action/",
	json={
		"email": email,
		"resetPasswordToken": resetPasswordToken,
		"newPassword": newPassword,
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
