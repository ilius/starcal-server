#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventType
argv[4]: eventId
argv[5...]: accessEmails
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email, password, eventType, eventId = sys.argv[1:5]
accessEmails = sys.argv[5:]

r = requests.put(
    "http://%s:8080/event/%s/%s/access/" % (
        host,
        eventType,
        eventId,
    ),
    auth=HTTPDigestAuth(email, password),
    json={
        'isPublic': False,
        'accessEmails': accessEmails,
        #'publicJoinOpen': False,
        #'maxAttendees': 0,
    },
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
