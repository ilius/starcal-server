package storage

// MongoDB Collection Names:
const (
	C_user       = "users"       // UserModel
	C_userLogins = "user_logins" // UserLoginAttemptModel
	C_group      = "event_group" // EventGroupModel

	C_eventMeta = "event_meta"     // EventMetaModel
	C_revision  = "event_revision" // EventRevisionModel
	C_eventData = "event_data"     // BaseEventModel

	C_attending    = "event_attending"      // EventAttendingModel
	C_resetPwToken = "reset_password_token" // ResetPasswordTokenModel

	C_invitation         = "event_invitation"      // Insert-only, EventInvitationModel
	C_userChangeLog      = "user_change_log"       // Insert-only, UserChangeLogModel
	C_eventMetaChangeLog = "event_meta_change_log" // Insert-only, EventMetaChangeLogModel
	C_resetPwLog         = "reset_password_log"    // Insert-only, ResetPasswordLogModel

	C_errorsPrefix = "errors_"
)
