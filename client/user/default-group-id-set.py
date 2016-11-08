#!/usr/bin/python3
"""
argv[1]: defaultGroupId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email = os.getenv("starcal_email")
password = os.getenv("starcal_password")
defaultGroupId = sys.argv[1]

r = requests.put(
    "http://%s:8080/user/default-group-id/" % host,
    auth=HTTPDigestAuth(email, password),
    json={
        'defaultGroupId': defaultGroupId,
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
