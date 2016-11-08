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

host = os.getenv("starcal_host", "127.0.0.1")
email = os.getenv("starcal_email")
password = os.getenv("starcal_password")
eventId = sys.argv[1]

timeFormat = "%Y-%m-%dT%H:%M:%SZ"

nowEpoch = int(time.time())
#nowDt = datetime.now()

params = {
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "task 1",
    "description": "desc 1",
    "icon": "task.png",

    "startTime": strftime(timeFormat, gmtime(nowEpoch)),
    "endTime": strftime(timeFormat, gmtime(nowEpoch - 3600)),
    "durationUnit": 0,
}

r = requests.put(
    "http://%s:8080/event/task/%s/" % (host, eventId),
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
