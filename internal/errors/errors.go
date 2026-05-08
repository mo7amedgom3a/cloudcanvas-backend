package errors

import (
	"gorm.io/gorm"
)

// Platform-specific error codes
const (
	// Database errors
	CodeDatabaseConnectionFailed  = "DATABASE_CONNECTION_FAILED"
	CodeDatabaseQueryFailed       = "DATABASE_QUERY_FAILED"
	CodeDatabaseTransactionFailed = "DATABASE_TRANSACTION_FAILED"
	CodeDatabaseMigrationFailed   = "DATABASE_MIGRATION_FAILED"
	CodeDatabaseConfigError       = "DATABASE_CONFIG_ERROR"

	// Repository errors
	CodeRepositoryNotFound       = "REPOSITORY_NOT_FOUND"
	CodeRepositoryCreateFailed   = "REPOSITORY_CREATE_FAILED"
	CodeRepositoryUpdateFailed   = "REPOSITORY_UPDATE_FAILED"
	CodeRepositoryDeleteFailed   = "REPOSITORY_DELETE_FAILED"
	CodeRepositoryQueryFailed    = "REPOSITORY_QUERY_FAILED"
	CodeRepositoryDuplicateEntry = "REPOSITORY_DUPLICATE_ENTRY"

	// Resource repository errors
	CodeResourceNotFound     = "RESOURCE_NOT_FOUND"
	CodeResourceCreateFailed = "RESOURCE_CREATE_FAILED"
	CodeResourceUpdateFailed = "RESOURCE_UPDATE_FAILED"
	CodeResourceDeleteFailed = "RESOURCE_DELETE_FAILED"
	CodeResourceInvalidID    = "RESOURCE_INVALID_ID"

	// Project repository errors
	CodeProjectNotFound     = "PROJECT_NOT_FOUND"
	CodeProjectCreateFailed = "PROJECT_CREATE_FAILED"
	CodeProjectUpdateFailed = "PROJECT_UPDATE_FAILED"
	CodeProjectDeleteFailed = "PROJECT_DELETE_FAILED"

	// User repository errors
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeUserCreateFailed   = "USER_CREATE_FAILED"
	CodeUserUpdateFailed   = "USER_UPDATE_FAILED"
	CodeUserDeleteFailed   = "USER_DELETE_FAILED"
	CodeUserDuplicateEmail = "USER_DUPLICATE_EMAIL"

	// Pricing repository errors
	CodePricingNotFound     = "PRICING_NOT_FOUND"
	CodePricingCreateFailed = "PRICING_CREATE_FAILED"
	CodePricingUpdateFailed = "PRICING_UPDATE_FAILED"

	// Auth errors
	CodeAuthUnauthorized       = "AUTH_UNAUTHORIZED"
	CodeAuthForbidden          = "AUTH_FORBIDDEN"
	CodeAuthTokenInvalid       = "AUTH_TOKEN_INVALID"
	CodeAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
)

// NewDatabaseConnectionFailed creates an error for database connection failures
func NewDatabaseConnectionFailed(cause error) *AppError {
	return Wrap(cause, CodeDatabaseConnectionFailed, KindInternal, "Failed to connect to database")
}

// NewDatabaseQueryFailed creates an error for database query failures
func NewDatabaseQueryFailed(operation string, cause error) *AppError {
	return Wrap(cause, CodeDatabaseQueryFailed, KindInternal, "Database query failed").
		WithOp(operation)
}

// NewDatabaseTransactionFailed creates an error for transaction failures
func NewDatabaseTransactionFailed(operation string, cause error) *AppError {
	return Wrap(cause, CodeDatabaseTransactionFailed, KindInternal, "Database transaction failed").
		WithOp(operation)
}

// NewDatabaseConfigError creates an error for database configuration issues
func NewDatabaseConfigError(reason string) *AppError {
	return New(CodeDatabaseConfigError, KindInternal, "Database configuration error").
		WithMeta("reason", reason)
}

// NewRepositoryNotFound creates an error for when a repository record is not found
func NewRepositoryNotFound(resourceType string, id interface{}) *AppError {
	return New(CodeRepositoryNotFound, KindNotFound, "Repository record not found").
		WithMeta("resource_type", resourceType).
		WithMeta("id", id)
}

// NewRepositoryCreateFailed creates an error for repository create failures
func NewRepositoryCreateFailed(resourceType string, cause error) *AppError {
	return Wrap(cause, CodeRepositoryCreateFailed, KindInternal, "Failed to create repository record").
		WithMeta("resource_type", resourceType)
}

