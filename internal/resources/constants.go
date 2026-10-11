package resources

import "time"

const (
	AccessTTL       = 15 * time.Minute
	RefreshTTL      = 720 * time.Hour
	UniqueViolation = "23505"
	UserIDKey       = "user_id"
)
