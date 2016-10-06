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

nowDt = datetime.now()

params = {
    #"eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
    "timeZone": "Asia/Tehran",
    "calType": "gregorian",
    "month": nowDt.month,
    "day": nowDt.day,
    "startYear": nowDt.year - 30,
    "startYearEnable": True,
    "summary": "yeary 1",
    "description": "desc 1",
    "icon": "borthday.png",
}

email, password = sys.argv[1:3]

try:
    params["groupId"] = sys.argv[3]
except IndexError:
    pass

r = requests.post(
    "http://127.0.0.1:8080/event/yearly/",
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
