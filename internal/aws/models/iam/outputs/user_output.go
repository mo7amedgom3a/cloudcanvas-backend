package outputs

import (
	"time"

	"cloudcanvas-backend/internal/aws/configs"
)

// UserOutput represents AWS IAM user output/response data after creation
type UserOutput struct {
	// AWS-generated identifiers
	ARN      string `json:"arn"` // e.g., "arn:aws:iam::123456789012:user/my-user"
	ID       string `json:"id"`  // Same as name for users
	Name     string `json:"name"`
	UniqueID string `json:"unique_id"` // Stable unique identifier

	// Configuration (from input)
	Path                string  `json:"path"`
	PermissionsBoundary *string `json:"permissions_boundary,omitempty"`

	// AWS-specific output fields
	CreateDate       time.Time     `json:"create_date"`
	PasswordLastUsed *time.Time    `json:"password_last_used,omitempty"`
	Tags             []configs.Tag `json:"tags,omitempty"`
	IsVirtual        bool          `json:"is_virtual,omitempty"`
}
