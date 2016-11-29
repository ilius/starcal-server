#!/usr/bin/python3
"""
argv[1]: groupId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
groupId = sys.argv[1]

r = requests.get(
    "http://%s:9001/event/groups/%s/events/" % (host, groupId),
    auth=HTTPDigestAuth(email, password),
)
print(r)
try:
    data = r.json()
except:
    print("data is not json")
    print(r.text)
else:
    error = data.get("error", "")
    if error:
        print(error)
    else:
        pprint(data, width=80)
