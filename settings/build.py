#!/usr/bin/python3
# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import sys
import os
from os.path import join, isfile, isdir, dirname, abspath
import json
import subprocess
from pprint import pprint

import defaults

secretSettingsParams = {
	"MONGO_PASSWORD",
	"JWT_TOKEN_SECRET",
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

defaultsDict = {
	key: value for key, value in
	defaults.__dict__.items()
	if key.upper() == key
}

settingsDict = defaultsDict.copy()

hostModulePath = join(myDir, "hosts", hostName + ".py")
if isfile(hostModulePath):
	with open(hostModulePath, encoding="utf-8") as hostFp:
		hostModuleCode = hostFp.read()
	hostGlobals = {}
	exec(hostModuleCode, hostGlobals)
	# exec(object[, globals[, locals]])
	# If only globals is given, locals defaults to it
	for param, value in hostGlobals.items():
		if param.startswith("_"):
			continue
		if param.upper() != param:
			print("skipping non-uppercase parameter %r" % param)
			continue
		if param not in defaultsDict:
			print("skipping unknown parameter %r" % param)
			continue
		valueType = type(defaultsDict[param])
		if type(value) != valueType:
			raise ValueError(
				"invalid type for parameter %r, " % param +
				"must be %s, " % valueType.__name__ +
				"not %s" % type(value).__name__
			)
		settingsDict[param] = value
else:
	print('No settings file found for host %r' % hostName)


hostOS = settingsDict.pop("OS")
hostArch = settingsDict.pop("ARCH")

#pprint(settingsDict)

constLines = [
	"\tHOST = %s" % json.dumps(hostName),
]
printLines = [
	'\tfmt.Printf("HOST=%#v\\n", HOST)'
]
for param, value in sorted(settingsDict.items()):
	valueType = type(value)
	if valueType in (str, int, float, bool):
		constLines.append("\t%s = %s" % (param, json.dumps(value)))
	else:
		# FIXME
		print(
			"skipping unknown (non-const) value type %s" % valueType +
			", param %s" % param
		)
		# valueRepr = str(value)
		# varLines.append("\t%s = %s" % (param, valueRepr))
	if param not in secretSettingsParams:
		printLines.append('\tfmt.Printf("%s=%%#v\\n", %s)' % (param, param))


constBlock = "const (\n" + "\n".join(constLines) + "\n)\n"
printFunc = "func PrintSettings() {\n%s\n}" % "\n".join(printLines)

#print(constBlock)

goSettingsDir = join(srcDir, "scal", "settings")
goSettingsFile = join(goSettingsDir, "settings.go")

if not isdir(goSettingsDir):
	os.mkdir(goSettingsDir)
with open(goSettingsFile, "w") as goFp:
	goFp.write("""// This is an auto-generated code. DO NOT MODIFY
package settings
import "fmt"

%s

%s""" % (
	constBlock,
	printFunc,
))


if hostOS:
	os.putenv("GOOS", hostOS)
if hostArch:
	os.putenv("GOARCH", hostArch)

os.putenv("GOPATH", rootDir)
status = subprocess.call([
	"go",
	"build",
	"-o", "server-%s" % hostName,
	"server.go",
])

if "--no-remove" not in sys.argv:
	os.remove(goSettingsFile)

sys.exit(status)
