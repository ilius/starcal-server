#!/usr/bin/python3
# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import sys
import os
from os.path import join, isfile, isdir, dirname, abspath
import json
from pprint import pprint

from build_common import *

interactive = "--interactive" in sys.argv

hostFile = join(myDir, "hosts", hostName + ".py")

if not isfile(hostFile):
	print(f"Creating new host file: {hostFile}")
	with open(hostFile, "w", encoding="utf8") as hostFp:
		hostFp.write("")

if isfile(goSettingsFile):
	defaultsFile = join(myDir, "defaults.py")
	configLastModified = max(
		os.stat(defaultsFile).st_mtime,
		os.stat(hostFile).st_mtime,
	)
	if os.stat(goSettingsFile).st_mtime >= configLastModified:
		print("Re-using existing settings.go")
		goBuildAndExit(True)


defaultsDict = {
	key: value for key, value in
	defaults.__dict__.items()
	if key.upper() == key
}

settingsDict = defaultsDict.copy()


with open(hostFile, encoding="utf-8") as hostFp:
	hostCode = hostFp.read()

hostGlobals = dict(hostGlobalsCommon)
exec(hostCode, hostGlobals)
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
		print(f"skipping non-uppercase parameter {param!r}")
		continue
	if param not in defaultsDict:
		print(f"skipping unknown parameter {param!r}")
		continue
	valueTypeExpected = type(defaultsDict[param])
	valueTypeActual = type(value)
	if isinstance(value, GoExpr):
		valueTypeActual = value.getPyType()
	if valueTypeActual != valueTypeExpected:
		raise ValueError(
			f"invalid type for parameter {param!r}, " +
			f"must be {valueTypeExpected.__name__}, " +
			f"not {valueTypeActual.__name__}"
		)
	settingsDict[param] = value


hostNewLines = []


def askForParam(param: str) -> None:
	value = prompt(f"{param} = ")
	settingsDict[param] = value
	hostNewLines.append(f"{param} = {value!r}")


for param, value in settingsDict.items():
	if "SECRET" in param and value == "":
		if interactive:
			askForParam(param)
		else:
			sys.stderr.write(
				f"{param} can not be empty\n" +
				f"Set (and export) environment variable STARCAL_{param}\n" +
				f"Or define {param} in host settings file\n"
			)
			sys.exit(1)

if not settingsDict.get("PASSWORD_SALT"):
	if interactive:
		askForParam("PASSWORD_SALT")
	else:
		sys.stderr.write(
			"PASSWORD_SALT can not be empty\n" +
			"Set (and export) environment variable STARCAL_PASSWORD_SALT\n" +
			"Or define PASSWORD_SALT in host settings file\n"
		)
		sys.exit(1)

if hostNewLines:
	hostCode += "\n\n" + "\n".join(hostNewLines) + "\n"
	print(f"Writing file: {hostFile}")
	with open(hostFile, "w", encoding="utf-8") as hostFp:
		hostFp.write(hostCode)


hostOS = settingsDict.pop("OS")
hostArch = settingsDict.pop("ARCH")

#pprint(settingsDict)

constLines = [
	"\tHOST = " + json.dumps(hostName),
]
varLines = []
printLines = [
	'\tfmt.Printf("HOST=%#v\\n", HOST)'
]

importLines = set(["fmt"])

for param, value in sorted(settingsDict.items()):
	valueType = type(value)
	if valueType in (str, int, float, bool):
		constLines.append(f"\t{param} = {json.dumps(value)}")
	elif isinstance(value, GoExpr):
		varLines.append(f"\t{param} = {value.getExpr()}")
		importLines.update(set(value.getImports()))
	elif valueType == list:
		itemTypes = set()
		itemValuesGo = []
		if len(value) == 0:
			print(f"Empty list {param}, assuming list of strings")
			itemTypes.add("string")
		else:
			for item in value:
				itemType, itemValueGo = encodeGoValue(item)
				if not itemType:
					print(f"Unsupported type for {item!r} in list {param}")
					sys.exit(1)
				itemTypes.add(itemType)
				itemValuesGo.append(itemValueGo)

		if len(itemTypes) > 1:
			print(f"List {param} has more than one item type: {list(itemTypes)!r}")
			sys.exit(1)

		valueGo = "[]" + itemTypes.pop() + "{" + ", ".join(itemValuesGo) + "}"
		varLines.append(f"\t{param} = {valueGo}")
	elif valueType == dict:
		keysValuesGo = {}
		keyTypes = set()
		valueTypes = set()
		if len(value) == 0:
			print(f"Empty dict {param}, assuming generic: map[string]interface{{}}")
			keyTypes.add("string")
			valueTypes.add("interface{}")
		else:
			for k, v in value.items():
				k_type, k_value = encodeGoValue(k)
				if not k_type:
					print(f"Unsupported type for key {k!r} in dict {param}")
					sys.exit(1)
				v_type, v_value = encodeGoValue(v)
				if not v_type:
					print(f"Unsupported type for key {v!r} in dict {param}")
					sys.exit(1)
				keyTypes.add(k_type)
				valueTypes.add(v_type)
				keysValuesGo[k_value] = v_value

		if len(keyTypes) > 1:
			print(f"Dict {param} has more than one key type: {list(keyTypes)!r}")
			sys.exit(1)
		if len(valueTypes) > 1:
			print(f"Dict {param} has more than one key type: {list(valueTypes)!r}")
			sys.exit(1)

		keyType = keyTypes.pop()
		valueType = valueTypes.pop()
		typeGo = f"map[{keyType}]{valueType}"
		valueGo = typeGo + "{" + "".join(
			"\n\t\t" + k + ": " + v + ","
			for k, v in keysValuesGo.items()
		) + "\n\t}"
		varLines.append(f"\t{param} = {valueGo}")
	else:
		# FIXME
		print(
			f"skipping unknown value type {valueType}" +
			f", param {param}"
		)
		# valueRepr = str(value)
		# varLines.append(f"\t{param} = {valueRepr}")
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
printFunc = "func PrintSettings() {\n" + "\n".join(printLines) + "\n}"

#print(constBlock)


if not isdir(goSettingsDir):
	os.mkdir(goSettingsDir)
with open(goSettingsFile, "w") as goFp:
	goFp.write(f"""// This is an auto-generated code. DO NOT MODIFY
package settings
{importBlock}

{constBlock}

{varBlock}

{printFunc}""")


if hostOS:
	os.putenv("GOOS", hostOS)
if hostArch:
	os.putenv("GOARCH", hostArch)

if "--no-build" in sys.argv:
	sys.exit(0)

keepSettingsGo = hostMetaParams["KEEP_SETTINGS_GO"]
if "--no-remove" in sys.argv:
	keepSettingsGo = True

goBuildAndExit(keepSettingsGo)
