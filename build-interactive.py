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
confDir = join(homeDir, ".starcal-server")
hostHistPath = join(confDir, "host-history")


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


if not os.getenv("NO_TOUCH_SUBMODULES"):
	# takes ~ 0.2 seconds if submodules are already initialized / cloned
	print("Running: git submodule update --init")
	subprocess.call(["git", "submodule", "update", "--init"])


env = dict(os.environ)
env["STARCAL_HOST"] = hostName
exit_code = subprocess.call(
	["python3", "settings/build.py", "--interactive"] + sys.argv[1:],
	env=env,
)
