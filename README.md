# Krafti Vibe

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Fiber Version](https://img.shields.io/badge/Fiber-v2.52-00ACD7?style=flat)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> Enterprise-grade Multi-Tenant SaaS Platform for Artisan Service Management

Krafti Vibe is a comprehensive backend platform designed to power artisan service businesses. Built with Go and Fiber, it provides robust multi-tenancy, role-based access control, and a complete suite of features for managing bookings, payments, communications, and analytics.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Documentation](#documentation)
- [API Reference](#api-reference)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

## Features

### Core Capabilities

#### Multi-Tenancy & Access Control
- **Row-Level Security**: PostgreSQL RLS for complete data isolation
- **Flexible Tenant Types**: Support for solo artisans, small businesses, and corporations
- **Role-Based Access**: Comprehensive RBAC with 5+ distinct roles
- **JWT Authentication**: Secure token-based auth with refresh tokens

#### Booking Management
- **Intelligent Scheduling**: Real-time availability with conflict detection
- **Recurring Bookings**: Support for weekly/monthly appointments
- **Flexible Pricing**: Service packages, add-ons, and dynamic pricing
- **Deposit Handling**: Secure payment holds and processing
- **Status Workflow**: Complete booking lifecycle management

#### Payment Processing
- **Multiple Providers**: Stripe, PayPal, and extensible gateway support
- **Split Payments**: Automated commission and revenue sharing
- **Refund Management**: Policy-based automated refunds
- **Invoice Generation**: Professional PDF invoices with tax calculation
- **Subscription Billing**: Recurring charges for platform tenants

#### Communication Hub
- **In-App Messaging**: Real-time chat between customers and artisans
- **Multi-Channel Notifications**: Email, SMS, and push notifications
- **Template Engine**: Dynamic content with variable substitution
- **Delivery Tracking**: Read receipts and bounce handling
- **User Preferences**: Granular notification control

#### Reviews & Ratings
- **Multi-Dimensional Ratings**: Quality, professionalism, value, and timeliness
- **Photo Reviews**: Before/after image support
- **Artisan Responses**: Engage with customer feedback
- **Moderation Tools**: Flag inappropriate content
- **Helpful Voting**: Community-driven review quality

#### Analytics & Reporting
- **Real-Time Dashboards**: Live metrics and KPIs
- **Custom Reports**: Scheduled and on-demand report generation
- **Export Formats**: PDF, Excel, CSV support
- **Tenant Analytics**: Usage patterns and business insights
- **Revenue Tracking**: Comprehensive financial reporting

#### File Management
- **Secure Uploads**: Validated and sanitized file handling
- **Image Processing**: Automatic resizing and optimization
- **Document Storage**: Invoices, certificates, and contracts
- **CDN Integration**: Fast global content delivery
- **Virus Scanning**: Automated malware detection

#### Promotional Tools
- **Discount Codes**: Percentage and fixed-amount discounts
- **Usage Limits**: Per-user and total redemption caps
- **Date Restrictions**: Time-bound promotional campaigns
- **Service/Artisan Specific**: Targeted promotions
- **Analytics**: Track campaign performance

## Architecture

### System Design

```
┌─────────────────────────────────────────────────────────────┐
│                       API Gateway                            │
│                    (Fiber Middleware)                        │
├─────────────────────────────────────────────────────────────┤
│  Auth │ Tenant Isolation │ Rate Limiting │ Logging          │
└────────────────────┬────────────────────────────────────────┘
                     │
      ┌──────────────┼──────────────┐
      │              │              │
┌─────▼─────┐  ┌────▼────┐  ┌─────▼─────┐
│  Service  │  │ Service │  │  Service  │
│   Layer   │  │  Layer  │  │   Layer   │
└─────┬─────┘  └────┬────┘  └─────┬─────┘
      │              │              │
      └──────────────┼──────────────┘
                     │
            ┌────────▼────────┐
            │   Repository    │
            │      Layer      │
            └────────┬────────┘
                     │
      ┌──────────────┼──────────────┐
      │              │              │
┌─────▼─────┐  ┌────▼────┐  ┌─────▼─────┐
│PostgreSQL │  │  Redis  │  │   File    │
│    DB     │  │  Cache  │  │  Storage  │
└───────────┘  └─────────┘  └───────────┘
```

### Multi-Tenant Model

```
Platform
├── Tenant A (Organization)
│   ├── Admin Users
│   ├── Artisan Users
│   │   └── Member Users
│   └── Customer Users
├── Tenant B (Organization)
│   └── ...
└── Super Admin (Platform Level)
```

### Key Design Patterns

- **Repository Pattern**: Clean separation of data access logic
- **Service Layer**: Business logic encapsulation
- **DTO Pattern**: Request/response transformation
- **Middleware Chain**: Request processing pipeline
- **Event-Driven**: Async operations via background jobs

## Tech Stack

### Backend
- **Framework**: [Fiber v2](https://gofiber.io/) - Express-inspired web framework
- **Language**: Go 1.24+
- **ORM**: [GORM](https://gorm.io/) - Developer-friendly ORM
- **Database**: PostgreSQL 15+ with Row-Level Security
- **Cache**: Redis 7+ for session and query caching
- **Auth**: JWT with Logto integration support
- **Validation**: Built-in request validation
- **Logging**: Structured logging with Zap

### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Monitoring**: Prometheus metrics
- **Hot Reload**: Air for development
- **API Docs**: OpenAPI/Swagger (planned)

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/Krafti_Vibe.git
   cd Krafti_Vibe
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run database migrations**
   ```bash
   # Migration commands here
   ```

5. **Start the development server**
   ```bash
   # Using Air (hot reload)
   air

   # Or standard go run
   go run cmd/api/main.go
   ```

6. **Access the API**
   ```
   http://localhost:8080
   ```

### Docker Setup

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Environment Configuration

Key environment variables:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=krafti_vibe
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# File Storage
STORAGE_DRIVER=local
STORAGE_PATH=./uploads
MAX_UPLOAD_SIZE=10485760

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@kraftivibe.com

# Payment Gateways
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
PAYPAL_CLIENT_ID=
PAYPAL_CLIENT_SECRET=
```

## Project Structure

```
Krafti_Vibe/
├── cmd/
│   └── api/                    # Application entry point
├── internal/
│   ├── auth/                   # Authentication logic
│   ├── config/                 # Configuration management
│   ├── domain/
│   │   └── models/            # Domain models/entities
│   ├── infrastructure/         # External integrations
│   ├── middleware/            # HTTP middlewares
│   ├── pkg/
│   │   └── errors/            # Custom error handling
│   ├── repository/            # Data access layer
│   │   └── *.go               # Repository implementations
│   ├── router/                # Route definitions
│   └── service/               # Business logic layer
│       ├── dto/               # Data transfer objects
│       ├── enterprise/        # Enterprise features
│       └── *.go               # Service implementations
├── scripts/                    # Utility scripts
├── docs/                       # Documentation
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── air.toml                   # Hot reload config
└── go.mod
```

## Documentation

- **[Project Specification](PROJECT_SPEC.md)** - Comprehensive technical specification
- **[Quick Start Guide](QUICKSTART.md)** - Get up and running quickly
- **[API Documentation](docs/API.md)** - REST API reference
- **[Service Implementation](docs/service_implementation_summary.md)** - Service layer details
- **[Logto Integration](docs/logto_integration_guide.md)** - Authentication setup
- **[Repository Completion](REPOSITORY_COMPLETION.md)** - Repository layer status

### Service-Specific Documentation

- [File Upload Service](docs/file_upload_service.md)
- [Message Service](docs/message_service.md)
- [Enterprise Features](internal/service/enterprise/README.md)

## API Reference

### Authentication Endpoints

```http
POST   /api/v1/auth/login       # User login
POST   /api/v1/auth/register    # User registration
POST   /api/v1/auth/refresh     # Refresh access token
POST   /api/v1/auth/logout      # User logout
```

### Booking Endpoints

```http
GET    /api/v1/bookings         # List bookings
POST   /api/v1/bookings         # Create booking
GET    /api/v1/bookings/:id     # Get booking details
PATCH  /api/v1/bookings/:id     # Update booking
DELETE /api/v1/bookings/:id     # Cancel booking
```

### Artisan Endpoints

```http
GET    /api/v1/artisans                    # List artisans
GET    /api/v1/artisans/:id                # Get artisan profile
GET    /api/v1/artisans/:id/availability   # Get availability
GET    /api/v1/artisans/:id/reviews        # Get reviews
GET    /api/v1/artisans/:id/services       # Get services
```

### Payment Endpoints

```http
GET    /api/v1/payments                # List payments
POST   /api/v1/payments                # Create payment
GET    /api/v1/payments/:id            # Get payment details
POST   /api/v1/payments/:id/refund     # Process refund
GET    /api/v1/payments/methods        # List payment methods
```

For complete API documentation, see [docs/API.md](docs/API.md).

## Development

### Code Style

This project follows standard Go conventions and style guidelines:

- Use `gofmt` for code formatting
- Follow effective Go best practices
- Write meaningful commit messages
- Add tests for new features

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/service/...

# Run with verbose output
go test -v ./...
```

### Building

```bash
# Build for current platform
go build -o bin/krafti-vibe cmd/api/main.go

# Build for production
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/krafti-vibe-linux cmd/api/main.go
```

### Database Migrations

```bash
# Create new migration
make migration name=create_users_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status
```

## Testing

### Unit Tests

```bash
# Run service layer tests
go test ./internal/service/...

# Run repository layer tests
go test ./internal/repository/...
```

### Integration Tests

```bash
# Run integration tests (requires Docker)
make test-integration
```

### Load Testing

```bash
# Run load tests
make test-load
```

## Deployment

### Production Build

```bash
# Build optimized binary
make build-prod

# Create Docker image
docker build -t krafti-vibe:latest .
```

### Environment Setup

1. Set up PostgreSQL database with replication
2. Configure Redis cluster for high availability
3. Set up file storage (S3 or equivalent)
4. Configure CDN for static assets
5. Set up monitoring and alerting
6. Configure backup strategy

### Container Deployment

```bash
# Deploy with Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# Deploy to Kubernetes
kubectl apply -f k8s/
```

### Performance Tuning

- Enable database connection pooling
- Configure Redis caching strategy
- Set up CDN for static assets
- Enable response compression
- Implement rate limiting per tenant tier

## Monitoring & Observability

### Metrics

The application exposes Prometheus metrics at `/metrics`:

- Request latency and throughput
- Database query performance
- Cache hit/miss rates
- Business metrics (bookings, revenue, etc.)

### Logging

Structured JSON logging with context:

```go
logger.Info("booking created",
    "booking_id", bookingID,
    "tenant_id", tenantID,
    "artisan_id", artisanID,
)
```

### Health Checks

```http
GET /health       # Application health
GET /ready        # Readiness probe
```

## Security

### Implemented Security Measures

- **Authentication**: JWT-based with refresh tokens
- **Authorization**: Role-based access control (RBAC)
- **Data Isolation**: PostgreSQL Row-Level Security
- **Input Validation**: Request validation middleware
- **SQL Injection**: Parameterized queries via GORM
- **XSS Protection**: Output sanitization
- **Rate Limiting**: Per-tenant API throttling
- **CORS**: Configurable cross-origin policies
- **Secure Headers**: Security headers middleware

### Security Best Practices

- Regularly update dependencies
- Use environment variables for secrets
- Enable HTTPS in production
- Implement audit logging
- Regular security audits
- Follow OWASP guidelines

## Performance

### Optimization Strategies

- **Database**: Connection pooling, indexes, query optimization
- **Caching**: Redis for hot data and session storage
- **Pagination**: Cursor-based for large datasets
- **Compression**: Gzip/Brotli response compression
- **CDN**: Static asset delivery
- **Background Jobs**: Async processing for heavy operations

### Benchmarks

```
BenchmarkBookingCreate     5000  250000 ns/op  1024 B/op  10 allocs/op
BenchmarkBookingList      10000  150000 ns/op   512 B/op   5 allocs/op
BenchmarkPaymentProcess    3000  400000 ns/op  2048 B/op  15 allocs/op
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md).

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Review Process

- All PRs require at least one approval
- All tests must pass
- Code coverage should not decrease
- Follow existing code style

## Roadmap

### Phase 1: MVP (Completed)
- ✅ Multi-tenant architecture
- ✅ User authentication and authorization
- ✅ Booking management system
- ✅ Payment processing
- ✅ Basic notifications
- ✅ Reviews and ratings

### Phase 2: Enhanced Features (In Progress)
- 🚧 WebSocket real-time updates
- 🚧 Advanced analytics dashboard
- 🚧 Mobile app API enhancements
- 🚧 Recurring booking improvements
- 🚧 Advanced reporting

### Phase 3: Enterprise (Planned)
- ⏳ White-labeling support
- ⏳ API marketplace
- ⏳ Advanced integrations
- ⏳ Multi-language support
- ⏳ Compliance certifications

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/yourusername/Krafti_Vibe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/Krafti_Vibe/discussions)
- **Email**: support@kraftivibe.com

## Acknowledgments

- Built with [Fiber](https://gofiber.io/)
- Powered by [GORM](https://gorm.io/)
- Inspired by best practices from the Go community

---

**Made with ❤️ by the Krafti Vibe Team**

**Version**: 1.0.0
**Last Updated**: December 2025
