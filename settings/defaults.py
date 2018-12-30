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
JWT_ALG = "HS256"
PASSWORD_SALT = os.getenv("STARCAL_PASSWORD_SALT", "")

RESET_PASSWORD_TOKEN_LENGTH = 32
RESET_PASSWORD_EXP_SECONDS = 30 * 60
RESET_PASSWORD_REJECT_SECONDS = 60
RESET_PASSWORD_TOKEN_EMAIL_TEMPLATE = """You or someone else has requested a password reset for your StarCalendar account

If it was you, you can use the following token to reset your password:
Reset Password Token: {{.Token}}
This token will be expired at {{.ExpireTime}}

If it wasn't you, you can ignore this email.
But you should know that this request was sent from this IP: {{.IssueRemoteIp}}

Have fun using StarCalendar
"""

RESET_PASSWORD_DONE_EMAIL_TEMPLATE = """Hi {{.Name}}

Your StarCalendar password has been reset by this IP: {{.RemoteIp}}

Have fun using StarCalendar
"""

ALLOW_MISMATCH_EVENT_TYPE = False

EVENT_INVITE_SECRET = os.getenv("STARCAL_EVENT_INVITE_SECRET", "")
EVENT_INVITE_TOKEN_EXP_SECONDS = 7 * 24 * 3600
EVENT_INVITE_TOKEN_ALG = "HS256"

EVENT_INVITE_EMAIL_TEMPLATE = """Hi {{.Name}}

You are invited to event "{{.EventModel.Summary}}", by {{.SenderName}} <{{.SenderEmail}}>
Click on this link to join the event:
{{.Host}}/event/{{.EventType}}/{{.EventId}}/join?token={{.TokenEscaped}}

This invitation Email is sent via StarCalendar by one of the users
Have fun using StarCalendar
"""

ALLOW_REJOIN_EVENT = False

CONFIRM_EMAIL_SECRET = os.getenv("STARCAL_CONFIRM_EMAIL_SECRET", "")


CONFIRM_EMAIL_EMAIL_TEMPLATE = """Hi {{.Name}}

Please open this link in your browser to confirm your email address:
{{.ConfirmationURL}}

The link above will be expired on {{.ExpirationTime}}
You also need to open the link with the same IP address as you requested with.

Have fun using StarCalendar
"""

ADMIN_EMAILS = [""] # type: List[str]

STORE_FAILED_LOGINS = True
STORE_SUCCESSFUL_LOGINS = True
STORE_LOCKED_SUCCESSFUL_LOGINS = True

USER_INFO_LAST_LOGINS_LIMIT = 5

USER_LOGIN_HISTORY_DEFAULT_LIMIT = 20

API_PAGE_LIMIT_DEFAULT = 50
API_PAGE_LIMITS = {
	"api_v1.GetGroupEventList": 150, # ~60 bytes -> eventId, eventType
	"api_v1.GetGroupEventListWithSha1": 100, # ~115 bytes -> eventId, eventType, lastSha1
	"api_v1.GetGroupModifiedEvents": 20, # ~550 bytes -> Full Event
	"api_v1.GetGroupMovedEvents": 60, # ~160 bytes -> eventId, oldGroupId, newGroupId, time

	"api_v1.GetUngroupedEvents": 150, # ~60 bytes -> eventId, eventType
	"api_v1.GetMyEventList": 150, # ~60 bytes -> eventId, eventType
	"api_v1.GetMyEventsFull": 20, # ~550 bytes -> Full Event

	"api_v1.GetGroupList": 100, # ~100 bytes -> groupId, ownerEmail, title

	"api_v1.GetUserLoginHistory": 70, # ~150 bytes
}

# 5.0 means that user can specify the page limit up to 5 times the default limit
API_PAGE_LIMIT_MAX_RATIO = 5.0

# Code	Test	Doc		File	Method
# [x]	[x]		[x]		group_handlers.go	GetGroupEventList
# [x]	[x]		[x]		group_handlers.go	GetGroupEventListWithSha1
# [x]	[x]		[x]		group_handlers.go	GetGroupModifiedEvents		(just "limit" param)
# [x]	[x]		[x]		group_handlers.go	GetGroupMovedEvents			(just "limit" param)
# [x]	[x]		[-]		handlers.go			GetUngroupedEvents
# [x]	[x]		[x]		handlers.go			GetMyEventList
# [x]	[x]		[x]		handlers.go			GetMyEventsFull
# [ ]	[ ]		[ ]		GetGroupList
# [ ]	[ ]		[ ]		GetUserLoginHistory