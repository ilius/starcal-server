#!/usr/bin/python3

import sys
import requests
from pprint import pprint

r = requests.post(
    "http://127.0.0.1:8080/events/task/get/",
    json={
        'eventId': sys.argv[1],
    },
)
print(r)
try:
    data = r.json()
except:
    print('data is not json')
    print(r.text)
else:
    error = data.get('error', '')
    if error:
        print(error)
    else:
        pprint(data, width=80)




