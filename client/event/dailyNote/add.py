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

todayJd = datetime.now().toordinal() + 1721425

params = {
    #"eventId": "57d5e9fee576da5246cbe122",# must show: "you can't specify 'eventId'"
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "summary": "daily note 1",
    "description": "desc 1",
    "icon": "note.png",

    "jd": todayJd,
}

email, password = sys.argv[1:3]

try:
    params["groupId"] = sys.argv[3]
except IndexError:
    pass

r = requests.post(
    "http://127.0.0.1:8080/event/dailyNote/",
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
