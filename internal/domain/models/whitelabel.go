package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

// WhiteLabel represents tenant-specific branding and customization settings
type WhiteLabel struct {
	BaseModel

	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;uniqueIndex;not null"`
	Tenant   *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`

	// Domain Configuration
	CustomDomain        string `json:"custom_domain,omitempty" gorm:"size:255;uniqueIndex" validate:"omitempty,fqdn"`
	CustomDomainEnabled bool   `json:"custom_domain_enabled" gorm:"default:false"`
	SSLEnabled          bool   `json:"ssl_enabled" gorm:"default:true"`

	// Branding Assets
	LogoURL         string `json:"logo_url,omitempty" gorm:"size:500"`
	LogoDarkURL     string `json:"logo_dark_url,omitempty" gorm:"size:500"` // For dark mode
	FaviconURL      string `json:"favicon_url,omitempty" gorm:"size:500"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty" gorm:"size:500"`
	SplashScreenURL string `json:"splash_screen_url,omitempty" gorm:"size:500"`

	// Color Scheme
	PrimaryColor       string `json:"primary_color,omitempty" gorm:"size:7;default:'#3B82F6'" validate:"omitempty,hexcolor"`
	SecondaryColor     string `json:"secondary_color,omitempty" gorm:"size:7;default:'#10B981'" validate:"omitempty,hexcolor"`
	AccentColor        string `json:"accent_color,omitempty" gorm:"size:7;default:'#F59E0B'" validate:"omitempty,hexcolor"`
	BackgroundColor    string `json:"background_color,omitempty" gorm:"size:7;default:'#FFFFFF'" validate:"omitempty,hexcolor"`
	SurfaceColor       string `json:"surface_color,omitempty" gorm:"size:7;default:'#F9FAFB'" validate:"omitempty,hexcolor"`
	TextColor          string `json:"text_color,omitempty" gorm:"size:7;default:'#111827'" validate:"omitempty,hexcolor"`
	TextSecondaryColor string `json:"text_secondary_color,omitempty" gorm:"size:7;default:'#6B7280'" validate:"omitempty,hexcolor"`
	ErrorColor         string `json:"error_color,omitempty" gorm:"size:7;default:'#EF4444'" validate:"omitempty,hexcolor"`
	WarningColor       string `json:"warning_color,omitempty" gorm:"size:7;default:'#F59E0B'" validate:"omitempty,hexcolor"`
	SuccessColor       string `json:"success_color,omitempty" gorm:"size:7;default:'#10B981'" validate:"omitempty,hexcolor"`
	InfoColor          string `json:"info_color,omitempty" gorm:"size:7;default:'#3B82F6'" validate:"omitempty,hexcolor"`

	// Typography
	FontFamily        string `json:"font_family,omitempty" gorm:"size:255;default:'Inter, system-ui, sans-serif'"`
	HeadingFontFamily string `json:"heading_font_family,omitempty" gorm:"size:255"`
	FontSize          string `json:"font_size,omitempty" gorm:"size:10;default:'16px'"`
	FontWeight        string `json:"font_weight,omitempty" gorm:"size:10;default:'400'"`

	// Theme Settings
	Theme           ThemeConfig `json:"theme" gorm:"type:jsonb"`
	DarkModeEnabled bool        `json:"dark_mode_enabled" gorm:"default:false"`

	// Company Information
	CompanyName        string `json:"company_name,omitempty" gorm:"size:255"`
	CompanyDescription string `json:"company_description,omitempty" gorm:"type:text"`
	CompanyTagline     string `json:"company_tagline,omitempty" gorm:"size:255"`
	CompanyAddress     string `json:"company_address,omitempty" gorm:"type:text"`
	CompanyPhone       string `json:"company_phone,omitempty" gorm:"size:20"`
	CompanyEmail       string `json:"company_email,omitempty" gorm:"size:255" validate:"omitempty,email"`
	SupportEmail       string `json:"support_email,omitempty" gorm:"size:255" validate:"omitempty,email"`
	SupportPhone       string `json:"support_phone,omitempty" gorm:"size:20"`
	SupportURL         string `json:"support_url,omitempty" gorm:"size:500" validate:"omitempty,url"`

	// Legal & Policy Links
	TermsOfServiceURL string `json:"terms_of_service_url,omitempty" gorm:"size:500" validate:"omitempty,url"`
	PrivacyPolicyURL  string `json:"privacy_policy_url,omitempty" gorm:"size:500" validate:"omitempty,url"`
	CookiePolicyURL   string `json:"cookie_policy_url,omitempty" gorm:"size:500" validate:"omitempty,url"`
	AcceptableUseURL  string `json:"acceptable_use_url,omitempty" gorm:"size:500" validate:"omitempty,url"`
	RefundPolicyURL   string `json:"refund_policy_url,omitempty" gorm:"size:500" validate:"omitempty,url"`

	// Social Media Links
	SocialLinks SocialLinks `json:"social_links" gorm:"type:jsonb"`

	// Email Branding
	EmailSettings EmailBranding `json:"email_settings" gorm:"type:jsonb"`

	// Advanced Customization
	CustomCSS       string     `json:"custom_css,omitempty" gorm:"type:text"`
	CustomJS        string     `json:"custom_js,omitempty" gorm:"type:text"`
	CustomHead      string     `json:"custom_head,omitempty" gorm:"type:text"` // Custom HTML in <head>
	CustomMetaTags  CustomMeta `json:"custom_meta_tags" gorm:"type:jsonb"`
	CustomAnalytics Analytics  `json:"custom_analytics" gorm:"type:jsonb"`

	// Localization
	DefaultLanguage  string   `json:"default_language" gorm:"size:10;default:'en'"`
	SupportedLocales []string `json:"supported_locales" gorm:"type:text[]"`
	Timezone         string   `json:"timezone" gorm:"size:50;default:'UTC'"`
	DateFormat       string   `json:"date_format" gorm:"size:50;default:'YYYY-MM-DD'"`
	TimeFormat       string   `json:"time_format" gorm:"size:50;default:'HH:mm'"`
	Currency         string   `json:"currency" gorm:"size:3;default:'USD'"`

	// SEO Settings
	SEOSettings SEOConfig `json:"seo_settings" gorm:"type:jsonb"`

	// Feature Toggles (UI specific)
	UISettings UISettings `json:"ui_settings" gorm:"type:jsonb"`

	// Miscellaneous
	CopyrightText string `json:"copyright_text,omitempty" gorm:"size:255"`
	PoweredByText string `json:"powered_by_text,omitempty" gorm:"size:255"`
	HidePoweredBy bool   `json:"hide_powered_by" gorm:"default:false"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`
}

