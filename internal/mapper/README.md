# Mapper Layer

The **Mapper Layer** acts as a bridge between three distinct representation models of cloud infrastructure in CloudCanvas:

1. **Domain Representation (`resource.Resource`)**: A cloud-agnostic model representing a logical cloud resource, including basic parameters (`ID`, `Name`, `Provider`, `Region`, relationships like `ParentID` and `DependsOn`, and a generic configuration schema inside `Metadata`).
2. **Database GORM Model (`models.Resource`)**: The GORM entity designed for Postgres/CockroachDB persistence, linking to projects, resource types, containment structures, and dependency graphs.
3. **AWS-Specific Infrastructure Models (`internal/aws/models/*`)**: Fully-typed AWS entities (such as `networking.VPC`, `networking.Subnet`, `storage.S3Bucket`) that implement provider-specific validation, default values, tagging rules, and structures mapped directly to Terraform or the AWS Go SDK.

---

## Architectural Position

```
     ┌────────────────────────┐
     │      Parser Layer      │ (Processes Frontend Canvas JSON/IR)
     └───────────┬────────────┘
                 │ (Produces)
                 ▼
     ┌────────────────────────┐
     │  Domain Representation │ (resource.Resource - cloud-agnostic)
     └───────────┬────────────┘
                 │
        Mapped ◄─┴─► Mapped
          via          via
      ToDBResource  ToAWSModel
          and          and
    ToDomainResource FromAWSModel
          │            │
          ▼            ▼
 ┌──────────────┐   ┌──────────────┐
 │ Database GORM│   │  AWS Models  │ (networking.VPC, s3.Bucket, etc.)
 │   (postgres) │   │ (Terraform)  │
 └──────────────┘   └──────────────┘
```

---

## Available Mapping API

All mapper routines are defined under the package name `resource` in `internal/mapper/`.

### 1. Database Model ↔ Domain Resource Conversion

*   **`ToDomainResource(dbRes *models.Resource) (*Resource, error)`**
    *   **Description**: Converts a GORM DB resource entity (`models.Resource`) to a domain-level `Resource`.
    *   **Relationship Resolution**: If database relationships (`ParentResources` and `FromDependencies`) are preloaded in the GORM record, they are automatically resolved and mapped to `ParentID` and `DependsOn` slices using their cloud-agnostic/original IDs.
    *   **Configuration Extraction**: Deserializes the database JSONB `Config` field into the resource's `Metadata` map.

*   **`ToDBResource(domainRes *Resource, projectID uuid.UUID, resourceTypeID uint) (*models.Resource, error)`**
    *   **Description**: Converts a domain-level `Resource` into a database `models.Resource` suitable for GORM persistence.
    *   **UUID Robustness**: Automatically checks if the domain ID is a valid UUID. If it is, it parses and reuses it; if not (e.g. frontend temp node IDs like `vpc-1`), it generates a new random `uuid.UUID` while preserving the original string under `OriginalID` for reference resolution.
    *   **Metadata Packaging**: Serializes the resource's `Metadata` map into a JSONB `Config` datatype.

### 2. Domain Resource ↔ AWS Model Mapping

Both `Resource` and `ResourceOutput` support native conversion methods to AWS-specific structs.

*   **`ToAWSModel(target interface{}) error`**
    *   **Description**: Serializes the generic `Metadata` map from the domain resource and unmarshals it into a strongly-typed AWS model struct (e.g., `*networking.VPC` or `*s3.Bucket`).
    *   **Use Case**: Converting diagram nodes into AWS structures to execute provider-specific validations (`target.Validate()`) or generate Terraform code.

*   **`FromAWSModel(awsModel interface{}) error`**
    *   **Description**: Serializes a strongly-typed AWS model structure and populates the domain resource's `Metadata` map.
    *   **Use Case**: Storing output or updated configuration states from the cloud platform back into the cloud-agnostic model.

---

## Practical Examples

### 1. Transforming GORM Database Models to AWS-Specific Models

```go
package main

import (
    "fmt"
    "log"

    "cloudcanvas-backend/internal/mapper"
    awsnetworking "cloudcanvas-backend/internal/aws/models/networking"
    "cloudcanvas-backend/internal/models"
)

func ProcessResource(dbResource *models.Resource) {
    // 1. Convert DB GORM Entity to Domain Representation
    domainResource, err := resource.ToDomainResource(dbResource)
    if err != nil {
        log.Fatalf("Failed to convert DB to domain model: %v", err)
    }

    // 2. Map domain configuration to typed AWS VPC
    if domainResource.Type.Name == "vpc" {
        var awsVPC awsnetworking.VPC
        if err := domainResource.ToAWSModel(&awsVPC); err != nil {
            log.Fatalf("Failed to map to AWS model: %v", err)
        }

        // 3. Leverage AWS-specific validations
        if err := awsVPC.Validate(); err != nil {
            fmt.Printf("AWS VPC is invalid: %v\n", err)
            return
        }
        
        fmt.Printf("VPC '%s' successfully parsed and validated for AWS!\n", awsVPC.Name)
    }
}
```

### 2. Creating Database Records from a Domain Resource

```go
package main

import (
    "context"
    "log"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "cloudcanvas-backend/internal/mapper"
)

func SaveCanvasNode(ctx context.Context, db *gorm.DB, domainRes *resource.Resource, projID uuid.UUID, typeID uint) {
    // 1. Convert domain model into persistence entity
    dbRecord, err := resource.ToDBResource(domainRes, projID, typeID)
    if err != nil {
        log.Fatalf("Failed conversion: %v", err)
    }

    // 2. Persist to DB using GORM
    if err := db.WithContext(ctx).Create(dbRecord).Error; err != nil {
        log.Fatalf("Database save failed: %v", err)
    }
}
```
