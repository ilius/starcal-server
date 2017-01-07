#!/usr/bin/python3
import os

OS = "" # "linux", "freebsd", "darwin", "android", ...
ARCH = "" # "amd64", "386", "arm64", "arm", ...

MONGO_HOST = "127.0.0.1:27017"
MONGO_DB_NAME = "starcal"
MONGO_USERNAME = ""
MONGO_PASSWORD = ""
JWT_TOKEN_SECRET = os.getenv("STARCAL_JWT_TOKEN_SECRET", "")
JWT_TOKEN_EXP_SECONDS = 7 * 24 * 3600
PASSWORD_SALT = "starcal"
ALLOW_MISMATCH_EVENT_TYPE = False

EVENT_INVITE_EMAIL_TEMPLATE = """Hi {{.Name}}

You are invited to event "{{.EventModel.Summary}}", by {{.SenderName}} <{{.SenderEmail}}>
Click on this link to join the event:
{{.JoinURL}}

This invitation Email is sent via StarCalendar by one of the users
Have fun using StarCalendar
"""