// ThemeConfig defines theme customization settings
type ThemeConfig struct {
	BorderRadius    string `json:"border_radius" validate:"omitempty"`    // "8px", "0.5rem"
	BoxShadow       string `json:"box_shadow" validate:"omitempty"`       // Shadow style
	ButtonStyle     string `json:"button_style" validate:"omitempty"`     // "rounded", "square", "pill"
	InputStyle      string `json:"input_style" validate:"omitempty"`      // "outlined", "filled", "underlined"
	LayoutStyle     string `json:"layout_style" validate:"omitempty"`     // "boxed", "fluid"
	SidebarPosition string `json:"sidebar_position" validate:"omitempty"` // "left", "right"
	HeaderStyle     string `json:"header_style" validate:"omitempty"`     // "sticky", "fixed", "static"
	NavigationStyle string `json:"navigation_style" validate:"omitempty"` // "horizontal", "vertical"
	CardStyle       string `json:"card_style" validate:"omitempty"`       // "elevated", "outlined", "flat"
	AnimationSpeed  string `json:"animation_speed" validate:"omitempty"`  // "fast", "normal", "slow"
	Spacing         string `json:"spacing" validate:"omitempty"`          // "compact", "normal", "comfortable"
}

// SocialLinks defines social media profile links
type SocialLinks struct {
	Facebook  string `json:"facebook,omitempty" validate:"omitempty,url"`
	Twitter   string `json:"twitter,omitempty" validate:"omitempty,url"`
	Instagram string `json:"instagram,omitempty" validate:"omitempty,url"`
	LinkedIn  string `json:"linkedin,omitempty" validate:"omitempty,url"`
	YouTube   string `json:"youtube,omitempty" validate:"omitempty,url"`
	TikTok    string `json:"tiktok,omitempty" validate:"omitempty,url"`
	Pinterest string `json:"pinterest,omitempty" validate:"omitempty,url"`
	GitHub    string `json:"github,omitempty" validate:"omitempty,url"`
	Website   string `json:"website,omitempty" validate:"omitempty,url"`
}

// EmailBranding defines email template branding settings
type EmailBranding struct {
	LogoURL         string `json:"logo_url,omitempty" validate:"omitempty,url"`
	HeaderColor     string `json:"header_color,omitempty" validate:"omitempty,hexcolor"`
	FooterColor     string `json:"footer_color,omitempty" validate:"omitempty,hexcolor"`
	ButtonColor     string `json:"button_color,omitempty" validate:"omitempty,hexcolor"`
	ButtonTextColor string `json:"button_text_color,omitempty" validate:"omitempty,hexcolor"`
	FontFamily      string `json:"font_family,omitempty"`
	FromName        string `json:"from_name,omitempty"`
	FromEmail       string `json:"from_email,omitempty" validate:"omitempty,email"`
	ReplyToEmail    string `json:"reply_to_email,omitempty" validate:"omitempty,email"`
	FooterText      string `json:"footer_text,omitempty"`
	UnsubscribeText string `json:"unsubscribe_text,omitempty"`
}

// CustomMeta defines custom meta tags for SEO and social sharing
type CustomMeta struct {
	Title          string            `json:"title,omitempty"`
	Description    string            `json:"description,omitempty"`
	Keywords       []string          `json:"keywords,omitempty"`
	Author         string            `json:"author,omitempty"`
	OGTitle        string            `json:"og_title,omitempty"`
	OGDescription  string            `json:"og_description,omitempty"`
	OGImage        string            `json:"og_image,omitempty" validate:"omitempty,url"`
	OGType         string            `json:"og_type,omitempty"`
	TwitterCard    string            `json:"twitter_card,omitempty"`
	TwitterSite    string            `json:"twitter_site,omitempty"`
	TwitterCreator string            `json:"twitter_creator,omitempty"`
	AdditionalMeta map[string]string `json:"additional_meta,omitempty"`
}

