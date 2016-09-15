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

timeFormat = "%Y-%m-%dT%H:%M:%SZ"

nowEpoch = int(time.time())
#nowDt = datetime.now()

params = {
    "eventId": sys.argv[3],
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "startTime": strftime(timeFormat, gmtime(nowEpoch)),
    "endTime": strftime(timeFormat, gmtime(nowEpoch - 3600)),
    "summary": "task 1",
    "description": "desc 1",
    "icon": "task.png",
    "durationUnit": 0,
}

r = requests.post(
    "http://127.0.0.1:8080/events/task/update/",
    auth=HTTPDigestAuth(sys.argv[1], sys.argv[2]),
    json=params,
)
print(r)
#print(r.text)
if r.text.strip():
    data = r.json()
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)










