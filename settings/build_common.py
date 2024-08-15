# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import json
import os
import subprocess
import sys
from os.path import abspath, dirname, join

import defaults
from prompt_toolkit import prompt as promptLow

goZeroValueByType = {
	str: '""',
	int: "0",
	float: "0.0",
	bool: "false",
}

secretSettingsParams = {
	"MONGO_PASSWORD",
}
myDir = dirname(abspath(__file__))
rootDir = dirname(myDir)
pkgDir = join(rootDir, "pkg")
settingsDir = join(rootDir, "settings")

goSettingsDir = join(pkgDir, "scal", "settings")
goSettingsFile = join(goSettingsDir, "settings.go")


hostName = os.getenv("STARCAL_HOST")
print(f"Generating settings based on: STARCAL_HOST = {hostName!r}")
if not hostName:
	raise ValueError(
		"Set (and export) environment varibale `STARCAL_HOST` "
		"before running this script\n"
		"For example: export STARCAL_HOST=localhost",
	)


def goBuildAndExit(keepSettingsGo: bool):
	status = subprocess.call(
		[
			"go",
			"build",
			"-o",
			f"server-{hostName}",
			"server.go",
		]
	)

	if keepSettingsGo:
		print("Keeping settings.go file")
	else:
		os.remove(goSettingsFile)

	sys.exit(status)


class GoExpr:
	def __init__(
		self,
		pyType: type,
		goType: str,
		expr: str,
		imports: list[str] | None = None,
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

	def getImports(self) -> list[str]:
		if not self._imports:
			return []
		return self._imports


def goGetenv(varName: str) -> GoExpr:
	return GoExpr(
		str,
		"string",
		f"os.Getenv({json.dumps(varName)})",
		imports=["os"],
	)


def passwordStore(*args) -> str:
	from subprocess import PIPE, Popen

	cmd = Popen(["pass"] + list(args), stdout=PIPE)
	stdout, stderr = cmd.communicate()
	return stdout.decode("utf-8").strip().split("\n")[-1]


def goSecretCBC(valueEncBase64: str) -> GoExpr:
	from base64 import b64decode

	b64decode(valueEncBase64)  # just to validate
	return GoExpr(
		str,
		"string",
		f"secretCBC({json.dumps(valueEncBase64)})",
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
def encodeGoValue(v) -> tuple[str, str]:
	t = type(v)
	if t == str:
		return "string", json.dumps(v)
	if t == int:
		return "int", str(v)
	if t == float:
		return "float64", str(v)
	if t == bool:
		return "bool", json.dumps(v)
	if isinstance(v, GoExpr):
		return v.getGoType(), v.getExpr()
	return "", str(v)


def prompt(
	message: str,
	multiline: bool = False,
	**kwargs,
):
	text = promptLow(message=message, **kwargs)
	if multiline and text == "!m":
		print("Entering Multi-line mode, press Alt+Enter to end")
		text = promptLow(message="", multiline=True, **kwargs)
	return text
