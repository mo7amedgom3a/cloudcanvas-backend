package resource

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"cloudcanvas-backend/internal/models"
)

// ToDomainResource converts a database models.Resource GORM model into a mapper resource.Resource.
// It maps database fields and extracts relations (such as ParentID and DependsOn) if they are preloaded.
func ToDomainResource(dbRes *models.Resource) (*Resource, error) {
	if dbRes == nil {
		return nil, nil
	}

	// Unmarshal Config JSON into Metadata map
	var metadata map[string]interface{}
	if len(dbRes.Config) > 0 {
		if err := json.Unmarshal(dbRes.Config, &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse resource config: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Determine region from metadata if present
	region := ""
	if r, ok := metadata["region"].(string); ok {
		region = r
	} else if r, ok := metadata["Region"].(string); ok {
		region = r
	}

	// Map ResourceType
	resType := ResourceType{
		ID:         fmt.Sprintf("%d", dbRes.ResourceTypeID),
		Name:       dbRes.ResourceType.Name,
		IsRegional: dbRes.ResourceType.IsRegional,
		IsGlobal:   dbRes.ResourceType.IsGlobal,
	}
	if dbRes.ResourceType.Category != nil {
		resType.Category = dbRes.ResourceType.Category.Name
	}
	if dbRes.ResourceType.Kind != nil {
		resType.Kind = dbRes.ResourceType.Kind.Name
	}

	// Identify Parent ID if preloaded
	var parentID *string
	if len(dbRes.ParentResources) > 0 {
		pID := dbRes.ParentResources[0].ParentResource.OriginalID
		if pID == "" {
			pID = dbRes.ParentResources[0].ParentResourceID.String()
		}
		parentID = &pID
	}

	// Identify dependencies if preloaded (this resource as From, depending on To)
	var dependsOn []string
	if len(dbRes.FromDependencies) > 0 {
		dependsOn = make([]string, 0, len(dbRes.FromDependencies))
		for _, dep := range dbRes.FromDependencies {
			depID := dep.ToResource.OriginalID
			if depID == "" {
				depID = dep.ToResourceID.String()
			}
			dependsOn = append(dependsOn, depID)
		}
	}

	id := dbRes.OriginalID
	if id == "" {
		id = dbRes.ID.String()
	}

	return &Resource{
		ID:        id,
		Name:      dbRes.Name,
		Type:      resType,
		Provider:  CloudProvider(dbRes.ResourceType.CloudProvider),
		Region:    region,
		ParentID:  parentID,
		DependsOn: dependsOn,
		Metadata:  metadata,
	}, nil
}

// ToDBResource converts a mapper resource.Resource into a database models.Resource GORM model.
// If the Resource's ID is a valid UUID, it is parsed and used as the primary key. Otherwise, a new UUID is generated.
func ToDBResource(domainRes *Resource, projectID uuid.UUID, resourceTypeID uint) (*models.Resource, error) {
	if domainRes == nil {
		return nil, nil
	}

	var resID uuid.UUID
	var err error
	if domainRes.ID != "" {
		resID, err = uuid.Parse(domainRes.ID)
		if err != nil {
			// Not a valid UUID (e.g. frontend node ID like "vpc-1"), generate a new random UUID
			resID = uuid.New()
		}
	} else {
		resID = uuid.New()
	}

	// Marshal Metadata map into Config JSON
	var configJSON datatypes.JSON
	if domainRes.Metadata != nil {
		jsonData, err := json.Marshal(domainRes.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata to config: %w", err)
		}
		configJSON = datatypes.JSON(jsonData)
	} else {
		configJSON = datatypes.JSON([]byte("{}"))
	}

	return &models.Resource{
		ID:             resID,
		OriginalID:     domainRes.ID,
		ProjectID:      projectID,
		ResourceTypeID: resourceTypeID,
		Name:           domainRes.Name,
		IsVisualOnly:   false,
		Config:         configJSON,
	}, nil
}

// ToAWSModel maps the Resource's Metadata into an AWS-specific model struct (e.g. networking.VPC).
func (r *Resource) ToAWSModel(target interface{}) error {
	if r.Metadata == nil {
		return nil
	}
	jsonData, err := json.Marshal(r.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal resource metadata: %w", err)
	}
	return json.Unmarshal(jsonData, target)
}

// FromAWSModel serializes an AWS-specific model struct and populates the Resource's Metadata.
func (r *Resource) FromAWSModel(awsModel interface{}) error {
	jsonData, err := json.Marshal(awsModel)
	if err != nil {
		return fmt.Errorf("failed to marshal AWS model: %w", err)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(jsonData, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal into metadata map: %w", err)
	}
	r.Metadata = metadata
	return nil
}

// ToAWSModel maps the ResourceOutput's Metadata into an AWS-specific model struct.
func (ro *ResourceOutput) ToAWSModel(target interface{}) error {
	if ro.Metadata == nil {
		return nil
	}
	jsonData, err := json.Marshal(ro.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal resource output metadata: %w", err)
	}
	return json.Unmarshal(jsonData, target)
}

// FromAWSModel serializes an AWS-specific model struct and populates the ResourceOutput's Metadata.
func (ro *ResourceOutput) FromAWSModel(awsModel interface{}) error {
	jsonData, err := json.Marshal(awsModel)
	if err != nil {
		return fmt.Errorf("failed to marshal AWS model: %w", err)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(jsonData, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal into metadata map: %w", err)
	}
	ro.Metadata = metadata
	return nil
}
