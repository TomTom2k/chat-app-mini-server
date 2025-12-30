# Chat App Server

Server backend cho ứng dụng chat được xây dựng với Go và Clean Architecture.

## Cấu trúc thư mục

```
server/
├── cmd/app/              # Entry point của ứng dụng
├── internal/
│   ├── config/           # Quản lý cấu hình
│   ├── domain/           # Domain layer (entities, interfaces)
│   │   ├── entity/       # Domain entities
│   │   └── repository.go # Repository interfaces
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── di/          # Dependency Injection container
│   │   ├── mongodb/     # MongoDB connection
│   │   ├── repository/  # Repository implementations
│   │   └── server/      # HTTP server setup
│   ├── interface/       # Interface layer
│   │   └── http/        # HTTP handlers
│   └── usecase/         # Use case layer (business logic)
└── pkg/                  # Shared packages
    ├── jwt/             # JWT utilities
    └── utils/           # Utility functions
```

## Clean Architecture

Dự án tuân thủ Clean Architecture với các layer:

1. **Domain Layer**: Chứa entities và repository interfaces (không phụ thuộc vào framework)
2. **Use Case Layer**: Chứa business logic
3. **Interface Layer**: HTTP handlers, API endpoints
4. **Infrastructure Layer**: Database, external services, DI container

## Cài đặt

1. Cài đặt dependencies:
```bash
go mod download
```

2. Cài đặt các package cần thiết:
```bash
go get github.com/gin-contrib/cors
go get github.com/golang-jwt/jwt/v5
```

3. Tạo file `.env`:
```env
SERVER_PORT=8080
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=chat_app
JWT_SECRET=your-secret-key-change-in-production
ENVIRONMENT=development
```

4. Chạy server:
```bash
go run cmd/app/main.go
```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Đăng ký user mới
- `POST /api/auth/login` - Đăng nhập

Xem chi tiết trong `API_DOCUMENTATION.md`

## Dependency Injection

Tất cả dependencies được quản lý trong `internal/infrastructure/di/container.go`:

- Repository được khởi tạo và inject vào UseCase
- UseCase được inject vào Handler
- Config được truyền qua các layer cần thiết

## Notes

- MongoDB connection được quản lý tập trung trong `internal/infrastructure/mongodb`
- JWT secret được truyền qua config, không hardcode
- Server setup và routes được tách riêng trong `internal/infrastructure/server`




