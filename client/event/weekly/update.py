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
import time
from time import strftime, gmtime
import random

host = os.getenv("starcal_host", "127.0.0.1")
email, password, eventId = sys.argv[1:4]

todayJd = datetime.now().toordinal() + 1721425
dayStartSeconds = random.randint(0, 24*3600-1)

params = {
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "weekly event 1",
    "description": "desc 1",
    "icon": "",

    "startJd": todayJd - 365,
    "endJd": todayJd + 2*365,
    "cycleWeeks": random.randint(1, 4),
    "dayStartSeconds": dayStartSeconds,
    "dayEndSeconds": dayStartSeconds + 3600,
}

r = requests.put(
    "http://%s:8080/event/weekly/%s/" % (host, eventId),
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
