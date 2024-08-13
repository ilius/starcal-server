#!/usr/bin/python3


import os

import requests

host = os.getenv("STARCAL_HOST", "127.0.0.1")

r = requests.get(
	f"http://{host}:9001/util/api-version/",
)
print(r.text)
print(r.headers)
