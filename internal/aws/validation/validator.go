package validation

import (
	"encoding/json"
	"fmt"

	awsasg "cloudcanvas-backend/internal/aws/models/compute/autoscaling"
	awsec2 "cloudcanvas-backend/internal/aws/models/compute/ec2"
	awslt "cloudcanvas-backend/internal/aws/models/compute/ec2/launch_template"
	awsinst "cloudcanvas-backend/internal/aws/models/compute/instance_types"
	awslambda "cloudcanvas-backend/internal/aws/models/compute/lambda"
	awslb "cloudcanvas-backend/internal/aws/models/compute/load_balancer"
	awsiam "cloudcanvas-backend/internal/aws/models/iam"
	awsnetworking "cloudcanvas-backend/internal/aws/models/networking"
	awsebs "cloudcanvas-backend/internal/aws/models/storage/ebs"
	awss3 "cloudcanvas-backend/internal/aws/models/storage/s3"
	
)

// ModelCreator defines a constructor function for instantiating empty AWS model structs
type ModelCreator func() interface{}

// Global registry mapping resource type strings directly to concrete AWS model creators
var registry = map[string]ModelCreator{
	// Networking
	"vpc":                func() interface{} { return &awsnetworking.VPC{} },
	"subnet":             func() interface{} { return &awsnetworking.Subnet{} },
	"security_group":     func() interface{} { return &awsnetworking.SecurityGroup{} },
	"internet_gateway":   func() interface{} { return &awsnetworking.InternetGateway{} },
	"nat_gateway":        func() interface{} { return &awsnetworking.NATGateway{} },
	"elastic_ip":         func() interface{} { return &awsnetworking.ElasticIP{} },
	"route_table":        func() interface{} { return &awsnetworking.RouteTable{} },
	"network_acl":        func() interface{} { return &awsnetworking.NetworkACL{} },
	"network_interface":  func() interface{} { return &awsnetworking.NetworkInterface{} },
	"vpc_endpoint":       func() interface{} { return &awsnetworking.VPCEndpoint{} },

	// Storage
	"s3_bucket":          func() interface{} { return &awss3.Bucket{} },
	"s3_bucket_version":  func() interface{} { return &awss3.BucketVersioning{} },
	"s3_bucket_encrypt":  func() interface{} { return &awss3.BucketEncryption{} },
	"s3_bucket_acl":      func() interface{} { return &awss3.BucketACL{} },
	"ebs_volume":         func() interface{} { return &awsebs.Volume{} },

	// IAM
	"iam_role":             func() interface{} { return &awsiam.Role{} },
	"iam_policy":           func() interface{} { return &awsiam.Policy{} },
	"iam_user":             func() interface{} { return &awsiam.User{} },
	"iam_group":            func() interface{} { return &awsiam.Group{} },
	"iam_instance_profile": func() interface{} { return &awsiam.InstanceProfile{} },

	// Compute
	"ec2_instance":            func() interface{} { return &awsec2.Instance{} },
	"instance_type":           func() interface{} { return &awsinst.InstanceTypeInfo{} },
	"launch_template":         func() interface{} { return &awslt.LaunchTemplate{} },
	"lambda_function":         func() interface{} { return &awslambda.Function{} },
	"load_balancer":           func() interface{} { return &awslb.LoadBalancer{} },
	"lb_listener":             func() interface{} { return &awslb.Listener{} },
	"lb_target_group":         func() interface{} { return &awslb.TargetGroup{} },
	"lb_target_group_attach":  func() interface{} { return &awslb.TargetGroupAttachment{} },
	"autoscaling_group":       func() interface{} { return &awsasg.AutoScalingGroup{} },
	"autoscaling_policy":      func() interface{} { return &awsasg.ScalingPolicy{} },
}

// Validator interface matches any model that has its own Validate() method
type Validator interface {
	Validate() error
}

// Validate unmarshals a generic map of resource config parameters into its matching, typed AWS model struct
// and executes its native Validate() method to enforce AWS validations dynamically.
func Validate(resourceType string, config map[string]interface{}) error {
	creator, exists := registry[resourceType]
	if !exists {
		return fmt.Errorf("no validator registered for resource type: %s", resourceType)
	}

	model := creator()

	// Serialize configuration map to intermediate JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Populate the target AWS model struct fields
	if err := json.Unmarshal(jsonData, model); err != nil {
		return fmt.Errorf("invalid configuration structure for %s: %w", resourceType, err)
	}

	// Dynamically run the model's native Validate rule set
	if validator, ok := model.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
