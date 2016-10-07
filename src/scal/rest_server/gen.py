#!/usr/bin/python3
import os
from os.path import join, dirname, realpath
import re

import django
from django.template import Template, loader, Context
from django.conf import settings


myDir = dirname(realpath(__file__))
myParentDir = dirname(myDir)
templatesDir = join(myParentDir, 'templates')


activeEventTypes = (
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
)


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
    baseParams = extractEventBaseParams()
    for eventType in activeEventTypes:
        eventTypeCap = eventType[0].upper() + eventType[1:]
        typeParams = extractEventTypeParams(eventType)
        goText = tpl.render(Context(dict(
            EVENT_TYPE=eventType,
            EVENT_TYPE_CAP=eventTypeCap, # instead of {{EVENT_TYPE|capfirst}}
            #EVENT_BASE_PARAMS=baseParams,
            #EVENT_TYPE_PARAMS=typeParams,
            EVENT_PARAMS=baseParams + typeParams,
        )))
        with open(join(
            myDir,
            "event_handlers_%s.go" % eventType,
        ), "w") as goFp:
            goFp.write(goText)


def parseModelVarLine(line):
    """
    return (param, _type)
    or raise ValueError
    """
    try:
        _type, param = re.findall(' ([^\\s]*?)\\s*?`.*json:"(.*?)"', line)[0]
    except IndexError:
        raise ValueError
    if not param:
        raise ValueError
    if param == '-':
        raise ValueError
    param, _, opt = param.partition(',')
    assert opt in ('', 'omitempty')
    return (param, _type)


def extractEventTypeParams(eventType):
    params = []
    eventTypeCap = eventType[0].upper() + eventType[1:]
    with open(join(myParentDir, 'event_lib/%s.go' % eventType)) as goFp:
        text = goFp.read()
    for line in re.findall(
        'type %sEventModel.*?{.*?}' % eventTypeCap,
        text,
        re.S,
    )[0].split('\n')[2:-1]:
        try:
            params.append(parseModelVarLine(line))
        except ValueError:
            pass
    return params


def extractEventBaseParams():
    params = []
    with open(join(myParentDir, 'event_lib/base.go')) as goFp:
        text = goFp.read()
    for line in re.findall(
        'type BaseEventModel.*?{.*?}',
        text,
        re.S,
    )[0].split('\n')[2:-1]:
        try:
            param, _type = parseModelVarLine(line)
        except ValueError:
            continue
        if param == 'sha1':
            continue
        params.append((param, _type))
    return params


def testExtractEventTypeParams():
    print('---- base:', extractEventBaseParams())
    for eventType in activeEventTypes:
        print(eventType, extractEventTypeParams(eventType))

if __name__=="__main__":
    genEventTypeHandlers()
    #testExtractEventTypeParams()
