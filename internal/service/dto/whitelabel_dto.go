package dto

import (
	"time"

	"Krafti_Vibe/internal/domain/models"

	"github.com/google/uuid"
)

// ============================================================================
// WhiteLabel Request DTOs
// ============================================================================

// CreateWhiteLabelRequest represents a request to create whitelabel configuration
type CreateWhiteLabelRequest struct {
	// Domain Configuration
	CustomDomain        string `json:"custom_domain,omitempty" validate:"omitempty,fqdn"`
	CustomDomainEnabled bool   `json:"custom_domain_enabled"`
	SSLEnabled          bool   `json:"ssl_enabled"`

	// Branding Assets
	LogoURL         string `json:"logo_url,omitempty" validate:"omitempty,url"`
	LogoDarkURL     string `json:"logo_dark_url,omitempty" validate:"omitempty,url"`
	FaviconURL      string `json:"favicon_url,omitempty" validate:"omitempty,url"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty" validate:"omitempty,url"`
	SplashScreenURL string `json:"splash_screen_url,omitempty" validate:"omitempty,url"`

	// Color Scheme
	PrimaryColor       string `json:"primary_color,omitempty" validate:"omitempty,hexcolor"`
	SecondaryColor     string `json:"secondary_color,omitempty" validate:"omitempty,hexcolor"`
	AccentColor        string `json:"accent_color,omitempty" validate:"omitempty,hexcolor"`
	BackgroundColor    string `json:"background_color,omitempty" validate:"omitempty,hexcolor"`
	SurfaceColor       string `json:"surface_color,omitempty" validate:"omitempty,hexcolor"`
	TextColor          string `json:"text_color,omitempty" validate:"omitempty,hexcolor"`
	TextSecondaryColor string `json:"text_secondary_color,omitempty" validate:"omitempty,hexcolor"`
	ErrorColor         string `json:"error_color,omitempty" validate:"omitempty,hexcolor"`
	WarningColor       string `json:"warning_color,omitempty" validate:"omitempty,hexcolor"`
	SuccessColor       string `json:"success_color,omitempty" validate:"omitempty,hexcolor"`
	InfoColor          string `json:"info_color,omitempty" validate:"omitempty,hexcolor"`

	// Typography
	FontFamily        string `json:"font_family,omitempty"`
	HeadingFontFamily string `json:"heading_font_family,omitempty"`
	FontSize          string `json:"font_size,omitempty"`
	FontWeight        string `json:"font_weight,omitempty"`

	// Theme Settings
	Theme           models.ThemeConfig `json:"theme"`
	DarkModeEnabled bool               `json:"dark_mode_enabled"`

	// Company Information
	CompanyName        string `json:"company_name,omitempty"`
	CompanyDescription string `json:"company_description,omitempty"`
	CompanyTagline     string `json:"company_tagline,omitempty"`
	CompanyAddress     string `json:"company_address,omitempty"`
	CompanyPhone       string `json:"company_phone,omitempty"`
	CompanyEmail       string `json:"company_email,omitempty" validate:"omitempty,email"`
	SupportEmail       string `json:"support_email,omitempty" validate:"omitempty,email"`
	SupportPhone       string `json:"support_phone,omitempty"`
	SupportURL         string `json:"support_url,omitempty" validate:"omitempty,url"`

	// Legal & Policy Links
	TermsOfServiceURL string `json:"terms_of_service_url,omitempty" validate:"omitempty,url"`
	PrivacyPolicyURL  string `json:"privacy_policy_url,omitempty" validate:"omitempty,url"`
	CookiePolicyURL   string `json:"cookie_policy_url,omitempty" validate:"omitempty,url"`
	AcceptableUseURL  string `json:"acceptable_use_url,omitempty" validate:"omitempty,url"`
	RefundPolicyURL   string `json:"refund_policy_url,omitempty" validate:"omitempty,url"`

	// Social Media Links
	SocialLinks models.SocialLinks `json:"social_links"`

	// Email Branding
	EmailSettings models.EmailBranding `json:"email_settings"`

	// Advanced Customization
	CustomCSS      string             `json:"custom_css,omitempty"`
	CustomJS       string             `json:"custom_js,omitempty"`
	CustomHead     string             `json:"custom_head,omitempty"`
	CustomMetaTags models.CustomMeta  `json:"custom_meta_tags"`
	CustomAnalytics models.Analytics  `json:"custom_analytics"`

	// Localization
	DefaultLanguage  string   `json:"default_language,omitempty"`
	SupportedLocales []string `json:"supported_locales,omitempty"`
	Timezone         string   `json:"timezone,omitempty"`
	DateFormat       string   `json:"date_format,omitempty"`
	TimeFormat       string   `json:"time_format,omitempty"`
	Currency         string   `json:"currency,omitempty"`

	// SEO Settings
	SEOSettings models.SEOConfig `json:"seo_settings"`

	// Feature Toggles
	UISettings models.UISettings `json:"ui_settings"`

	// Miscellaneous
	CopyrightText string `json:"copyright_text,omitempty"`
	PoweredByText string `json:"powered_by_text,omitempty"`
	HidePoweredBy bool   `json:"hide_powered_by"`

	// Status
	IsActive bool `json:"is_active"`
}

