package idls

type CreateShareRequest struct {
	ChartID   int     `json:"chart_id" binding:"required"`
	Password  *string `json:"password,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type ShareResponse struct {
	ID        int     `json:"id"`
	Token     string  `json:"token"`
	ChartID   int     `json:"chart_id"`
	Password  *string `json:"password,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

type ShareListResponse struct {
	Items []ShareResponse `json:"items"`
	Total int64           `json:"total"`
}

type VerifyPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}
