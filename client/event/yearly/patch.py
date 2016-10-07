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

nowDt = datetime.now()

params = {
    "timeZone": "CET",
    "calType": "gregorian",
    "summary": "yearly event patched",
    "description": "desc patched",
    "icon": "",

    "month": nowDt.month,
    "day": nowDt.day,
    "startYear": nowDt.year - 30,
    "startYearEnable": True,
}

r = requests.put(
    "http://127.0.0.1:8080/event/yearly/%s/" % eventId,
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
