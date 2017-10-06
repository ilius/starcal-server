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

EVENT_INVITE_EMAIL_TEMPLATE = """Hi {{.Name}}

You are invited to event "{{.EventModel.Summary}}", by {{.SenderName}} <{{.SenderEmail}}>
Click on this link to join the event:
{{.Host}}/event/{{.EventType}}/{{.EventId}}/join

This invitation Email is sent via StarCalendar by one of the users
Have fun using StarCalendar
"""

CONFIRM_EMAIL_SECRET = os.getenv("STARCAL_CONFIRM_EMAIL_SECRET", "")


CONFIRM_EMAIL_EMAIL_TEMPLATE = """Hi {{.Name}}

Please open this link in your browser to confirm your email address:
{{.ConfirmationURL}}

The link above will be expired on {{.ExpirationTime}}
You also need to open the link with the same IP address as you requested with.

Have fun using StarCalendar
"""