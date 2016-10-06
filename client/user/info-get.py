#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password = sys.argv[1:3]

r = requests.get(
    "http://127.0.0.1:8080/user/info/",
    auth=HTTPDigestAuth(email, password),
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
