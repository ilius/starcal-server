#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
"""

import sys
import os
import requests
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email, password = sys.argv[1:3]

r = requests.post(
    "http://%s:8080/user/register/" % host,
    json={
        'email': email,
        'password': password,
    },
)
print(r)
try:
    data = r.json()
except:
    print("non-json data")
    print(r.text)
else:
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)
