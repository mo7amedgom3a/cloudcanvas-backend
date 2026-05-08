package resource

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"cloudcanvas-backend/internal/aws/validation"
	awsnetworking "cloudcanvas-backend/internal/aws/models/networking"
	"cloudcanvas-backend/internal/models"
)

func TestToAWSModel(t *testing.T) {
	// Create a mapper resource with metadata containing VPC configuration
	metadata := map[string]interface{}{
		"name":                 "test-vpc",
		"region":               "us-east-1",
		"cidr":                 "10.0.0.0/16",
		"enable_dns_support":   true,
		"enable_dns_hostnames": true,
	}

	res := &Resource{
		ID:       "vpc-1",
		Name:     "test-vpc",
		Provider: "aws",
		Region:   "us-east-1",
		Metadata: metadata,
	}

	// Map to AWS networking.VPC
	var awsVPC awsnetworking.VPC
	err := res.ToAWSModel(&awsVPC)
	if err != nil {
		t.Fatalf("Failed to map to AWS model: %v", err)
	}

	// Validate fields
	if awsVPC.Name != "test-vpc" {
		t.Errorf("Expected Name 'test-vpc', got '%s'", awsVPC.Name)
	}
	if awsVPC.Region != "us-east-1" {
		t.Errorf("Expected Region 'us-east-1', got '%s'", awsVPC.Region)
	}
	if awsVPC.CIDR != "10.0.0.0/16" {
		t.Errorf("Expected CIDR '10.0.0.0/16', got '%s'", awsVPC.CIDR)
	}
	if !awsVPC.EnableDNSSupport {
		t.Error("Expected EnableDNSSupport to be true")
	}
	if !awsVPC.EnableDNSHostnames {
		t.Error("Expected EnableDNSHostnames to be true")
	}
}

func TestFromAWSModel(t *testing.T) {
	// Create an AWS networking.VPC model
	awsVPC := &awsnetworking.VPC{
		Name:               "test-vpc-2",
		Region:             "us-west-2",
		CIDR:               "172.16.0.0/16",
		EnableDNSSupport:   true,
		EnableDNSHostnames: false,
	}

	res := &Resource{
		ID:   "vpc-2",
		Name: "test-vpc-2",
	}

	err := res.FromAWSModel(awsVPC)
	if err != nil {
		t.Fatalf("Failed to load AWS model: %v", err)
	}

	// Validate metadata map
	if res.Metadata["name"] != "test-vpc-2" {
		t.Errorf("Expected metadata name 'test-vpc-2', got '%v'", res.Metadata["name"])
	}
	if res.Metadata["region"] != "us-west-2" {
		t.Errorf("Expected metadata region 'us-west-2', got '%v'", res.Metadata["region"])
	}
	if res.Metadata["cidr"] != "172.16.0.0/16" {
		t.Errorf("Expected metadata cidr '172.16.0.0/16', got '%v'", res.Metadata["cidr"])
	}
	if res.Metadata["enable_dns_support"] != true {
		t.Errorf("Expected metadata enable_dns_support true, got '%v'", res.Metadata["enable_dns_support"])
	}
	if res.Metadata["enable_dns_hostnames"] != false {
		t.Errorf("Expected metadata enable_dns_hostnames false, got '%v'", res.Metadata["enable_dns_hostnames"])
	}
}

func TestToDBResource(t *testing.T) {
	projectID := uuid.New()
	resourceTypeID := uint(42)

	metadata := map[string]interface{}{
		"cidr": "10.0.0.0/16",
	}

	res := &Resource{
		ID:       "vpc-uuid-style-or-not",
		Name:     "my-db-vpc",
		Metadata: metadata,
	}

	dbRes, err := ToDBResource(res, projectID, resourceTypeID)
	if err != nil {
		t.Fatalf("ToDBResource failed: %v", err)
	}

	if dbRes.OriginalID != "vpc-uuid-style-or-not" {
		t.Errorf("Expected OriginalID 'vpc-uuid-style-or-not', got '%s'", dbRes.OriginalID)
	}
	if dbRes.ProjectID != projectID {
		t.Errorf("Expected ProjectID %s, got %s", projectID, dbRes.ProjectID)
	}
	if dbRes.ResourceTypeID != resourceTypeID {
		t.Errorf("Expected ResourceTypeID %d, got %d", resourceTypeID, dbRes.ResourceTypeID)
	}
	if dbRes.Name != "my-db-vpc" {
		t.Errorf("Expected Name 'my-db-vpc', got '%s'", dbRes.Name)
	}

	var parsedConfig map[string]interface{}
	err = json.Unmarshal(dbRes.Config, &parsedConfig)
	if err != nil {
		t.Fatalf("Failed to parse GORM Config JSON: %v", err)
	}
	if parsedConfig["cidr"] != "10.0.0.0/16" {
		t.Errorf("Expected config cidr '10.0.0.0/16', got '%v'", parsedConfig["cidr"])
	}
}

