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

email, password, eventId = sys.argv[1:4]

todayJd = datetime.now().toordinal() + 1721425

params = {
    "timeZone": "CET",
    "calType": "jalali",
    "summary": "note patched",
    "description": "desc patched",
    "icon": "note-patched.png",

    "jd": todayJd-2,
}

r = requests.patch(
    "http://127.0.0.1:8080/event/dailyNote/%s/" % eventId,
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