// UpdateWhiteLabelRequest represents a request to update whitelabel configuration
type UpdateWhiteLabelRequest struct {
	// Domain Configuration
	CustomDomain        *string `json:"custom_domain,omitempty" validate:"omitempty,fqdn"`
	CustomDomainEnabled *bool   `json:"custom_domain_enabled,omitempty"`
	SSLEnabled          *bool   `json:"ssl_enabled,omitempty"`

	// Branding Assets
	LogoURL         *string `json:"logo_url,omitempty" validate:"omitempty,url"`
	LogoDarkURL     *string `json:"logo_dark_url,omitempty" validate:"omitempty,url"`
	FaviconURL      *string `json:"favicon_url,omitempty" validate:"omitempty,url"`
	AppleTouchIcon  *string `json:"apple_touch_icon,omitempty" validate:"omitempty,url"`
	SplashScreenURL *string `json:"splash_screen_url,omitempty" validate:"omitempty,url"`

	// Color Scheme
	PrimaryColor       *string `json:"primary_color,omitempty" validate:"omitempty,hexcolor"`
	SecondaryColor     *string `json:"secondary_color,omitempty" validate:"omitempty,hexcolor"`
	AccentColor        *string `json:"accent_color,omitempty" validate:"omitempty,hexcolor"`
	BackgroundColor    *string `json:"background_color,omitempty" validate:"omitempty,hexcolor"`
	SurfaceColor       *string `json:"surface_color,omitempty" validate:"omitempty,hexcolor"`
	TextColor          *string `json:"text_color,omitempty" validate:"omitempty,hexcolor"`
	TextSecondaryColor *string `json:"text_secondary_color,omitempty" validate:"omitempty,hexcolor"`
	ErrorColor         *string `json:"error_color,omitempty" validate:"omitempty,hexcolor"`
	WarningColor       *string `json:"warning_color,omitempty" validate:"omitempty,hexcolor"`
	SuccessColor       *string `json:"success_color,omitempty" validate:"omitempty,hexcolor"`
	InfoColor          *string `json:"info_color,omitempty" validate:"omitempty,hexcolor"`

	// Typography
	FontFamily        *string `json:"font_family,omitempty"`
	HeadingFontFamily *string `json:"heading_font_family,omitempty"`
	FontSize          *string `json:"font_size,omitempty"`
	FontWeight        *string `json:"font_weight,omitempty"`

	// Theme Settings
	Theme           *models.ThemeConfig `json:"theme,omitempty"`
	DarkModeEnabled *bool               `json:"dark_mode_enabled,omitempty"`

	// Company Information
	CompanyName        *string `json:"company_name,omitempty"`
	CompanyDescription *string `json:"company_description,omitempty"`
	CompanyTagline     *string `json:"company_tagline,omitempty"`
	CompanyAddress     *string `json:"company_address,omitempty"`
	CompanyPhone       *string `json:"company_phone,omitempty"`
	CompanyEmail       *string `json:"company_email,omitempty" validate:"omitempty,email"`
	SupportEmail       *string `json:"support_email,omitempty" validate:"omitempty,email"`
	SupportPhone       *string `json:"support_phone,omitempty"`
	SupportURL         *string `json:"support_url,omitempty" validate:"omitempty,url"`

	// Legal & Policy Links
	TermsOfServiceURL *string `json:"terms_of_service_url,omitempty" validate:"omitempty,url"`
	PrivacyPolicyURL  *string `json:"privacy_policy_url,omitempty" validate:"omitempty,url"`
	CookiePolicyURL   *string `json:"cookie_policy_url,omitempty" validate:"omitempty,url"`
	AcceptableUseURL  *string `json:"acceptable_use_url,omitempty" validate:"omitempty,url"`
	RefundPolicyURL   *string `json:"refund_policy_url,omitempty" validate:"omitempty,url"`

	// Social Media Links
	SocialLinks *models.SocialLinks `json:"social_links,omitempty"`

	// Email Branding
	EmailSettings *models.EmailBranding `json:"email_settings,omitempty"`

	// Advanced Customization
	CustomCSS       *string            `json:"custom_css,omitempty"`
	CustomJS        *string            `json:"custom_js,omitempty"`
	CustomHead      *string            `json:"custom_head,omitempty"`
	CustomMetaTags  *models.CustomMeta `json:"custom_meta_tags,omitempty"`
	CustomAnalytics *models.Analytics  `json:"custom_analytics,omitempty"`

	// Localization
	DefaultLanguage  *string   `json:"default_language,omitempty"`
	SupportedLocales *[]string `json:"supported_locales,omitempty"`
	Timezone         *string   `json:"timezone,omitempty"`
	DateFormat       *string   `json:"date_format,omitempty"`
	TimeFormat       *string   `json:"time_format,omitempty"`
	Currency         *string   `json:"currency,omitempty"`

	// SEO Settings
	SEOSettings *models.SEOConfig `json:"seo_settings,omitempty"`

	// Feature Toggles
	UISettings *models.UISettings `json:"ui_settings,omitempty"`

	// Miscellaneous
	CopyrightText *string `json:"copyright_text,omitempty"`
	PoweredByText *string `json:"powered_by_text,omitempty"`
	HidePoweredBy *bool   `json:"hide_powered_by,omitempty"`

	// Status
	IsActive *bool `json:"is_active,omitempty"`
}

