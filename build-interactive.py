#!/usr/bin/python3

import sys
import os
from os.path import dirname, join, abspath, isdir, isfile
import subprocess

from prompt_toolkit.history import FileHistory
from prompt_toolkit.auto_suggest import AutoSuggestFromHistory

rootDir = dirname(abspath(__file__))
settingsDir = join(rootDir, "settings")
sys.path.insert(0, settingsDir)

from settings.build_common import (
	prompt,
)


homeDir = os.getenv("HOME")
if not homeDir:
	raise ValueError("HOME is not set")
confDir = join(homeDir, ".starcal-server", "cli")
hostHistPath = join(confDir, "host-history")

os.makedirs(confDir, exist_ok=True)

def getHostName() -> str:
	while True:
		hostName = prompt(
			"Host name: ",
			history=FileHistory(hostHistPath),
			auto_suggest=AutoSuggestFromHistory(),
		)
		if hostName:
			return hostName
	raise ValueError("host name is not given")


hostName = getHostName()
print()

env = dict(os.environ)
env["STARCAL_HOST"] = hostName
exit_code = subprocess.call(
	["python3", "settings/build.py", "--interactive"] + sys.argv[1:],
	env=env,
)
