#!/usr/bin/python3
# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import sys
import os
from os.path import join, isfile, isdir, dirname, abspath
import json
import subprocess
from pprint import pprint

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


hostName = os.getenv("STARCAL_HOST")
print("Generating settings based on: STARCAL_HOST = %r" % hostName)
if not hostName:
	raise ValueError(
		"Set (and export) environment varibale `STARCAL_HOST` " +
		"before running this script\n" +
		"For example: export STARCAL_HOST=localhost",
	)

goSettingsDir = join(srcDir, "scal", "settings")
goSettingsFile = join(goSettingsDir, "settings.go")

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


if isfile(goSettingsFile):
	defaultsFile = join(myDir, "defaults.py")
	hostFile = join(myDir, "hosts", hostName + ".py")
	if os.stat(goSettingsFile).st_mtime - max(os.stat(defaultsFile).st_mtime, os.stat(hostFile).st_mtime) >= 0:
		print("Re-using existing settings.go")
		goBuildAndExit(True)

defaultsDict = {
	key: value for key, value in
	defaults.__dict__.items()
	if key.upper() == key
}

settingsDict = defaultsDict.copy()

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
	"keepSettingsGo": False,
}

hostGlobalsCommon = {
	"GoExpr": GoExpr,
	"goGetenv": goGetenv,
	"passwordStore": passwordStore,
	"goSecretCBC": goSecretCBC,
}

hostModulePath = join(myDir, "hosts", hostName + ".py")
if isfile(hostModulePath):
	with open(hostModulePath, encoding="utf-8") as hostFp:
		hostModuleCode = hostFp.read()

	hostGlobals = dict(hostGlobalsCommon)
	exec(hostModuleCode, hostGlobals)
	# exec(object[, globals[, locals]])
	# If only globals is given, locals defaults to it
	for param, value in hostGlobals.items():
		if param in hostGlobalsCommon:
			continue
		if param in hostMetaParams:
			hostMetaParams[param] = value
			continue
		if param.startswith("_"):
			continue
		if param.upper() != param:
			print("skipping non-uppercase parameter %r" % param)
			continue
		if param not in defaultsDict:
			print("skipping unknown parameter %r" % param)
			continue
		valueTypeExpected = type(defaultsDict[param])
		valueTypeActual = type(value)
		if isinstance(value, GoExpr):
			valueTypeActual = value.getPyType()
		if valueTypeActual != valueTypeExpected:
			raise ValueError(
				"invalid type for parameter %r, " % param +
				"must be %s, " % valueTypeExpected.__name__ +
				"not %s" % valueTypeActual.__name__
			)
		settingsDict[param] = value
else:
	print('No settings file found for host %r' % hostName)


for param, value in settingsDict.items():
	if "SECRET" in param and value == "":
		sys.stderr.write(
			"%s can not be empty\n" % param +
			"Set (and export) environment variable STARCAL_%s\n" % param +
			"Or define %s in host settings file\n" % param
		)
		sys.exit(1)

if not settingsDict.get("PASSWORD_SALT"):
	sys.stderr.write(
		"PASSWORD_SALT can not be empty\n" +
		"Set (and export) environment variable STARCAL_PASSWORD_SALT\n" +
		"Or define PASSWORD_SALT in host settings file\n"
	)
	sys.exit(1)


hostOS = settingsDict.pop("OS")
hostArch = settingsDict.pop("ARCH")

#pprint(settingsDict)

constLines = [
	"\tHOST = %s" % json.dumps(hostName),
]
varLines = []
printLines = [
	'\tfmt.Printf("HOST=%#v\\n", HOST)'
]

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

importLines = set(["fmt"])

for param, value in sorted(settingsDict.items()):
	valueType = type(value)
	if valueType in (str, int, float, bool):
		constLines.append("\t%s = %s" % (param, json.dumps(value)))
	elif isinstance(value, GoExpr):
		varLines.append("\t%s = %s" % (param, value.getExpr()))
		importLines.update(set(value.getImports()))
	elif valueType == list:
		itemTypes = set()
		itemValuesGo = []
		if len(value) == 0:
			print("Empty list %s, assuming list of strings" % param)
			itemTypes.add("string")
		else:
			for item in value:
				itemType, itemValueGo = encodeGoValue(item)
				if not itemType:
					print("Unsupported type for %r in list %s" %(item, param))
					sys.exit(1)
				itemTypes.add(itemType)
				itemValuesGo.append(itemValueGo)

		if len(itemTypes) > 1:
			print("List %s has more than one item type: %r" % (param, list(itemTypes)))
			sys.exit(1)

		valueGo = "[]" + itemTypes.pop() + "{" + ", ".join(itemValuesGo) + "}"
		varLines.append("\t%s = %s" % (param, valueGo))
	elif valueType == dict:
		keysValuesGo = {}
		keyTypes = set()
		valueTypes = set()
		if len(value) == 0:
			print("Empty dict %s, assuming generic: map[string]interface{}" % param)
			keyTypes.add("string")
			valueTypes.add("interface{}")
		else:
			for k, v in value.items():
				k_type, k_value = encodeGoValue(k)
				if not k_type:
					print("Unsupported type for key %r in dict %s" %(k, param))
					sys.exit(1)
				v_type, v_value = encodeGoValue(v)
				if not v_type:
					print("Unsupported type for key %r in dict %s" %(v, param))
					sys.exit(1)
				keyTypes.add(k_type)
				valueTypes.add(v_type)
				keysValuesGo[k_value] = v_value

		if len(keyTypes) > 1:
			print("Dict %s has more than one key type: %r" % (param, list(keyTypes)))
			sys.exit(1)
		if len(valueTypes) > 1:
			print("Dict %s has more than one key type: %r" % (param, list(valueTypes)))
			sys.exit(1)
		
		typeGo = "map[%s]%s" % (keyTypes.pop(), valueTypes.pop())
		valueGo = typeGo + "{" + "".join(
			"\n\t\t" + k + ": " + v + ","
			for k, v in keysValuesGo.items()
		) + "\n\t}"
		varLines.append("\t%s = %s" % (param, valueGo))
	else:
		# FIXME
		print(
			"skipping unknown value type %s" % valueType +
			", param %s" % param
		)
		# valueRepr = str(value)
		# varLines.append("\t%s = %s" % (param, valueRepr))
	if "SECRET" in param:
		continue
	if param in secretSettingsParams:
		continue
	printLines.append('\tfmt.Printf("%s=%%#v\\n", %s)' % (param, param))



importBlock = "import (\n" + "\n".join(
	'\t"' + line + '"' for line in importLines
) + "\n)\n"


constBlock = "const (\n" + "\n".join(constLines) + "\n)\n"
varBlock = "var (\n" + "\n".join(varLines) + "\n)\n"
printFunc = "func PrintSettings() {\n%s\n}" % "\n".join(printLines)

#print(constBlock)


if not isdir(goSettingsDir):
	os.mkdir(goSettingsDir)
with open(goSettingsFile, "w") as goFp:
	goFp.write("""// This is an auto-generated code. DO NOT MODIFY
package settings
%s

%s

%s

%s""" % (
	importBlock,
	constBlock,
	varBlock,
	printFunc,
))


if hostOS:
	os.putenv("GOOS", hostOS)
if hostArch:
	os.putenv("GOARCH", hostArch)

if "--no-build" in sys.argv:
	sys.exit(0)

keepSettingsGo = hostMetaParams["keepSettingsGo"] or "--no-remove" in sys.argv
goBuildAndExit(keepSettingsGo)
