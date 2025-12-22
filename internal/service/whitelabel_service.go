package service

import (
	"context"

	"Krafti_Vibe/internal/domain/models"
	"Krafti_Vibe/internal/pkg/errors"
	"Krafti_Vibe/internal/repository"
	"Krafti_Vibe/internal/service/dto"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WhiteLabelService defines the interface for whitelabel operations
type WhiteLabelService interface {
	// CRUD operations
	CreateWhiteLabel(ctx context.Context, tenantID uuid.UUID, req *dto.CreateWhiteLabelRequest) (*dto.WhiteLabelResponse, error)
	GetWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.WhiteLabelResponse, error)
	GetWhiteLabelByTenant(ctx context.Context, tenantID uuid.UUID) (*dto.WhiteLabelResponse, error)
	GetPublicWhiteLabel(ctx context.Context, tenantID uuid.UUID) (*dto.PublicWhiteLabelResponse, error)
	GetPublicWhiteLabelByDomain(ctx context.Context, domain string) (*dto.PublicWhiteLabelResponse, error)
	UpdateWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateWhiteLabelRequest) (*dto.WhiteLabelResponse, error)
	DeleteWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Partial updates
	UpdateColorScheme(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateColorSchemeRequest) (*dto.WhiteLabelResponse, error)
	UpdateBranding(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateBrandingRequest) (*dto.WhiteLabelResponse, error)
	UpdateDomain(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateDomainRequest) (*dto.WhiteLabelResponse, error)

	// Activation
	ActivateWhiteLabel(ctx context.Context, tenantID uuid.UUID) error
	DeactivateWhiteLabel(ctx context.Context, tenantID uuid.UUID) error

	// Domain validation
	CheckDomainAvailability(ctx context.Context, domain string, tenantID uuid.UUID) (bool, error)
}

type whiteLabelService struct {
	repos  *repository.Repositories
	logger log.AllLogger
}

// NewWhiteLabelService creates a new whitelabel service
func NewWhiteLabelService(repos *repository.Repositories, logger log.AllLogger) WhiteLabelService {
	return &whiteLabelService{
		repos:  repos,
		logger: logger,
	}
}