func TestToDomainResource(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	depID := uuid.New()

	dbRes := &models.Resource{
		ID:             childID,
		OriginalID:     "subnet-1",
		ProjectID:      uuid.New(),
		ResourceTypeID: 5,
		Name:           "my-subnet",
		Config:         datatypes.JSON([]byte(`{"cidr":"10.0.1.0/24", "region":"us-east-1"}`)),
		ResourceType: models.ResourceType{
			ID:            5,
			Name:          "subnet",
			CloudProvider: "aws",
			IsRegional:    true,
			IsGlobal:      false,
			Category: &models.ResourceCategory{
				Name: "Networking",
			},
			Kind: &models.ResourceKind{
				Name: "Subnet",
			},
		},
		ParentResources: []models.ResourceContainment{
			{
				ParentResourceID: parentID,
				ChildResourceID:  childID,
				ParentResource: models.Resource{
					ID:         parentID,
					OriginalID: "vpc-1",
					Name:       "my-vpc",
				},
			},
		},
		FromDependencies: []models.ResourceDependency{
			{
				FromResourceID: childID,
				ToResourceID:   depID,
				ToResource: models.Resource{
					ID:         depID,
					OriginalID: "igw-1",
					Name:       "my-igw",
				},
			},
		},
	}

	domainRes, err := ToDomainResource(dbRes)
	if err != nil {
		t.Fatalf("ToDomainResource failed: %v", err)
	}

	if domainRes.ID != "subnet-1" {
		t.Errorf("Expected ID 'subnet-1', got '%s'", domainRes.ID)
	}
	if domainRes.Name != "my-subnet" {
		t.Errorf("Expected Name 'my-subnet', got '%s'", domainRes.Name)
	}
	if domainRes.Provider != "aws" {
		t.Errorf("Expected Provider 'aws', got '%s'", domainRes.Provider)
	}
	if domainRes.Region != "us-east-1" {
		t.Errorf("Expected Region 'us-east-1', got '%s'", domainRes.Region)
	}
	if domainRes.Type.Name != "subnet" {
		t.Errorf("Expected Type.Name 'subnet', got '%s'", domainRes.Type.Name)
	}
	if domainRes.Type.Category != "Networking" {
		t.Errorf("Expected Type.Category 'Networking', got '%s'", domainRes.Type.Category)
	}
	if domainRes.Type.Kind != "Subnet" {
		t.Errorf("Expected Type.Kind 'Subnet', got '%s'", domainRes.Type.Kind)
	}
	if !domainRes.Type.IsRegional || domainRes.Type.IsGlobal {
		t.Error("Expected Type IsRegional=true and IsGlobal=false")
	}

	if domainRes.ParentID == nil || *domainRes.ParentID != "vpc-1" {
		t.Errorf("Expected ParentID 'vpc-1', got '%v'", domainRes.ParentID)
	}

	if len(domainRes.DependsOn) != 1 || domainRes.DependsOn[0] != "igw-1" {
		t.Errorf("Expected DependsOn ['igw-1'], got '%v'", domainRes.DependsOn)
	}

	if domainRes.Metadata["cidr"] != "10.0.1.0/24" {
		t.Errorf("Expected metadata cidr '10.0.1.0/24', got '%v'", domainRes.Metadata["cidr"])
	}
}

func TestMapperWithValidation(t *testing.T) {
	// Create a mapper resource representing an S3 bucket
	res := &Resource{
		ID:       "s3-1",
		Name:     "my-s3-bucket",
		Provider: "aws",
		Type: ResourceType{
			Name: "s3_bucket",
		},
		Metadata: map[string]interface{}{
			"bucket":        "my-valid-s3-bucket-name",
			"force_destroy": true,
		},
	}

	// Dynamically validate the resource metadata config based on its type
	err := validation.Validate(res.Type.Name, res.Metadata)
	if err != nil {
		t.Fatalf("Expected valid S3 bucket configuration to pass validation, got: %v", err)
	}

	// Create an invalid S3 bucket config
	invalidRes := &Resource{
		ID:       "s3-2",
		Name:     "my-invalid-s3-bucket",
		Provider: "aws",
		Type: ResourceType{
			Name: "s3_bucket",
		},
		Metadata: map[string]interface{}{
			"bucket": "My Invalid S3 Bucket", // Spaces and capital letters are invalid in S3 names
		},
	}

	err = validation.Validate(invalidRes.Type.Name, invalidRes.Metadata)
	if err == nil {
		t.Fatal("Expected validation errors for S3 bucket with invalid name")
	}

	if !strings.Contains(err.Error(), "invalid bucket name") {
		t.Errorf("Expected S3 bucket naming format error, got: %v", err)
	}
}