// UpdateColorSchemeRequest represents a request to update only color scheme
type UpdateColorSchemeRequest struct {
	PrimaryColor       string `json:"primary_color" validate:"required,hexcolor"`
	SecondaryColor     string `json:"secondary_color" validate:"required,hexcolor"`
	AccentColor        string `json:"accent_color" validate:"required,hexcolor"`
	BackgroundColor    string `json:"background_color,omitempty" validate:"omitempty,hexcolor"`
	SurfaceColor       string `json:"surface_color,omitempty" validate:"omitempty,hexcolor"`
	TextColor          string `json:"text_color,omitempty" validate:"omitempty,hexcolor"`
	TextSecondaryColor string `json:"text_secondary_color,omitempty" validate:"omitempty,hexcolor"`
	ErrorColor         string `json:"error_color,omitempty" validate:"omitempty,hexcolor"`
	WarningColor       string `json:"warning_color,omitempty" validate:"omitempty,hexcolor"`
	SuccessColor       string `json:"success_color,omitempty" validate:"omitempty,hexcolor"`
	InfoColor          string `json:"info_color,omitempty" validate:"omitempty,hexcolor"`
}

// UpdateBrandingRequest represents a request to update only branding assets
type UpdateBrandingRequest struct {
	LogoURL         string `json:"logo_url" validate:"required,url"`
	LogoDarkURL     string `json:"logo_dark_url,omitempty" validate:"omitempty,url"`
	FaviconURL      string `json:"favicon_url,omitempty" validate:"omitempty,url"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty" validate:"omitempty,url"`
	SplashScreenURL string `json:"splash_screen_url,omitempty" validate:"omitempty,url"`
}