// NewRepositoryUpdateFailed creates an error for repository update failures
func NewRepositoryUpdateFailed(resourceType string, cause error) *AppError {
	return Wrap(cause, CodeRepositoryUpdateFailed, KindInternal, "Failed to update repository record").
		WithMeta("resource_type", resourceType)
}

// NewRepositoryDeleteFailed creates an error for repository delete failures
func NewRepositoryDeleteFailed(resourceType string, cause error) *AppError {
	return Wrap(cause, CodeRepositoryDeleteFailed, KindInternal, "Failed to delete repository record").
		WithMeta("resource_type", resourceType)
}

// NewRepositoryDuplicateEntry creates an error for duplicate entry violations
func NewRepositoryDuplicateEntry(resourceType string, field string, value interface{}) *AppError {
	return New(CodeRepositoryDuplicateEntry, KindConflict, "Duplicate entry in repository").
		WithMeta("resource_type", resourceType).
		WithMeta("field", field).
		WithMeta("value", value)
}

// HandleGormError converts GORM errors to AppError
func HandleGormError(err error, resourceType string, operation string) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if appErr := AsAppError(err); appErr != nil {
		return appErr.WithOp(operation)
	}

	// Handle GORM-specific errors
	if err == gorm.ErrRecordNotFound {
		return NewRepositoryNotFound(resourceType, "unknown").
			WithOp(operation)
	}

	// For other GORM errors, wrap them
	return Wrap(err, CodeDatabaseQueryFailed, KindInternal, "Database operation failed").
		WithOp(operation).
		WithMeta("resource_type", resourceType)
}

// NewResourceNotFound creates an error for when a resource is not found
func NewResourceNotFound(resourceID interface{}) *AppError {
	return New(CodeResourceNotFound, KindNotFound, "Resource not found").
		WithMeta("resource_id", resourceID)
}

// NewResourceCreateFailed creates an error for resource creation failures
func NewResourceCreateFailed(cause error) *AppError {
	return Wrap(cause, CodeResourceCreateFailed, KindInternal, "Failed to create resource")
}

// NewResourceUpdateFailed creates an error for resource update failures
func NewResourceUpdateFailed(cause error) *AppError {
	return Wrap(cause, CodeResourceUpdateFailed, KindInternal, "Failed to update resource")
}

// NewResourceDeleteFailed creates an error for resource deletion failures
func NewResourceDeleteFailed(cause error) *AppError {
	return Wrap(cause, CodeResourceDeleteFailed, KindInternal, "Failed to delete resource")
}

// NewResourceInvalidID creates an error for invalid resource ID
func NewResourceInvalidID(id interface{}) *AppError {
	return New(CodeResourceInvalidID, KindValidation, "Invalid resource ID").
		WithMeta("id", id)
}

// NewProjectNotFound creates an error for when a project is not found
func NewProjectNotFound(projectID interface{}) *AppError {
	return New(CodeProjectNotFound, KindNotFound, "Project not found").
		WithMeta("project_id", projectID)
}

// NewProjectCreateFailed creates an error for project creation failures
func NewProjectCreateFailed(cause error) *AppError {
	return Wrap(cause, CodeProjectCreateFailed, KindInternal, "Failed to create project")
}

// NewUserNotFound creates an error for when a user is not found
func NewUserNotFound(userID interface{}) *AppError {
	return New(CodeUserNotFound, KindNotFound, "User not found").
		WithMeta("user_id", userID)
}

// NewUserDuplicateEmail creates an error for duplicate email
func NewUserDuplicateEmail(email string) *AppError {
	return New(CodeUserDuplicateEmail, KindConflict, "User with this email already exists").
		WithMeta("email", email)
}

// NewAuthUnauthorized creates an error for unauthorized access
func NewAuthUnauthorized(reason string) *AppError {
	return New(CodeAuthUnauthorized, KindUnauthorized, "Unauthorized access").
		WithMeta("reason", reason)
}

// NewAuthForbidden creates an error for forbidden access
func NewAuthForbidden(reason string) *AppError {
	return New(CodeAuthForbidden, KindForbidden, "Forbidden access").
		WithMeta("reason", reason)
}

// NewAuthTokenInvalid creates an error for invalid token
func NewAuthTokenInvalid(reason string) *AppError {
	return New(CodeAuthTokenInvalid, KindUnauthorized, "Invalid authentication token").
		WithMeta("reason", reason)
}

// NewAuthInvalidCredentials creates an error for invalid credentials
func NewAuthInvalidCredentials() *AppError {
	return New(CodeAuthInvalidCredentials, KindUnauthorized, "Invalid credentials")
}
