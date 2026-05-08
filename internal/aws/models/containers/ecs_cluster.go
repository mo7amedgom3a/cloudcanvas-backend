package containers

import "cloudcanvas-backend/internal/aws/configs"

// ECSCluster represents an AWS ECS Cluster
type ECSCluster struct {
	Name                     string        `json:"name"`
	ContainerInsightsEnabled bool          `json:"container_insights_enabled,omitempty"`
	ExecuteCommandEnabled    bool          `json:"execute_command_enabled,omitempty"`
	KMSKeyID                 string        `json:"kms_key_id,omitempty"` // For execute command encryption
	LogGroup                 string        `json:"log_group,omitempty"`  // For execute command logging
	Tags                     []configs.Tag `json:"tags,omitempty"`
}
