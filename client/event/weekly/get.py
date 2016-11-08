#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: eventId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint
from datetime import datetime, timedelta

host = os.getenv("starcal_host", "127.0.0.1")
email, password, eventId = sys.argv[1:4]

r = requests.get(
    "http://%s:8080/event/weekly/%s/" % (host, eventId),
    auth=HTTPDigestAuth(email, password),
)
print(r)
try:
    data = r.json()
except:
    print("non-json data")
    print(r.text)
else:
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)
        print('Start Date:', datetime.fromordinal(data["startJd"] - 1721425))
        print('End Date:', datetime.fromordinal(data["endJd"] - 1721425))
