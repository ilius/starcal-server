#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: groupId, optional
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime, gmtime 
import random

todayJd = datetime.now().toordinal() + 1721425
dayStartSeconds = random.randint(0, 24*3600-1)

params = {
    #"eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "startJd": todayJd - 365,
    "endJd": todayJd + 365,
    "cycleWeeks": random.randint(1, 4),
    "dayStartSeconds": dayStartSeconds,
    "dayEndSeconds": dayStartSeconds + 3600,
    "summary": "daily note 1",
    "description": "desc 1",
    "icon": "",
    "durationUnit": 0,
}

email, password = sys.argv[1:3]

try:
    params["groupId"] = sys.argv[3]
except IndexError:
    pass

r = requests.post(
    "http://127.0.0.1:8080/event/weekly/",
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
