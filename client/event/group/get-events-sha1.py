#!/usr/bin/python3
"""argv[1]: groupId."""

import json
import os
import sys

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")
token = os.getenv("STARCAL_TOKEN")
if not token:
	print("Please set and export STARCAL_TOKEN")
	sys.exit(1)

groupId = sys.argv[1]

limit = 0
if len(sys.argv) == 3:
	limit = int(sys.argv[2])

baseUrl = f"http://{host}:9001/event/groups/{groupId}/events-sha1/"

count = 0

exStartId = ""
while True:
	url = f"{baseUrl}?limit={limit}&exStartId={exStartId}"
	print(url)
	r = requests.get(
		url,
		headers={"Authorization": "bearer " + token},
	)
	print(r)
	try:
		data = r.json()
	except json.decoder.JSONDecodeError:
		print("data is not json")
		print(r.text)
		break
	error = data.get("error", "")
	if error:
		print(error)
		break

	pageCount = len(data.get("events", []))

	# pprint(data, width=80)
	print(pageCount)
	count += pageCount

	lastId = data.get("lastId", "")
	if not lastId:
		break
	exStartId = lastId
	print("--------------------")

print("Total count:", count)