// Analytics defines analytics and tracking integrations
type Analytics struct {
	GoogleAnalyticsID  string   `json:"google_analytics_id,omitempty"`
	GoogleTagManagerID string   `json:"google_tag_manager_id,omitempty"`
	FacebookPixelID    string   `json:"facebook_pixel_id,omitempty"`
	HotjarID           string   `json:"hotjar_id,omitempty"`
	MixpanelID         string   `json:"mixpanel_id,omitempty"`
	SegmentWriteKey    string   `json:"segment_write_key,omitempty"`
	IntercomAppID      string   `json:"intercom_app_id,omitempty"`
	CustomScripts      []string `json:"custom_scripts,omitempty"` // Additional tracking scripts
}

// SEOConfig defines SEO optimization settings
type SEOConfig struct {
	SiteName          string            `json:"site_name,omitempty"`
	SiteDescription   string            `json:"site_description,omitempty"`
	SiteKeywords      []string          `json:"site_keywords,omitempty"`
	RobotsConfig      string            `json:"robots_config,omitempty"` // "index,follow", "noindex,nofollow"
	CanonicalURL      string            `json:"canonical_url,omitempty" validate:"omitempty,url"`
	SitemapURL        string            `json:"sitemap_url,omitempty" validate:"omitempty,url"`
	VerificationCodes map[string]string `json:"verification_codes,omitempty"` // Google, Bing, etc.
}

// UISettings defines UI-specific feature toggles and settings
type UISettings struct {
	ShowSearch           bool   `json:"show_search"`
	ShowNotifications    bool   `json:"show_notifications"`
	ShowUserProfile      bool   `json:"show_user_profile"`
	ShowLanguageSelector bool   `json:"show_language_selector"`
	ShowHelpCenter       bool   `json:"show_help_center"`
	ShowChatWidget       bool   `json:"show_chat_widget"`
	CompactMode          bool   `json:"compact_mode"`
	FullPageLayout       bool   `json:"full_page_layout"`
	StickyHeader         bool   `json:"sticky_header"`
	Breadcrumbs          bool   `json:"breadcrumbs"`
	LoadingAnimation     string `json:"loading_animation,omitempty"` // "spinner", "skeleton", "progress"
}

// Scan implementations for JSONB fields
func (t *ThemeConfig) Scan(value any) error {
	if value == nil {
		*t = ThemeConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ThemeConfig")
	}
	return json.Unmarshal(bytes, t)
}

func (t ThemeConfig) Value() (driver.Value, error) {
	if (t == ThemeConfig{}) {
		return nil, nil
	}
	return json.Marshal(t)
}

func (s *SocialLinks) Scan(value any) error {
	if value == nil {
		*s = SocialLinks{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SocialLinks")
	}
	return json.Unmarshal(bytes, s)
}

func (s SocialLinks) Value() (driver.Value, error) {
	if (s == SocialLinks{}) {
		return nil, nil
	}
	return json.Marshal(s)
}

func (e *EmailBranding) Scan(value any) error {
	if value == nil {
		*e = EmailBranding{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan EmailBranding")
	}
	return json.Unmarshal(bytes, e)
}

func (e EmailBranding) Value() (driver.Value, error) {
	if (e == EmailBranding{}) {
		return nil, nil
	}
	return json.Marshal(e)
}

func (c *CustomMeta) Scan(value any) error {
	if value == nil {
		*c = CustomMeta{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CustomMeta")
	}
	return json.Unmarshal(bytes, c)
}

func (c CustomMeta) Value() (driver.Value, error) {
	// Check if struct is empty by marshaling and checking for empty JSON object
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

func (a *Analytics) Scan(value any) error {
	if value == nil {
		*a = Analytics{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Analytics")
	}
	return json.Unmarshal(bytes, a)
}

func (a Analytics) Value() (driver.Value, error) {
	// Check if struct is empty by marshaling and checking for empty JSON object
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

func (s *SEOConfig) Scan(value any) error {
	if value == nil {
		*s = SEOConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan SEOConfig")
	}
	return json.Unmarshal(bytes, s)
}

func (s SEOConfig) Value() (driver.Value, error) {
	// Check if struct is empty by marshaling and checking for empty JSON object
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	if string(data) == "{}" || string(data) == "null" {
		return nil, nil
	}
	return data, nil
}

func (u *UISettings) Scan(value any) error {
	if value == nil {
		*u = UISettings{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan UISettings")
	}
	return json.Unmarshal(bytes, u)
}

func (u UISettings) Value() (driver.Value, error) {
	if (u == UISettings{}) {
		return nil, nil
	}
	return json.Marshal(u)
}

// TableName specifies the table name for the WhiteLabel model
func (WhiteLabel) TableName() string {
	return "white_labels"
}
