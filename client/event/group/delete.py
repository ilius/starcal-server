#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: groupId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email, password, groupId = sys.argv[1:4]

r = requests.delete(
    "http://%s:8080/event/groups/%s/" % (host, groupId),
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
