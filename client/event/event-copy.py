#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("starcal_host", "127.0.0.1")
email = os.getenv("starcal_email")
password = os.getenv("starcal_password")
eventId = sys.argv[1]

r = requests.post(
    "http://%s:8080/event/copy/" % host,
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
