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

todayJd = datetime.now().toordinal() + 1721425

params = {
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "task 1",
    "description": "desc 1",
    "icon": "task.png",

    "jd": todayJd,
}

r = requests.put(
    "http://%s:8080/event/dailyNote/%s/" % (host, eventId),
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
