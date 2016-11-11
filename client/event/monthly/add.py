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
import random

todayJd = datetime.now().toordinal() + 1721425
dayStartSeconds = random.randint(0, 24*3600-1)

params = {
    #"eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "monthly event 1",
    "description": "desc 1",
    "icon": "",

    "startJd": todayJd - 365,
    "endJd": todayJd + 365,
    "day": random.randint(1, 29),
    "dayStartSeconds": dayStartSeconds,
    "dayEndSeconds": dayStartSeconds + 3600,
}

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("starcal_email")
password = os.getenv("starcal_password")

try:
    params["groupId"] = sys.argv[1]
except IndexError:
    pass

r = requests.post(
    "http://%s:8080/event/monthly/" % host,
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
