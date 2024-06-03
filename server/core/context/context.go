package context

type ServerContextKey string

const (
	UserSubKey      ServerContextKey = "user-sub"
	UserUsernameKey ServerContextKey = "user-username"
)
