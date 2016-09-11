#!/usr/bin/python3

import sys
import requests
from pprint import pprint
from datetime import datetime, timedelta
import time
from time import strftime,localtime 

timeFormat = "%Y-%m-%dT%H:%M:%SZ"

nowEpoch = int(time.time())
#nowDt = datetime.now()

params = {
    "id": sys.argv[1],
    "timeZone": "Asia/Tehran",
    "calType": "jalali",
    "startTime": strftime(timeFormat, localtime(nowEpoch)),
    "endTime": strftime(timeFormat, localtime(nowEpoch - 3600)),
    "summary": "task 1",
    "description": "desc 1",
    "icon": "task.png",
    "durationUnit": 0,
}

r = requests.post(
    "http://127.0.0.1:8080/events/task/update/",
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










