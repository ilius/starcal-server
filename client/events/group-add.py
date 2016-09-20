#!/usr/bin/python3
"""
argv[1]: email
argv[2]: password
argv[3]: groupTitle
"""

import sys
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

r = requests.post(
    "http://127.0.0.1:8080/events/groups/" ,
    auth=HTTPDigestAuth(sys.argv[1], sys.argv[2]),
    json={
        "title": sys.argv[3],
        #"ownerEmail": "abcde@gmail.com", # must give error
        #"groupId": "57e199d5e576da125d153b70", # must give error
        #"readAccessEmails": "test-1@gmail.com", # must give error
        #"readAccessEmails": 12345, # must give error
        #"readAccessEmails": None, # no error, no effect
        #"readAccessEmails": ["test-1@gmail.com"],
        #"addAccessEmails": ["test-2@gmail.com"],
    }
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
