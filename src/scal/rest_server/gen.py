#!/usr/bin/python3
import os
from os.path import join, dirname, realpath
import string

myDir = dirname(realpath(__file__))

def genEventTypeHandlers():
    with open(join(myDir, "event_handlers.got")) as tplFp:
        tplText = tplFp.read()
    for eventType in (
        "allDayTask",
        "dailyNote",
        "largeScale",
        "lifeTime",
        "monthly",
        "task",
        "universityClass",
        "universityExam",
        "weekly",
        "yearly",
        #"custom",
    ):
        eventTypeCap = eventType[0].upper() + eventType[1:]
        goText = string.Template(tplText).substitute(
            EVENT_TYPE=eventType,
            EVENT_TYPE_CAP=eventTypeCap,
        )
        with open(join(
            myDir,
            "event_handlers_%s.go" % eventType,
        ), "w") as goFp:
            goFp.write(goText)

if __name__=="__main__":
    genEventTypeHandlers()

