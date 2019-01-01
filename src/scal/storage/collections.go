package storage

// MongoDB Collection Names:
const (
	C_user               = "users"
	C_userChangeLog      = "user_change_log"
	C_userLogins         = "user_logins"
	C_group              = "event_group"
	C_eventMeta          = "event_meta"
	C_attending          = "event_attending"
	C_eventMetaChangeLog = "event_meta_change_log"
	C_revision           = "event_revision"
	C_eventData          = "event_data"
	C_invitation         = "event_invitation"
	C_resetPwToken       = "reset_password_token"
	C_resetPwLog         = "reset_password_log"

	C_errorsPrefix = "errors_"
)
