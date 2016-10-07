#!/usr/bin/python3
import os
from os.path import join, dirname, realpath

import django
from django.template import Template, loader, Context
from django.conf import settings


myDir = dirname(realpath(__file__))
myParentDir = dirname(myDir)
templatesDir = join(myParentDir, 'templates')


def djangoInit():
    settings.configure(
        TEMPLATES = [
            {
                'BACKEND': 'django.template.backends.django.DjangoTemplates',
                'DIRS': [templatesDir],
                'APP_DIRS': False,
            }
        ]
    )
    django.setup()


def genEventTypeHandlers():
    djangoInit()
    tpl = loader.get_template('event_handlers.got')
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
        goText = tpl.render(Context(dict(
            EVENT_TYPE=eventType,
            EVENT_TYPE_CAP=eventTypeCap,
        )))
        with open(join(
            myDir,
            "event_handlers_%s.go" % eventType,
        ), "w") as goFp:
            goFp.write(goText)

if __name__=="__main__":
    genEventTypeHandlers()