func (s *whiteLabelService) CreateWhiteLabel(ctx context.Context, tenantID uuid.UUID, req *dto.CreateWhiteLabelRequest) (*dto.WhiteLabelResponse, error) {
	// Check if whitelabel already exists for this tenant
	existing, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err == nil && existing != nil {
		return nil, errors.NewConflictError("whitelabel configuration already exists for this tenant")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("failed to check existing whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to check existing whitelabel", err)
	}

	// Verify tenant exists
	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get tenant", "tenant_id", tenantID, "error", err)
		return nil, errors.NewNotFoundError("tenant")
	}

	// Check if tenant has whitelabeling feature enabled
	if !tenant.Features.WhiteLabeling {
		return nil, errors.NewForbiddenError("whitelabeling feature is not enabled for your plan")
	}

	// Check domain availability if custom domain is provided
	if req.CustomDomain != "" {
		available, err := s.repos.WhiteLabel.CheckDomainAvailability(ctx, req.CustomDomain, tenantID)
		if err != nil {
			s.logger.Error("failed to check domain availability", "error", err)
			return nil, errors.NewInternalError("failed to check domain availability", err)
		}
		if !available {
			return nil, errors.NewConflictError("custom domain is already in use")
		}
	}

	// Create whitelabel configuration
	whitelabel := &models.WhiteLabel{
		TenantID:            tenantID,
		CustomDomain:        req.CustomDomain,
		CustomDomainEnabled: req.CustomDomainEnabled,
		SSLEnabled:          req.SSLEnabled,
		LogoURL:             req.LogoURL,
		LogoDarkURL:         req.LogoDarkURL,
		FaviconURL:          req.FaviconURL,
		AppleTouchIcon:      req.AppleTouchIcon,
		SplashScreenURL:     req.SplashScreenURL,
		PrimaryColor:        req.PrimaryColor,
		SecondaryColor:      req.SecondaryColor,
		AccentColor:         req.AccentColor,
		BackgroundColor:     req.BackgroundColor,
		SurfaceColor:        req.SurfaceColor,
		TextColor:           req.TextColor,
		TextSecondaryColor:  req.TextSecondaryColor,
		ErrorColor:          req.ErrorColor,
		WarningColor:        req.WarningColor,
		SuccessColor:        req.SuccessColor,
		InfoColor:           req.InfoColor,
		FontFamily:          req.FontFamily,
		HeadingFontFamily:   req.HeadingFontFamily,
		FontSize:            req.FontSize,
		FontWeight:          req.FontWeight,
		Theme:               req.Theme,
		DarkModeEnabled:     req.DarkModeEnabled,
		CompanyName:         req.CompanyName,
		CompanyDescription:  req.CompanyDescription,
		CompanyTagline:      req.CompanyTagline,
		CompanyAddress:      req.CompanyAddress,
		CompanyPhone:        req.CompanyPhone,
		CompanyEmail:        req.CompanyEmail,
		SupportEmail:        req.SupportEmail,
		SupportPhone:        req.SupportPhone,
		SupportURL:          req.SupportURL,
		TermsOfServiceURL:   req.TermsOfServiceURL,
		PrivacyPolicyURL:    req.PrivacyPolicyURL,
		CookiePolicyURL:     req.CookiePolicyURL,
		AcceptableUseURL:    req.AcceptableUseURL,
		RefundPolicyURL:     req.RefundPolicyURL,
		SocialLinks:         req.SocialLinks,
		EmailSettings:       req.EmailSettings,
		CustomCSS:           req.CustomCSS,
		CustomJS:            req.CustomJS,
		CustomHead:          req.CustomHead,
		CustomMetaTags:      req.CustomMetaTags,
		CustomAnalytics:     req.CustomAnalytics,
		DefaultLanguage:     req.DefaultLanguage,
		SupportedLocales:    req.SupportedLocales,
		Timezone:            req.Timezone,
		DateFormat:          req.DateFormat,
		TimeFormat:          req.TimeFormat,
		Currency:            req.Currency,
		SEOSettings:         req.SEOSettings,
		UISettings:          req.UISettings,
		CopyrightText:       req.CopyrightText,
		PoweredByText:       req.PoweredByText,
		HidePoweredBy:       req.HidePoweredBy,
		IsActive:            req.IsActive,
	}

	if err := s.repos.WhiteLabel.Create(ctx, whitelabel); err != nil {
		s.logger.Error("failed to create whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to create whitelabel", err)
	}

	// Reload with relationships
	created, err := s.repos.WhiteLabel.GetByID(ctx, whitelabel.ID)
	if err != nil {
		s.logger.Error("failed to reload whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to reload whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(created), nil
}

func (s *whiteLabelService) GetWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*dto.WhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "id", id, "error", err)
		return nil, errors.NewNotFoundError("whitelabel")
	}

	// Verify tenant access
	if whitelabel.TenantID != tenantID {
		return nil, errors.NewForbiddenError("whitelabel does not belong to your tenant")
	}

	return dto.ToWhiteLabelResponse(whitelabel), nil
}

func (s *whiteLabelService) GetWhiteLabelByTenant(ctx context.Context, tenantID uuid.UUID) (*dto.WhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("whitelabel configuration not found for this tenant")
		}
		s.logger.Error("failed to get whitelabel by tenant", "tenant_id", tenantID, "error", err)
		return nil, errors.NewInternalError("failed to get whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(whitelabel), nil
}

func (s *whiteLabelService) GetPublicWhiteLabel(ctx context.Context, tenantID uuid.UUID) (*dto.PublicWhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("whitelabel configuration not found")
		}
		s.logger.Error("failed to get whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to get whitelabel", err)
	}

	// Only return if active
	if !whitelabel.IsActive {
		return nil, errors.NewNotFoundError("whitelabel configuration not active")
	}

	return dto.ToPublicWhiteLabelResponse(whitelabel), nil
}

func (s *whiteLabelService) GetPublicWhiteLabelByDomain(ctx context.Context, domain string) (*dto.PublicWhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByCustomDomain(ctx, domain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("whitelabel configuration not found for domain")
		}
		s.logger.Error("failed to get whitelabel by domain", "domain", domain, "error", err)
		return nil, errors.NewInternalError("failed to get whitelabel", err)
	}

	return dto.ToPublicWhiteLabelResponse(whitelabel), nil
}

