#!/usr/bin/python3
"""
argv[1]: eventType
argv[2]: eventId
argv[3]: newGroupId
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

oldGroupId, newGroupId = sys.argv[1:3]


limit = 20

oldGroupUrl = "http://%s:9001/event/groups/%s/" % (host, oldGroupId)
newGroupUrl = "http://%s:9001/event/groups/%s/" % (host, newGroupId)


exStartId = ""
while True:
	url = oldGroupUrl + "events/?limit=%d&exStartId=%s" % (limit, exStartId)
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

	for event in data.get("events", []):
		r = requests.put(
			"http://%s:9001/event/%s/%s/group/" % (
				host,
				event["eventType"],
				event["eventId"],
			),
			headers={"Authorization": "bearer " + token},
			json={
				"newGroupId": newGroupId,
			},
		)
		print(r)

	lastId = data.get("lastId", "")
	if not lastId:
		break

	exStartId = lastId
	print("------------------------")
