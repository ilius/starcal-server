#!/usr/bin/python3


import json
import os
import sys

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

limit = 0
if len(sys.argv) == 2:
	limit = int(sys.argv[1])

count = 0

exStartId = ""

while True:
	url = f"http://{host}:9001/event/ungrouped/?limit={limit}&exStartId={exStartId}"
	print(url)
	r = requests.get(
		url,
		headers={"Authorization": "bearer " + token},
	)
	print(r)
	try:
		data = r.json()
	except json.decoder.JSONDecodeError:
		print("non-json data")
		print(r.text)
		break
	error = data.get("error", "")
	if error:
		print(error)
		break

	pageCount = len(data.get("events", []))
	count += pageCount

	# pprint(data, width=80)
	print(pageCount)

	lastId = data.get("lastId", "")
	if not lastId:
		break
	exStartId = lastId
	print("--------------------")

print("Total count:", count)
