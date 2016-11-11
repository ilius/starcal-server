#!/usr/bin/python3
"""
argv[1]: groupId, optional
"""

import sys
import os
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
    #"eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "task 1",
    "description": "desc 1",
    "icon": "task.png",

    "startTime": strftime(timeFormat, gmtime(nowEpoch)),
    "endTime": strftime(timeFormat, gmtime(nowEpoch - 3600)),
    "durationUnit": 0,
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("starcal_email")
password = os.getenv("starcal_password")

try:
    params["groupId"] = sys.argv[1]
except IndexError:
    pass

r = requests.post(
    "http://%s:8080/event/task/" % host,
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
