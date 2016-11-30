#!/usr/bin/python3
"""
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime, gmtime 

params = {
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "custom event 5",
    "description": "",
    "icon": "",

    "rules": [
        {"type": "date", "value": "1390/02/30"},
        {"type": "dayTimeRange", "value": "12:30:00 14:30:00"},
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
