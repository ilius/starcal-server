#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventType
argv[4]: eventId
argv[5...]: accessEmails
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password, eventType, eventId = sys.argv[1:5]
accessEmails = sys.argv[5:]

r = requests.put(
    "http://127.0.0.1:8080/event/%s/%s/access/" % (
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
