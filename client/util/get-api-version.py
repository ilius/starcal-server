#!/usr/bin/python3
"""
"""

import sys
import os
import requests
from pprint import pprint

host = os.getenv("STARCAL_HOST", "127.0.0.1")

r = requests.get(
	f"http://{host}:9001/util/api-version/",
)
print(r.text)
print(r.headers)