// UpdateDomainRequest represents a request to configure custom domain
type UpdateDomainRequest struct {
	CustomDomain        string `json:"custom_domain" validate:"required,fqdn"`
	CustomDomainEnabled bool   `json:"custom_domain_enabled"`
	SSLEnabled          bool   `json:"ssl_enabled"`
}

// ============================================================================
// WhiteLabel Response DTOs
// ============================================================================

// WhiteLabelResponse represents a whitelabel configuration
type WhiteLabelResponse struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Domain Configuration
	CustomDomain        string `json:"custom_domain,omitempty"`
	CustomDomainEnabled bool   `json:"custom_domain_enabled"`
	SSLEnabled          bool   `json:"ssl_enabled"`

	// Branding Assets
	LogoURL         string `json:"logo_url,omitempty"`
	LogoDarkURL     string `json:"logo_dark_url,omitempty"`
	FaviconURL      string `json:"favicon_url,omitempty"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty"`
	SplashScreenURL string `json:"splash_screen_url,omitempty"`

	// Color Scheme
	PrimaryColor       string `json:"primary_color"`
	SecondaryColor     string `json:"secondary_color"`
	AccentColor        string `json:"accent_color"`
	BackgroundColor    string `json:"background_color"`
	SurfaceColor       string `json:"surface_color"`
	TextColor          string `json:"text_color"`
	TextSecondaryColor string `json:"text_secondary_color"`
	ErrorColor         string `json:"error_color"`
	WarningColor       string `json:"warning_color"`
	SuccessColor       string `json:"success_color"`
	InfoColor          string `json:"info_color"`

	// Typography
	FontFamily        string `json:"font_family"`
	HeadingFontFamily string `json:"heading_font_family,omitempty"`
	FontSize          string `json:"font_size"`
	FontWeight        string `json:"font_weight"`

	// Theme Settings
	Theme           models.ThemeConfig `json:"theme"`
	DarkModeEnabled bool               `json:"dark_mode_enabled"`

	// Company Information
	CompanyName        string `json:"company_name,omitempty"`
	CompanyDescription string `json:"company_description,omitempty"`
	CompanyTagline     string `json:"company_tagline,omitempty"`
	CompanyAddress     string `json:"company_address,omitempty"`
	CompanyPhone       string `json:"company_phone,omitempty"`
	CompanyEmail       string `json:"company_email,omitempty"`
	SupportEmail       string `json:"support_email,omitempty"`
	SupportPhone       string `json:"support_phone,omitempty"`
	SupportURL         string `json:"support_url,omitempty"`

	// Legal & Policy Links
	TermsOfServiceURL string `json:"terms_of_service_url,omitempty"`
	PrivacyPolicyURL  string `json:"privacy_policy_url,omitempty"`
	CookiePolicyURL   string `json:"cookie_policy_url,omitempty"`
	AcceptableUseURL  string `json:"acceptable_use_url,omitempty"`
	RefundPolicyURL   string `json:"refund_policy_url,omitempty"`

	// Social Media Links
	SocialLinks models.SocialLinks `json:"social_links"`

	// Email Branding
	EmailSettings models.EmailBranding `json:"email_settings"`

	// Advanced Customization
	CustomCSS       string            `json:"custom_css,omitempty"`
	CustomJS        string            `json:"custom_js,omitempty"`
	CustomHead      string            `json:"custom_head,omitempty"`
	CustomMetaTags  models.CustomMeta `json:"custom_meta_tags"`
	CustomAnalytics models.Analytics  `json:"custom_analytics"`

	// Localization
	DefaultLanguage  string   `json:"default_language"`
	SupportedLocales []string `json:"supported_locales"`
	Timezone         string   `json:"timezone"`
	DateFormat       string   `json:"date_format"`
	TimeFormat       string   `json:"time_format"`
	Currency         string   `json:"currency"`

	// SEO Settings
	SEOSettings models.SEOConfig `json:"seo_settings"`

	// Feature Toggles
	UISettings models.UISettings `json:"ui_settings"`

	// Miscellaneous
	CopyrightText string `json:"copyright_text,omitempty"`
	PoweredByText string `json:"powered_by_text,omitempty"`
	HidePoweredBy bool   `json:"hide_powered_by"`

	// Status
	IsActive bool `json:"is_active"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublicWhiteLabelResponse represents public-facing whitelabel configuration
// (excludes sensitive information like analytics keys and custom JS/CSS)
type PublicWhiteLabelResponse struct {
	// Branding Assets
	LogoURL         string `json:"logo_url,omitempty"`
	LogoDarkURL     string `json:"logo_dark_url,omitempty"`
	FaviconURL      string `json:"favicon_url,omitempty"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty"`
	SplashScreenURL string `json:"splash_screen_url,omitempty"`

	// Color Scheme
	PrimaryColor       string `json:"primary_color"`
	SecondaryColor     string `json:"secondary_color"`
	AccentColor        string `json:"accent_color"`
	BackgroundColor    string `json:"background_color"`
	SurfaceColor       string `json:"surface_color"`
	TextColor          string `json:"text_color"`
	TextSecondaryColor string `json:"text_secondary_color"`
	ErrorColor         string `json:"error_color"`
	WarningColor       string `json:"warning_color"`
	SuccessColor       string `json:"success_color"`
	InfoColor          string `json:"info_color"`

	// Typography
	FontFamily        string `json:"font_family"`
	HeadingFontFamily string `json:"heading_font_family,omitempty"`
	FontSize          string `json:"font_size"`
	FontWeight        string `json:"font_weight"`

	// Theme Settings
	Theme           models.ThemeConfig `json:"theme"`
	DarkModeEnabled bool               `json:"dark_mode_enabled"`

	// Company Information
	CompanyName        string `json:"company_name,omitempty"`
	CompanyTagline     string `json:"company_tagline,omitempty"`
	CompanyPhone       string `json:"company_phone,omitempty"`
	CompanyEmail       string `json:"company_email,omitempty"`
	SupportEmail       string `json:"support_email,omitempty"`
	SupportPhone       string `json:"support_phone,omitempty"`
	SupportURL         string `json:"support_url,omitempty"`

	// Legal & Policy Links
	TermsOfServiceURL string `json:"terms_of_service_url,omitempty"`
	PrivacyPolicyURL  string `json:"privacy_policy_url,omitempty"`
	CookiePolicyURL   string `json:"cookie_policy_url,omitempty"`
	AcceptableUseURL  string `json:"acceptable_use_url,omitempty"`
	RefundPolicyURL   string `json:"refund_policy_url,omitempty"`

	// Social Media Links
	SocialLinks models.SocialLinks `json:"social_links"`

	// Localization
	DefaultLanguage  string   `json:"default_language"`
	SupportedLocales []string `json:"supported_locales"`
	Timezone         string   `json:"timezone"`
	DateFormat       string   `json:"date_format"`
	TimeFormat       string   `json:"time_format"`
	Currency         string   `json:"currency"`

	// Public SEO
	SEOSettings PublicSEOConfig `json:"seo_settings"`

	// Public UI Settings
	UISettings models.UISettings `json:"ui_settings"`

	// Miscellaneous
	CopyrightText string `json:"copyright_text,omitempty"`
	PoweredByText string `json:"powered_by_text,omitempty"`
	HidePoweredBy bool   `json:"hide_powered_by"`
}

