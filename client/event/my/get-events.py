#!/usr/bin/python3
"""
"""

import sys
import os
import requests

from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

limit = 0
if len(sys.argv) == 2:
	limit = int(sys.argv[1])

baseUrl = "http://%s:9001/event/my/events/" % host

count = 0

exStartId = ""
while True:
	url = baseUrl + "?limit=%d&exStartId=%s" % (limit, exStartId)
	print(url)
	r = requests.get(
		url,
		headers={"Authorization": "bearer " + token},
	)
	print(r)
	try:
		data = r.json()
	except:
		print("data is not json")
		print(r.text)
		break
	
	error = data.get("error", "")
	if error:
		print(error)
		break
	
	pprint(data, width=80)
	count += len(data.get("events", []))

	lastId = data.get("lastId", "")
	if not lastId:
		break
	exStartId = lastId
	print("--------------------")
	

print("Total count:", count)
