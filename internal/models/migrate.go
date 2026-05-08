package models

import (
	"gorm.io/gorm"
)

// AutoMigrate migrates all the database models
func AutoMigrate(db *gorm.DB) error {
	// Execute custom extensions if needed (e.g. for UUID)
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	return db.AutoMigrate(
		&Category{},
		&ComplianceStandard{},
		&DependencyType{},
		&HiddenDependency{},
		&IACFormat{},
		&IACTarget{},
		&PricingComponent{},
		&PricingRate{},
		&Project{},
		&ProjectOutput{},
		&ProjectPricing{},
		&ProjectVariable{},
		&ProjectVersion{},
		&Resource{},
		&ResourceCategory{},
		&ResourceConstraint{},
		&ResourceContainment{},
		&ResourceDependency{},
		&ResourceKind{},
		&ResourcePricing{},
		&ResourceType{},
		&Review{},
		&ServicePricing{},
		&ServiceTypePricing{},
		&Technology{},
		&Template{},
		&TemplateCompliance{},
		&TemplateComponent{},
		&TemplateFeature{},
		&TemplateIACFormat{},
		&TemplateTechnology{},
		&TemplateUseCase{},
		&ResourceUIState{},
		&ProjectUIState{},
		&User{},
	)
}