// PublicSEOConfig represents public SEO information
type PublicSEOConfig struct {
	SiteName        string   `json:"site_name,omitempty"`
	SiteDescription string   `json:"site_description,omitempty"`
	SiteKeywords    []string `json:"site_keywords,omitempty"`
}

// ============================================================================
// Conversion Functions
// ============================================================================

// ToWhiteLabelResponse converts a WhiteLabel model to WhiteLabelResponse DTO
func ToWhiteLabelResponse(whitelabel *models.WhiteLabel) *WhiteLabelResponse {
	if whitelabel == nil {
		return nil
	}

	return &WhiteLabelResponse{
		ID:                  whitelabel.ID,
		TenantID:            whitelabel.TenantID,
		CustomDomain:        whitelabel.CustomDomain,
		CustomDomainEnabled: whitelabel.CustomDomainEnabled,
		SSLEnabled:          whitelabel.SSLEnabled,
		LogoURL:             whitelabel.LogoURL,
		LogoDarkURL:         whitelabel.LogoDarkURL,
		FaviconURL:          whitelabel.FaviconURL,
		AppleTouchIcon:      whitelabel.AppleTouchIcon,
		SplashScreenURL:     whitelabel.SplashScreenURL,
		PrimaryColor:        whitelabel.PrimaryColor,
		SecondaryColor:      whitelabel.SecondaryColor,
		AccentColor:         whitelabel.AccentColor,
		BackgroundColor:     whitelabel.BackgroundColor,
		SurfaceColor:        whitelabel.SurfaceColor,
		TextColor:           whitelabel.TextColor,
		TextSecondaryColor:  whitelabel.TextSecondaryColor,
		ErrorColor:          whitelabel.ErrorColor,
		WarningColor:        whitelabel.WarningColor,
		SuccessColor:        whitelabel.SuccessColor,
		InfoColor:           whitelabel.InfoColor,
		FontFamily:          whitelabel.FontFamily,
		HeadingFontFamily:   whitelabel.HeadingFontFamily,
		FontSize:            whitelabel.FontSize,
		FontWeight:          whitelabel.FontWeight,
		Theme:               whitelabel.Theme,
		DarkModeEnabled:     whitelabel.DarkModeEnabled,
		CompanyName:         whitelabel.CompanyName,
		CompanyDescription:  whitelabel.CompanyDescription,
		CompanyTagline:      whitelabel.CompanyTagline,
		CompanyAddress:      whitelabel.CompanyAddress,
		CompanyPhone:        whitelabel.CompanyPhone,
		CompanyEmail:        whitelabel.CompanyEmail,
		SupportEmail:        whitelabel.SupportEmail,
		SupportPhone:        whitelabel.SupportPhone,
		SupportURL:          whitelabel.SupportURL,
		TermsOfServiceURL:   whitelabel.TermsOfServiceURL,
		PrivacyPolicyURL:    whitelabel.PrivacyPolicyURL,
		CookiePolicyURL:     whitelabel.CookiePolicyURL,
		AcceptableUseURL:    whitelabel.AcceptableUseURL,
		RefundPolicyURL:     whitelabel.RefundPolicyURL,
		SocialLinks:         whitelabel.SocialLinks,
		EmailSettings:       whitelabel.EmailSettings,
		CustomCSS:           whitelabel.CustomCSS,
		CustomJS:            whitelabel.CustomJS,
		CustomHead:          whitelabel.CustomHead,
		CustomMetaTags:      whitelabel.CustomMetaTags,
		CustomAnalytics:     whitelabel.CustomAnalytics,
		DefaultLanguage:     whitelabel.DefaultLanguage,
		SupportedLocales:    whitelabel.SupportedLocales,
		Timezone:            whitelabel.Timezone,
		DateFormat:          whitelabel.DateFormat,
		TimeFormat:          whitelabel.TimeFormat,
		Currency:            whitelabel.Currency,
		SEOSettings:         whitelabel.SEOSettings,
		UISettings:          whitelabel.UISettings,
		CopyrightText:       whitelabel.CopyrightText,
		PoweredByText:       whitelabel.PoweredByText,
		HidePoweredBy:       whitelabel.HidePoweredBy,
		IsActive:            whitelabel.IsActive,
		CreatedAt:           whitelabel.CreatedAt,
		UpdatedAt:           whitelabel.UpdatedAt,
	}
}

