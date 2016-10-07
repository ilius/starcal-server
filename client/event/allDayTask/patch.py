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
from datetime import datetime, timedelta
import time
from time import strftime, gmtime
import random

email, password, eventId = sys.argv[1:4]

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
    "http://127.0.0.1:8080/event/allDayTask/%s/" % eventId,
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