func (s *whiteLabelService) UpdateWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req *dto.UpdateWhiteLabelRequest) (*dto.WhiteLabelResponse, error) {
	// Get existing whitelabel
	whitelabel, err := s.repos.WhiteLabel.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "id", id, "error", err)
		return nil, errors.NewNotFoundError("whitelabel")
	}

	// Verify tenant access
	if whitelabel.TenantID != tenantID {
		return nil, errors.NewForbiddenError("whitelabel does not belong to your tenant")
	}

	// Check domain availability if custom domain is being updated
	if req.CustomDomain != nil && *req.CustomDomain != whitelabel.CustomDomain {
		available, err := s.repos.WhiteLabel.CheckDomainAvailability(ctx, *req.CustomDomain, tenantID)
		if err != nil {
			s.logger.Error("failed to check domain availability", "error", err)
			return nil, errors.NewInternalError("failed to check domain availability", err)
		}
		if !available {
			return nil, errors.NewConflictError("custom domain is already in use")
		}
	}

	// Update fields
	if req.CustomDomain != nil {
		whitelabel.CustomDomain = *req.CustomDomain
	}
	if req.CustomDomainEnabled != nil {
		whitelabel.CustomDomainEnabled = *req.CustomDomainEnabled
	}
	if req.SSLEnabled != nil {
		whitelabel.SSLEnabled = *req.SSLEnabled
	}
	if req.LogoURL != nil {
		whitelabel.LogoURL = *req.LogoURL
	}
	if req.LogoDarkURL != nil {
		whitelabel.LogoDarkURL = *req.LogoDarkURL
	}
	if req.FaviconURL != nil {
		whitelabel.FaviconURL = *req.FaviconURL
	}
	if req.AppleTouchIcon != nil {
		whitelabel.AppleTouchIcon = *req.AppleTouchIcon
	}
	if req.SplashScreenURL != nil {
		whitelabel.SplashScreenURL = *req.SplashScreenURL
	}
	if req.PrimaryColor != nil {
		whitelabel.PrimaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		whitelabel.SecondaryColor = *req.SecondaryColor
	}
	if req.AccentColor != nil {
		whitelabel.AccentColor = *req.AccentColor
	}
	if req.BackgroundColor != nil {
		whitelabel.BackgroundColor = *req.BackgroundColor
	}
	if req.SurfaceColor != nil {
		whitelabel.SurfaceColor = *req.SurfaceColor
	}
	if req.TextColor != nil {
		whitelabel.TextColor = *req.TextColor
	}
	if req.TextSecondaryColor != nil {
		whitelabel.TextSecondaryColor = *req.TextSecondaryColor
	}
	if req.ErrorColor != nil {
		whitelabel.ErrorColor = *req.ErrorColor
	}
	if req.WarningColor != nil {
		whitelabel.WarningColor = *req.WarningColor
	}
	if req.SuccessColor != nil {
		whitelabel.SuccessColor = *req.SuccessColor
	}
	if req.InfoColor != nil {
		whitelabel.InfoColor = *req.InfoColor
	}
	if req.FontFamily != nil {
		whitelabel.FontFamily = *req.FontFamily
	}
	if req.HeadingFontFamily != nil {
		whitelabel.HeadingFontFamily = *req.HeadingFontFamily
	}
	if req.FontSize != nil {
		whitelabel.FontSize = *req.FontSize
	}
	if req.FontWeight != nil {
		whitelabel.FontWeight = *req.FontWeight
	}
	if req.Theme != nil {
		whitelabel.Theme = *req.Theme
	}
	if req.DarkModeEnabled != nil {
		whitelabel.DarkModeEnabled = *req.DarkModeEnabled
	}
	if req.CompanyName != nil {
		whitelabel.CompanyName = *req.CompanyName
	}
	if req.CompanyDescription != nil {
		whitelabel.CompanyDescription = *req.CompanyDescription
	}
	if req.CompanyTagline != nil {
		whitelabel.CompanyTagline = *req.CompanyTagline
	}
	if req.CompanyAddress != nil {
		whitelabel.CompanyAddress = *req.CompanyAddress
	}
	if req.CompanyPhone != nil {
		whitelabel.CompanyPhone = *req.CompanyPhone
	}
	if req.CompanyEmail != nil {
		whitelabel.CompanyEmail = *req.CompanyEmail
	}
	if req.SupportEmail != nil {
		whitelabel.SupportEmail = *req.SupportEmail
	}
	if req.SupportPhone != nil {
		whitelabel.SupportPhone = *req.SupportPhone
	}
	if req.SupportURL != nil {
		whitelabel.SupportURL = *req.SupportURL
	}
	if req.TermsOfServiceURL != nil {
		whitelabel.TermsOfServiceURL = *req.TermsOfServiceURL
	}
	if req.PrivacyPolicyURL != nil {
		whitelabel.PrivacyPolicyURL = *req.PrivacyPolicyURL
	}
	if req.CookiePolicyURL != nil {
		whitelabel.CookiePolicyURL = *req.CookiePolicyURL
	}
	if req.AcceptableUseURL != nil {
		whitelabel.AcceptableUseURL = *req.AcceptableUseURL
	}
	if req.RefundPolicyURL != nil {
		whitelabel.RefundPolicyURL = *req.RefundPolicyURL
	}
	if req.SocialLinks != nil {
		whitelabel.SocialLinks = *req.SocialLinks
	}
	if req.EmailSettings != nil {
		whitelabel.EmailSettings = *req.EmailSettings
	}
	if req.CustomCSS != nil {
		whitelabel.CustomCSS = *req.CustomCSS
	}
	if req.CustomJS != nil {
		whitelabel.CustomJS = *req.CustomJS
	}
	if req.CustomHead != nil {
		whitelabel.CustomHead = *req.CustomHead
	}
	if req.CustomMetaTags != nil {
		whitelabel.CustomMetaTags = *req.CustomMetaTags
	}
	if req.CustomAnalytics != nil {
		whitelabel.CustomAnalytics = *req.CustomAnalytics
	}
	if req.DefaultLanguage != nil {
		whitelabel.DefaultLanguage = *req.DefaultLanguage
	}
	if req.SupportedLocales != nil {
		whitelabel.SupportedLocales = *req.SupportedLocales
	}
	if req.Timezone != nil {
		whitelabel.Timezone = *req.Timezone
	}
	if req.DateFormat != nil {
		whitelabel.DateFormat = *req.DateFormat
	}
	if req.TimeFormat != nil {
		whitelabel.TimeFormat = *req.TimeFormat
	}
	if req.Currency != nil {
		whitelabel.Currency = *req.Currency
	}
	if req.SEOSettings != nil {
		whitelabel.SEOSettings = *req.SEOSettings
	}
	if req.UISettings != nil {
		whitelabel.UISettings = *req.UISettings
	}
	if req.CopyrightText != nil {
		whitelabel.CopyrightText = *req.CopyrightText
	}
	if req.PoweredByText != nil {
		whitelabel.PoweredByText = *req.PoweredByText
	}
	if req.HidePoweredBy != nil {
		whitelabel.HidePoweredBy = *req.HidePoweredBy
	}
	if req.IsActive != nil {
		whitelabel.IsActive = *req.IsActive
	}

	// Update whitelabel
	if err := s.repos.WhiteLabel.Update(ctx, whitelabel); err != nil {
		s.logger.Error("failed to update whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to update whitelabel", err)
	}

	// Reload with relationships
	updated, err := s.repos.WhiteLabel.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to reload whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to reload whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(updated), nil
}

