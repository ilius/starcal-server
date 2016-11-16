#!/usr/bin/python3
# DO NOT USE DJANGO OR ANY EXTERNAL LIBRARIES IN THIS SCRIPT

import os
from os.path import join, isfile, isdir, dirname, abspath
import json
import subprocess
from pprint import pprint

import defaults

myDir = dirname(abspath(__file__))
rootDir = dirname(myDir)
srcDir = join(rootDir, "src")


hostName = os.getenv("STARCAL_HOST")
print("Generating settings based on: STARCAL_HOST = %r" % hostName)
#if not hostName:
#    raise ValueError(
#        "Set (and export) environment varibale `STARCAL_HOST` " +
#        "before running this script"
#    )

defaultsDict = {
    key: value for key, value in
    defaults.__dict__.items()
    if key.upper() == key
}

settingsDict = defaultsDict.copy()

if hostName:
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

lines = ["const ("]
for param, value in sorted(settingsDict.items()):
    valueType = type(value)
    if valueType in (str, int, float):
        valueRepr = json.dumps(value)
    else:
        # FIXME
        print("unknown value type %s, not sure how to encode" % valueType)
        valueRepr = str(value)
    lines.append("    %s = %s" % (param, valueRepr))
lines.append(")")

varCode = "\n".join(lines)

#print(varCode)

goSettingsDir = join(srcDir, "scal", "settings")
goSettingsFile = join(goSettingsDir, "settings.go")

if not isdir(goSettingsDir):
    os.mkdir(goSettingsDir)
with open(goSettingsFile, "w") as goFp:
    goFp.write("""// This is an auto-generated code. DO NOT MODIFY
package settings

%s""" % varCode)


if hostOS:
    os.putenv("GOOS", hostOS)
if hostArch:
    os.putenv("GOARCH", hostArch)

os.putenv("GOPATH", rootDir)
subprocess.call([
    "go",
    "build",
    "-o", "server-%s" % hostName,
    "server.go",
])

os.remove(goSettingsFile)


