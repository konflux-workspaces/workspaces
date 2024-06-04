package context

type ServerContextKey string

const (
	UserSubKey                 ServerContextKey = "user-sub"
	UserSignupComplaintNameKey ServerContextKey = "usersignup-complaintname"
)