func (s *whiteLabelService) DeleteWhiteLabel(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get whitelabel to verify tenant access
	whitelabel, err := s.repos.WhiteLabel.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "id", id, "error", err)
		return errors.NewNotFoundError("whitelabel")
	}

	// Verify tenant access
	if whitelabel.TenantID != tenantID {
		return errors.NewForbiddenError("whitelabel does not belong to your tenant")
	}

	if err := s.repos.WhiteLabel.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete whitelabel", "error", err)
		return errors.NewInternalError("failed to delete whitelabel", err)
	}

	return nil
}

func (s *whiteLabelService) UpdateColorScheme(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateColorSchemeRequest) (*dto.WhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "error", err)
		return nil, errors.NewNotFoundError("whitelabel configuration not found")
	}

	// Update color scheme
	whitelabel.PrimaryColor = req.PrimaryColor
	whitelabel.SecondaryColor = req.SecondaryColor
	whitelabel.AccentColor = req.AccentColor

	if req.BackgroundColor != "" {
		whitelabel.BackgroundColor = req.BackgroundColor
	}
	if req.SurfaceColor != "" {
		whitelabel.SurfaceColor = req.SurfaceColor
	}
	if req.TextColor != "" {
		whitelabel.TextColor = req.TextColor
	}
	if req.TextSecondaryColor != "" {
		whitelabel.TextSecondaryColor = req.TextSecondaryColor
	}
	if req.ErrorColor != "" {
		whitelabel.ErrorColor = req.ErrorColor
	}
	if req.WarningColor != "" {
		whitelabel.WarningColor = req.WarningColor
	}
	if req.SuccessColor != "" {
		whitelabel.SuccessColor = req.SuccessColor
	}
	if req.InfoColor != "" {
		whitelabel.InfoColor = req.InfoColor
	}

	if err := s.repos.WhiteLabel.Update(ctx, whitelabel); err != nil {
		s.logger.Error("failed to update color scheme", "error", err)
		return nil, errors.NewInternalError("failed to update color scheme", err)
	}

	updated, err := s.repos.WhiteLabel.GetByID(ctx, whitelabel.ID)
	if err != nil {
		s.logger.Error("failed to reload whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to reload whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(updated), nil
}

func (s *whiteLabelService) UpdateBranding(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateBrandingRequest) (*dto.WhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "error", err)
		return nil, errors.NewNotFoundError("whitelabel configuration not found")
	}

	// Update branding assets
	whitelabel.LogoURL = req.LogoURL
	if req.LogoDarkURL != "" {
		whitelabel.LogoDarkURL = req.LogoDarkURL
	}
	if req.FaviconURL != "" {
		whitelabel.FaviconURL = req.FaviconURL
	}
	if req.AppleTouchIcon != "" {
		whitelabel.AppleTouchIcon = req.AppleTouchIcon
	}
	if req.SplashScreenURL != "" {
		whitelabel.SplashScreenURL = req.SplashScreenURL
	}

	if err := s.repos.WhiteLabel.Update(ctx, whitelabel); err != nil {
		s.logger.Error("failed to update branding", "error", err)
		return nil, errors.NewInternalError("failed to update branding", err)
	}

	updated, err := s.repos.WhiteLabel.GetByID(ctx, whitelabel.ID)
	if err != nil {
		s.logger.Error("failed to reload whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to reload whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(updated), nil
}

func (s *whiteLabelService) UpdateDomain(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateDomainRequest) (*dto.WhiteLabelResponse, error) {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "error", err)
		return nil, errors.NewNotFoundError("whitelabel configuration not found")
	}

	// Check domain availability if domain is being changed
	if req.CustomDomain != whitelabel.CustomDomain {
		available, err := s.repos.WhiteLabel.CheckDomainAvailability(ctx, req.CustomDomain, tenantID)
		if err != nil {
			s.logger.Error("failed to check domain availability", "error", err)
			return nil, errors.NewInternalError("failed to check domain availability", err)
		}
		if !available {
			return nil, errors.NewConflictError("custom domain is already in use")
		}
	}

	// Update domain settings
	whitelabel.CustomDomain = req.CustomDomain
	whitelabel.CustomDomainEnabled = req.CustomDomainEnabled
	whitelabel.SSLEnabled = req.SSLEnabled

	if err := s.repos.WhiteLabel.Update(ctx, whitelabel); err != nil {
		s.logger.Error("failed to update domain", "error", err)
		return nil, errors.NewInternalError("failed to update domain", err)
	}

	updated, err := s.repos.WhiteLabel.GetByID(ctx, whitelabel.ID)
	if err != nil {
		s.logger.Error("failed to reload whitelabel", "error", err)
		return nil, errors.NewInternalError("failed to reload whitelabel", err)
	}

	return dto.ToWhiteLabelResponse(updated), nil
}

