#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventId
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password, eventId = sys.argv[1:4]

r = requests.post(
    "http://127.0.0.1:8080/event/copy/",
    auth=HTTPDigestAuth(email, password),
    json={
        'eventId': eventId,
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
