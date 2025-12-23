package models

import (
	"github.com/google/uuid"
)

type ServiceCategory string

const (
	// 🏠 Home Maintenance & Construction
	ServiceCategoryMasonry         ServiceCategory = "masonry"          // bricklayer, tiler, concrete works
	ServiceCategoryCarpentry       ServiceCategory = "carpentry"        // furniture, fittings, roofing
	ServiceCategoryPlumbing        ServiceCategory = "plumbing"         // water systems, fixtures, drainage
	ServiceCategoryElectrical      ServiceCategory = "electrical"       // wiring, lighting, installation
	ServiceCategoryPainting        ServiceCategory = "painting"         // interior/exterior painting
	ServiceCategoryWelding         ServiceCategory = "welding"          // metal fabrication & repair
	ServiceCategoryGlaziery        ServiceCategory = "glaziery"         // glassworks, windows, mirrors
	ServiceCategoryRoofing         ServiceCategory = "roofing"          // roof installation & repair
	ServiceCategoryTiling          ServiceCategory = "tiling"           // floor/wall tiling
	ServiceCategoryConcreteWorks   ServiceCategory = "concrete_works"   // foundations, slabs, driveways
	ServiceCategoryLandscaping     ServiceCategory = "landscaping"      // lawns, stones, outdoor design
	ServiceCategoryGardening       ServiceCategory = "gardening"        // plants, lawn care, trimming
	ServiceCategoryPestControl     ServiceCategory = "pest_control"     // fumigation, rodent control
	ServiceCategoryCleaning        ServiceCategory = "cleaning"         // residential, office, post-construction
	ServiceCategoryHVAC            ServiceCategory = "hvac"             // air conditioning & ventilation
	ServiceCategorySecuritySystems ServiceCategory = "security_systems" // CCTV, alarms, gates
	ServiceCategoryApplianceRepair ServiceCategory = "appliance_repair" // home appliances maintenance
	ServiceCategoryPoolMaintenance ServiceCategory = "pool_maintenance" // pool cleaning & repair
	ServiceCategoryHandyman        ServiceCategory = "handyman"         // general home fixes & small jobs

	// 💻 Tech & Electronics
	ServiceCategoryPhoneRepair      ServiceCategory = "phone_repair"
	ServiceCategoryComputerRepair   ServiceCategory = "computer_repair"
	ServiceCategoryITSupport        ServiceCategory = "it_support"
	ServiceCategoryNetworking       ServiceCategory = "networking"        // setup, troubleshooting, cabling
	ServiceCategoryCCTVInstallation ServiceCategory = "cctv_installation" // common artisan service

	// 🚗 Automotive
	ServiceCategoryMechanic       ServiceCategory = "mechanic"        // auto repair & diagnostics
	ServiceCategoryAutoElectrical ServiceCategory = "auto_electrical" // wiring, diagnostics, alarms
	ServiceCategoryPanelBeating   ServiceCategory = "panel_beating"   // bodywork & painting
	ServiceCategoryAutoSpraying   ServiceCategory = "auto_spraying"   // car painting
	ServiceCategoryCarWash        ServiceCategory = "car_wash"

	// 🧰 Metal, Wood & Craft
	ServiceCategoryBlacksmith       ServiceCategory = "blacksmith"
	ServiceCategoryMetalFabrication ServiceCategory = "metal_fabrication"
	ServiceCategoryWoodwork         ServiceCategory = "woodwork"
	ServiceCategoryFurnitureMaking  ServiceCategory = "furniture_making"

	// 🧑‍🎨 Design & Construction Services
	ServiceCategoryArchitecturalDesign   ServiceCategory = "architectural_design"
	ServiceCategoryInteriorDesign        ServiceCategory = "interior_design"
	ServiceCategoryStructuralEngineering ServiceCategory = "structural_engineering"
	ServiceCategoryCivilEngineering      ServiceCategory = "civil_engineering"
	ServiceCategoryLandSurveying         ServiceCategory = "land_surveying"
	ServiceCategoryLandPlanning          ServiceCategory = "land_planning"
	ServiceCategoryMechanicalEngineering ServiceCategory = "mechanical_engineering"

	// 👗 Personal & Lifestyle Services
	ServiceCategoryHairBeauty    ServiceCategory = "hair_beauty"    // hairdresser, barber, beauty salon
	ServiceCategoryTailoring     ServiceCategory = "tailoring"      // dressmaking, alterations
	ServiceCategoryLaundry       ServiceCategory = "laundry"        // dry cleaning, ironing
	ServiceCategoryEventServices ServiceCategory = "event_services" // decoration, catering, rentals
	ServiceCategoryPhotography   ServiceCategory = "photography"    // event & product shoots

	// 🛠️ Miscellaneous
	ServiceCategoryTraining     ServiceCategory = "training"     // artisan training & skills
	ServiceCategoryConsultation ServiceCategory = "consultation" // project estimation, inspection
	ServiceCategoryOther        ServiceCategory = "other"
)

type Service struct {
	BaseModel

	// Multi-tenancy
	TenantID  uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_service_tenant_category"`
	ArtisanID *uuid.UUID `json:"artisan_id,omitempty" gorm:"type:uuid;index"` // null = org-wide service

	// Basic Info
	Name        string          `json:"name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Description string          `json:"description,omitempty" gorm:"type:text"`
	Category    ServiceCategory `json:"category" gorm:"type:varchar(50);not null;index:idx_service_tenant_category" validate:"required"`

	// Pricing
	Price         float64 `json:"price" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	Currency      string  `json:"currency" gorm:"size:3;default:'USD'" validate:"required,len=3"`
	DepositAmount float64 `json:"deposit_amount,omitempty" gorm:"type:decimal(10,2);default:0"`

	// Duration
	DurationMinutes int `json:"duration_minutes" gorm:"not null" validate:"required,min=5"`
	BufferMinutes   int `json:"buffer_minutes" gorm:"default:0" validate:"min=0"` // Cleanup time

	// Availability
	IsActive       bool   `json:"is_active" gorm:"default:true"`
	MaxBookingsDay int    `json:"max_bookings_day" gorm:"default:0"` // 0 = unlimited
	ImageURL       string `json:"image_url,omitempty" gorm:"size:500"`

	// Requirements
	RequiresDeposit bool     `json:"requires_deposit" gorm:"default:false"`
	Tags            []string `json:"tags,omitempty" gorm:"type:text[]"`

	// Metadata
	Metadata JSONB `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Relationships
	Tenant   *Tenant        `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Artisan  *User          `json:"artisan,omitempty" gorm:"foreignKey:ArtisanID"`
	Bookings []Booking      `json:"bookings,omitempty" gorm:"foreignKey:ServiceID"`
	Addons   []ServiceAddon `json:"addons,omitempty" gorm:"many2many:service_addon_relations"`
}

// Business Methods
func (s *Service) GetTotalDuration() int {
	return s.DurationMinutes + s.BufferMinutes
}

func (s *Service) GetDepositPercentage() float64 {
	if s.Price == 0 {
		return 0
	}
	return (s.DepositAmount / s.Price) * 100
}
