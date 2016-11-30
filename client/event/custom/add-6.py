#!/usr/bin/python3
"""
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

params = {
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "custom event 6",
    "description": "",
    "icon": "",

    "rules": [
        {"type": "start", "value": "1390/02/30 23:55:55"},
        {"type": "end", "value": "1396/03/20 00:00:00"},
        {"type": "cycleDays", "value": "3"},
        {"type": "dayTimeRange", "value": "14:00:00 15:30:00"}

    ],
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")

r = requests.post(
    "http://%s:9001/event/custom/" % host,
    auth=HTTPDigestAuth(email, password),
    json=params,
)
print(r)
try:
    data = r.json()
except:
    print("non-json data")
    print(r.text)
else:
    error = data.get("error", "")
    if error:
        print(error)
    else:
        pprint(data, width=80)