// ToPublicWhiteLabelResponse converts a WhiteLabel model to PublicWhiteLabelResponse DTO
func ToPublicWhiteLabelResponse(whitelabel *models.WhiteLabel) *PublicWhiteLabelResponse {
	if whitelabel == nil {
		return nil
	}

	return &PublicWhiteLabelResponse{
		LogoURL:             whitelabel.LogoURL,
		LogoDarkURL:         whitelabel.LogoDarkURL,
		FaviconURL:          whitelabel.FaviconURL,
		AppleTouchIcon:      whitelabel.AppleTouchIcon,
		SplashScreenURL:     whitelabel.SplashScreenURL,
		PrimaryColor:        whitelabel.PrimaryColor,
		SecondaryColor:      whitelabel.SecondaryColor,
		AccentColor:         whitelabel.AccentColor,
		BackgroundColor:     whitelabel.BackgroundColor,
		SurfaceColor:        whitelabel.SurfaceColor,
		TextColor:           whitelabel.TextColor,
		TextSecondaryColor:  whitelabel.TextSecondaryColor,
		ErrorColor:          whitelabel.ErrorColor,
		WarningColor:        whitelabel.WarningColor,
		SuccessColor:        whitelabel.SuccessColor,
		InfoColor:           whitelabel.InfoColor,
		FontFamily:          whitelabel.FontFamily,
		HeadingFontFamily:   whitelabel.HeadingFontFamily,
		FontSize:            whitelabel.FontSize,
		FontWeight:          whitelabel.FontWeight,
		Theme:               whitelabel.Theme,
		DarkModeEnabled:     whitelabel.DarkModeEnabled,
		CompanyName:         whitelabel.CompanyName,
		CompanyTagline:      whitelabel.CompanyTagline,
		CompanyPhone:        whitelabel.CompanyPhone,
		CompanyEmail:        whitelabel.CompanyEmail,
		SupportEmail:        whitelabel.SupportEmail,
		SupportPhone:        whitelabel.SupportPhone,
		SupportURL:          whitelabel.SupportURL,
		TermsOfServiceURL:   whitelabel.TermsOfServiceURL,
		PrivacyPolicyURL:    whitelabel.PrivacyPolicyURL,
		CookiePolicyURL:     whitelabel.CookiePolicyURL,
		AcceptableUseURL:    whitelabel.AcceptableUseURL,
		RefundPolicyURL:     whitelabel.RefundPolicyURL,
		SocialLinks:         whitelabel.SocialLinks,
		DefaultLanguage:     whitelabel.DefaultLanguage,
		SupportedLocales:    whitelabel.SupportedLocales,
		Timezone:            whitelabel.Timezone,
		DateFormat:          whitelabel.DateFormat,
		TimeFormat:          whitelabel.TimeFormat,
		Currency:            whitelabel.Currency,
		SEOSettings: PublicSEOConfig{
			SiteName:        whitelabel.SEOSettings.SiteName,
			SiteDescription: whitelabel.SEOSettings.SiteDescription,
			SiteKeywords:    whitelabel.SEOSettings.SiteKeywords,
		},
		UISettings:    whitelabel.UISettings,
		CopyrightText: whitelabel.CopyrightText,
		PoweredByText: whitelabel.PoweredByText,
		HidePoweredBy: whitelabel.HidePoweredBy,
	}
}
