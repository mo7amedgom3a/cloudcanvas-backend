package resource

import "time"
type CloudProvider string 
const (
	AWS CloudProvider = "aws"
	AZURE CloudProvider = "azure"
	GCP   CloudProvider = "gcp"
)
type ResourceType struct {
	ID         string
	Name       string
	Category   string
	Kind       string
	IsRegional bool
	IsGlobal   bool
}

type Resource struct {
	ID        string
	Name      string
	Type      ResourceType
	Provider  CloudProvider
	Region    string
	ParentID  *string
	DependsOn []string
	Metadata  map[string]interface{} // Additional metadata for pricing and configuration
}

// ResourceOutput represents resource output data after creation/update
// This includes cloud-generated identifiers and metadata
type ResourceOutput struct {
	ID        string
	ARN       *string
	Name      string
	Type      ResourceType
	Provider  CloudProvider
	Region    string
	CreatedAt *time.Time
	ParentID  *string
	DependsOn []string
	Metadata  map[string]interface{} 
}
