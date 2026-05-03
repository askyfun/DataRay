package entity

// Share represents a shared chart link
type Share struct {
	ID        int     `json:"id"`
	Token     string  `json:"token"`
	ChartID   int     `json:"chart_id"`
	Password  *string `json:"password"`
	ExpiresAt *string `json:"expires_at"`
	CreatedAt string  `json:"created_at"`
}

// ShareService defines operations for share management
type ShareService interface {
	// Create a new share link
	Create(chartID int, password *string, expiresAt *string) (*Share, error)

	// Get share by token
	GetByToken(token string) (*Share, error)

	// Validate password if required
	ValidatePassword(token, password string) error
}
