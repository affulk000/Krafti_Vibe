# Krafti Vibe

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Fiber Version](https://img.shields.io/badge/Fiber-v2.52-00ACD7?style=flat)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **Multi-Tenant SaaS Platform for Artisan Service Businesses**

Krafti Vibe is a complete backend platform purpose-built for artisan service businesses - from solo craftspeople to large service organizations. Built with Go and Fiber, it delivers enterprise-grade multi-tenancy, intelligent booking management, and comprehensive business operations in a single platform.

## 🎯 What Makes Krafti Vibe Different

**The Niche**: Traditional booking platforms fall short for artisan businesses. They either oversimplify (missing critical features like materials tracking, project milestones, custom pricing) or over-complicate (enterprise tools that cost too much and do too much). Krafti Vibe fills this gap perfectly.

**Built For**:
- 🔨 Home services (plumbers, electricians, cleaners)
- ✂️ Beauty & wellness (salons, spas, personal trainers)
- 🎨 Creative services (photographers, designers, decorators)
- 🔧 Repair & maintenance (appliance repair, handymen, locksmiths)
- 💆 Health & fitness (massage therapists, yoga instructors, physiotherapists)

## ✨ Core Features

### 🏢 Multi-Tenancy
- **Complete Isolation**: Row-level security ensures tenant data never leaks
- **Flexible Models**: Support solo artisans, small teams, or large organizations
- **White-Label Ready**: Custom domains and branding per tenant
- **Tiered Plans**: Free, Pro, Enterprise with usage-based billing

### 📅 Intelligent Booking System
- Real-time availability with conflict detection
- Recurring appointments (daily/weekly/monthly)
- Service packages & add-ons with dynamic pricing
- Deposit handling & payment holds
- Before/after photo documentation
- Customer notes & artisan instructions

### 💳 Payment Processing
- Multiple gateways (Stripe, PayPal)
- Automated commission splits
- Refund management with policies
- Professional invoice generation
- Subscription billing
- Revenue analytics & reporting

### 💬 Communication Hub
- In-app messaging between customers & artisans
- Multi-channel notifications (email, SMS, push)
- Template engine with dynamic variables
- Delivery tracking & read receipts
- Granular notification preferences

### ⭐ Reviews & Reputation
- Multi-dimensional ratings (quality, professionalism, timeliness, value)
- Photo reviews with before/after comparisons
- Artisan response system
- Review moderation & flagging
- Community helpful voting

### 📊 Business Intelligence
- Real-time dashboards & KPIs
- Custom report generation (PDF, Excel, CSV)
- Revenue tracking & forecasting
- Customer lifetime value analytics
- Artisan performance metrics
- Usage patterns & trends

### 🎁 Marketing Tools
- Discount codes (percentage & fixed)
- Usage limits & redemption tracking
- Date-restricted campaigns
- Service/artisan-specific promotions
- Campaign performance analytics

### 🗂️ Project Management
- Multi-phase project tracking
- Milestone-based payments
- Task assignments & dependencies
- Progress monitoring
- Client collaboration tools
- Document management

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 15+
- Redis 7+

### Installation

```bash
# Clone repository
git clone https://github.com/affulk000/Krafti_Vibe.git
cd Krafti_Vibe

# Install dependencies
go mod download

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Run with hot reload
air

# Or standard run
go run cmd/api/main.go
```

### Docker Setup

```bash
docker-compose up -d
```

Server runs at `http://localhost:8080`

## 🏗️ Architecture

Built with clean architecture principles:

```
API Layer (Fiber)
    ↓
Service Layer (Business Logic)
    ↓
Repository Layer (Data Access)
    ↓
Database (PostgreSQL + Redis)
```

**Key Technologies**:
- **Framework**: Fiber v2 (high-performance web framework)
- **Database**: PostgreSQL with Row-Level Security
- **Cache**: Redis for sessions & hot data
- **Auth**: Zitadel integration with JWT
- **ORM**: GORM with type-safe operations
- **Logging**: Structured logging with Zap

## 📦 Project Structure

```
Krafti_Vibe/
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/models/    # Business entities
│   ├── repository/       # Data access (13 repos, 200+ methods)
│   ├── service/          # Business logic (24 services)
│   ├── middleware/       # Auth, logging, rate limiting
│   └── router/           # Route definitions
├── scripts/              # Utilities & helpers
└── docs/                 # Documentation
```

## 📚 Documentation

### General
- [Project Specification](PROJECT_SPEC.md) - Complete technical spec
- [Quick Start Guide](QUICKSTART.md) - Detailed setup instructions
- [Migration Guide](MIGRATIONS.md) - Database migrations
- [Zitadel Auth](ZITADEL_AUTH_STATUS.md) - Authentication setup

### API Documentation
- **[Swagger UI](http://localhost:8080/swagger/)** - Interactive API documentation (requires running server)
- [API Quick Reference](docs/API_QUICK_REFERENCE.md) - Common endpoints & examples
- [Swagger Guide](docs/SWAGGER_GUIDE.md) - Complete API documentation guide
- [OpenAPI Spec](docs/swagger.yaml) - Machine-readable API specification

## 🔒 Security

- JWT authentication with refresh tokens
- Role-based access control (RBAC)
- PostgreSQL Row-Level Security
- Input validation & sanitization
- SQL injection protection
- XSS prevention
- CORS configuration
- Rate limiting per tenant
- Secure headers

## 📈 Current Status

| Component | Status |
|-----------|--------|
| **Domain Models** | ✅ 100% Complete (20+ models) |
| **Repository Layer** | ✅ 100% Complete (200+ methods) |
| **Service Layer** | ✅ 100% Complete (24 services) |
| **Middleware** | ✅ 100% Complete |
| **API Handlers** | 🚧 In Progress |
| **Authentication** | 🚧 30% Complete |
| **Testing** | 🚧 In Progress |

**Overall**: ~60% complete (core backend done, API layer in progress)

## 🛣️ Roadmap

### Next Up
- [ ] Complete REST API handlers
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Database migrations
- [ ] Comprehensive testing

### Future
- [ ] WebSocket real-time updates
- [ ] Background job processing
- [ ] Email & SMS integrations
- [ ] Calendar sync (Google, iCal)
- [ ] Mobile app APIs
- [ ] GraphQL endpoint (optional)
- [ ] AI-powered scheduling
- [ ] Fraud detection
- [ ] Multi-language support

## 💻 Development

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build for production
make build

# Run linter
golangci-lint run
```

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with:
- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [GORM](https://gorm.io/) - Feature-rich ORM
- [Zitadel](https://zitadel.com/) - Identity & access management

---

**Made with ❤️ for artisan businesses worldwide**

Version: 1.0.0 | Last Updated: December 2024
