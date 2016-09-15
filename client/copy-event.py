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

r = requests.post(
    "http://127.0.0.1:8080/events/copy/",
    auth=HTTPDigestAuth(sys.argv[1], sys.argv[2]),
    json={
        'eventId': sys.argv[3],
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




