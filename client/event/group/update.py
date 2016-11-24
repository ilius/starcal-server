#!/usr/bin/python3
"""
argv[1]: groupId
argv[2]: groupTitle
"""

import sys
import os
import requests
from requests.auth import HTTPDigestAuth
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
email = os.getenv("STARCAL_EMAIL")
password = os.getenv("STARCAL_PASSWORD")
groupId, groupTitle = sys.argv[1:3]

# not passing "readAccessEmails" will remove it if it was set before
# not passing "addAccessEmails" will remove it if it was set before

r = requests.put(
    "http://%s:9001/event/groups/%s/" % (host, groupId),
    auth=HTTPDigestAuth(email, password),
    json={
        "title": groupTitle,
        #"title": "", # must give error
        #"title": None, # must give error
        #"title": [], # must give error
        #"ownerEmail": "abcde@gmail.com", # must give error
        #"groupId": "57e199d5e576da125d153b70", # must give error
        #"readAccessEmails": "test-1@gmail.com", # must give error
        #"readAccessEmails": 12345, # must give error
        #"readAccessEmails": None, # will unset the value
        "readAccessEmails": ["test-1@gmail.com"],
        "addAccessEmails": ["test-2@gmail.com"],
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
