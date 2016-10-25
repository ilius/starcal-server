#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventType
argv[4]: eventId
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password, eventType, eventId = sys.argv[1:6]

r = requests.get(
    "http://127.0.0.1:8080/event/%s/%s/owner/" % (
        eventType,
        eventId,
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
