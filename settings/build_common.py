#!/usr/bin/python3
# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import sys
import os
from os.path import join, isfile, isdir, dirname, abspath
import json
import subprocess

from typing import (
	Tuple,
	Optional,
	List,
)

import defaults

secretSettingsParams = {
	"MONGO_PASSWORD",
	"PASSWORD_SALT",
}
myDir = dirname(abspath(__file__))
rootDir = dirname(myDir)
srcDir = join(rootDir, "src")

goSettingsDir = join(srcDir, "scal", "settings")
goSettingsFile = join(goSettingsDir, "settings.go")


hostName = os.getenv("STARCAL_HOST")
print("Generating settings based on: STARCAL_HOST = %r" % hostName)
if not hostName:
	raise ValueError(
		"Set (and export) environment varibale `STARCAL_HOST` " +
		"before running this script\n" +
		"For example: export STARCAL_HOST=localhost",
	)

def goBuildAndExit(keepSettingsGo: bool):
	os.putenv("GOPATH", rootDir)
	status = subprocess.call([
		"go",
		"build",
		"-o", "server-%s" % hostName,
		"server.go",
	])

	if keepSettingsGo:
		print("Keeping settings.go file")
	else:
		os.remove(goSettingsFile)

	sys.exit(status)

class GoExpr(object):
	def __init__(
		self,
		pyType: type,
		goType: str,
		expr: str,
		imports: Optional[List[str]] = None,
	) -> None:
		self._pyType = pyType
		self._goType = goType
		self._expr = expr
		self._imports = imports

	def getGoType(self) -> str:
		return self._goType

	def getPyType(self) -> type:
		return self._pyType

	def getExpr(self) -> str:
		return self._expr

	def getImports(self) -> List[str]:
		if not self._imports:
			return []
		return self._imports



def goGetenv(varName: str) -> GoExpr:
	return GoExpr(
		str,
		"string",
		"os.Getenv(%s)" % json.dumps(varName),
		imports=["os"],
	)

def passwordStore(*args) -> str:
	from subprocess import Popen, PIPE
	cmd = Popen(["pass"] + list(args), stdout=PIPE)
	stdout, stderr = cmd.communicate()
	lastLine = stdout.decode("utf-8").strip().split("\n")[-1]
	return lastLine

def goSecretCBC(valueEncBase64: str) -> GoExpr:
	from base64 import b64decode
	b64decode(valueEncBase64) # just to validate
	return GoExpr(
		str,
		"string",
		"secretCBC(%s)" % json.dumps(valueEncBase64),
	)

# variables that are not converted to Go code
# only change the behavior of build
hostMetaParams = {
	"KEEP_SETTINGS_GO": defaults.KEEP_SETTINGS_GO,
}

hostGlobalsCommon = {
	"host": hostName,
	"GoExpr": GoExpr,
	"goGetenv": goGetenv,
	"passwordStore": passwordStore,
	"goSecretCBC": goSecretCBC,
}

# returns (goTypeStr, goValueStr)
def encodeGoValue(v) -> Tuple[str, str]:
	t = type(v)
	if t == str:
		return "string", json.dumps(v)
	elif t == int:
		return "int", str(v)
	elif t == float:
		return "float64", str(v)
	elif t == bool:
		return "bool", json.dumps(v)
	elif isinstance(v, GoExpr):
		return v.getGoType(), v.getExpr()
	return "", str(v)


