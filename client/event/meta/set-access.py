#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3...]: accessEmails
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
eventType, eventId = sys.argv[1:3]
accessEmails = sys.argv[3:]

r = requests.put(
    "http://%s:9001/event/%s/%s/access/" % (
        host,
        eventType,
        eventId,
    ),
    auth=HTTPDigestAuth(email, password),
    json={
        "isPublic": False,
        "accessEmails": accessEmails,
        "publicJoinOpen": False,  # used for public events, just pass False for non-public events
        "maxAttendees": 0,  # used for public events, 0 means Unlimited
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
