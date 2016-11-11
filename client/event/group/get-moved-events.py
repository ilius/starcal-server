#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: sinceDateTime
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
groupId, sinceDateTime = sys.argv[1:3]

r = requests.get(
    "http://%s:8080/event/groups/%s/moved-events/%s/" % (
        host,
        groupId,
        sinceDateTime,
    ),
    auth=HTTPDigestAuth(email, password),
)
print(r)
try:
    data = r.json()
except:
    print('data is not json')
    print(r.text)
else:
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)
