#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3]: newGroupId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
eventType, eventId, newGroupId = sys.argv[1:4]

r = requests.put(
    "http://%s:9001/event/%s/%s/group/" % (
        host,
        eventType,
        eventId,
    ),
    auth=HTTPDigestAuth(email, password),
    json={
        "newGroupId": newGroupId,
    },
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