func (s *whiteLabelService) ActivateWhiteLabel(ctx context.Context, tenantID uuid.UUID) error {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "error", err)
		return errors.NewNotFoundError("whitelabel configuration not found")
	}

	if err := s.repos.WhiteLabel.Activate(ctx, whitelabel.ID); err != nil {
		s.logger.Error("failed to activate whitelabel", "error", err)
		return errors.NewInternalError("failed to activate whitelabel", err)
	}

	return nil
}

func (s *whiteLabelService) DeactivateWhiteLabel(ctx context.Context, tenantID uuid.UUID) error {
	whitelabel, err := s.repos.WhiteLabel.GetByTenantID(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to get whitelabel", "error", err)
		return errors.NewNotFoundError("whitelabel configuration not found")
	}

	if err := s.repos.WhiteLabel.Deactivate(ctx, whitelabel.ID); err != nil {
		s.logger.Error("failed to deactivate whitelabel", "error", err)
		return errors.NewInternalError("failed to deactivate whitelabel", err)
	}

	return nil
}

func (s *whiteLabelService) CheckDomainAvailability(ctx context.Context, domain string, tenantID uuid.UUID) (bool, error) {
	available, err := s.repos.WhiteLabel.CheckDomainAvailability(ctx, domain, tenantID)
	if err != nil {
		s.logger.Error("failed to check domain availability", "error", err)
		return false, errors.NewInternalError("failed to check domain availability", err)
	}
	return available, nil
}
