package outputs

import awss3 "cloudcanvas-backend/internal/aws/models/storage/s3"

// BucketACLOutput represents response for bucket ACL operations
type BucketACLOutput struct {
	ID                  string                     `json:"id"` // bucket name
	ACL                 *string                    `json:"acl,omitempty"`
	AccessControlPolicy *awss3.AccessControlPolicy `json:"access_control_policy,omitempty"`
}
