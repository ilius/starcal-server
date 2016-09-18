#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
"""

import sys
import requests
from pprint import pprint

r = requests.post(
    "http://127.0.0.1:8080/user/register/",
    json={
        'email': sys.argv[1],
        'password': sys.argv[2],
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
