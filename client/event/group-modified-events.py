#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: groupId
argv[4]: sinceDateTime
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

r = requests.get(
    "http://127.0.0.1:8080/event/groups/%s/modified-events/%s/" % (
        sys.argv[3],
        sys.argv[4],
    ),
    auth=HTTPDigestAuth(sys.argv[1], sys.argv[2]),
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
