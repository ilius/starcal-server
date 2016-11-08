#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: groupId
argv[4]: sinceDateTime
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email, password, groupId, sinceDateTime = sys.argv[1:5]

r = requests.get(
    "http://%s:8080/event/groups/%s/modified-events/%s/" % (
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
