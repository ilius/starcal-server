#!/usr/bin/python3
"""
argv[1]: eventId
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime, gmtime
import random

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
eventId = sys.argv[1]

todayJd = datetime.now().toordinal() + 1721425

params = {
    "timeZone": "CET",
    "calType": "jalali",
    "summary": "all-day task patched",
    "description": "desc patched",
    "icon": "task-patched.png",

    "startJd": todayJd-1,
    "endJd": todayJd + random.randint(1, 5),
    "durationEnable": False,
}

r = requests.patch(
    "http://%s:9001/event/allDayTask/%s/" % (host, eventId),
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
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)
