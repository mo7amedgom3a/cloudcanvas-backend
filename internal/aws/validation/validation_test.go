package validation

import (
	"strings"
	"testing"
)

func TestValidate_VPC(t *testing.T) {
	// 1. Valid VPC config
	validVPC := map[string]interface{}{
		"name":                 "my-test-vpc",
		"region":               "us-east-1",
		"cidr":                 "10.0.0.0/16",
		"enable_dns_support":   true,
		"enable_dns_hostnames": true,
		"instance_tenancy":     "default",
	}

	err := Validate("vpc", validVPC)
	if err != nil {
		t.Fatalf("Expected valid VPC config to pass, got error: %v", err)
	}

	// 2. Missing required field
	missingReq := map[string]interface{}{
		"name": "my-test-vpc",
		"cidr": "10.0.0.0/16",
		// missing "region"
	}
	err = Validate("vpc", missingReq)
	if err == nil {
		t.Error("Expected validation error for missing required 'region'")
	} else if !strings.Contains(err.Error(), "region is required") {
		t.Errorf("Expected 'region is required' error, got: %v", err)
	}

	// 3. Rule check: Invalid CIDR
	invalidCIDR := map[string]interface{}{
		"name":   "my-vpc",
		"region": "us-east-1",
		"cidr":   "invalid-cidr-string",
	}
	err = Validate("vpc", invalidCIDR)
	if err == nil {
		t.Error("Expected error for invalid CIDR")
	} else if !strings.Contains(err.Error(), "invalid cidr format") {
		t.Errorf("Expected 'invalid cidr format' error, got: %v", err)
	}
}

func TestValidate_S3Bucket(t *testing.T) {
	// 1. Valid Bucket
	validBucket := map[string]interface{}{
		"bucket":        "my-valid-s3-bucket",
		"force_destroy": true,
	}
	err := Validate("s3_bucket", validBucket)
	if err != nil {
		t.Fatalf("Expected valid S3 bucket to pass, got: %v", err)
	}

	// 2. Invalid Bucket Name (starts with capital letter, contains spaces, etc.)
	invalidBucket := map[string]interface{}{
		"bucket": "My-Invalid Bucket",
	}
	err = Validate("s3_bucket", invalidBucket)
	if err == nil {
		t.Error("Expected error for invalid S3 bucket name")
	} else if !strings.Contains(err.Error(), "invalid bucket name") {
		t.Errorf("Expected 'invalid bucket name' error, got: %v", err)
	}
}

func TestValidate_SecurityGroup(t *testing.T) {
	// 1. Description exceeds 255 character limit (Go model-level constraint rule check)
	longDesc := strings.Repeat("a", 256)
	sgConfig := map[string]interface{}{
		"name":        "my-security-group",
		"description": longDesc,
		"vpc_id":      "vpc-12345",
	}

	err := Validate("security_group", sgConfig)
	if err == nil {
		t.Error("Expected model-level error for security group description length")
	} else if !strings.Contains(err.Error(), "description must be 255 characters or less") {
		t.Errorf("Expected 'must be 255 characters or less' error, got: %v", err)
	}
}
