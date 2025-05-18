package contextkey

type ContextKey string

const (
	UserID ContextKey = "user_id"
	Role   ContextKey = "role"
)
