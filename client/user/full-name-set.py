#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: fullName
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password, fullName = sys.argv[1:4]

r = requests.put(
    "http://127.0.0.1:8080/user/full-name/",
    auth=HTTPDigestAuth(email, password),
    json={
        'fullName': fullName,
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
