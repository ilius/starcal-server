#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventType
argv[4]: eventId
argv[5]: newGroupId
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

email, password, eventType, eventId, newGroupId = sys.argv[1:6]

r = requests.put(
    "http://127.0.0.1:8080/event/%s/%s/groupId/" % (
        eventType,
        eventId,
    ),
    auth=HTTPDigestAuth(email, password),
    json={
        'newGroupId': newGroupId,
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
