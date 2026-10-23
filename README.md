# Barber One API - Modular Architecture

## Deskripsi

Proyek backend Barber One yang telah direfactor menjadi modular architecture dengan struktur folder yang lebih terorganisir dan mudah dipelihara.

## Struktur Folder

```
barber-backend/
├── internal/              # Internal application code
│   ├── config/           # Configuration management
│   │   ├── config.go
│   │   └── database.go
│   ├── controller/       # HTTP handlers (MVC Controller)
│   │   ├── auth_controller.go
│   │   ├── booking_controller.go
│   │   ├── service_controller.go
│   │   ├── barber_controller.go
│   │   ├── review_controller.go
│   │   ├── cashier_controller.go
│   │   ├── payment_controller.go
│   │   └── ticket_controller.go
│   ├── service/          # Business logic (MVC Service/Use Case)
│   │   ├── user_service.go
│   │   ├── booking_service.go
│   │   ├── service_service.go
│   │   ├── barber_service.go
│   │   ├── review_service.go
│   │   ├── cashier_service.go
│   │   ├── midtrans_service.go
│   │   └── worker_service.go
│   ├── repository/       # Data access layer (MVC Model)
│   │   ├── database.go
│   │   └── interface.go
│   ├── dto/              # Data Transfer Objects
│   │   ├── requests.go
│   │   └── responses.go
│   ├── model/            # Domain models
│   │   └── entities.go
│   ├── middleware/       # HTTP middleware
│   │   ├── cors.go
│   │   ├── session.go
│   │   └── utils.go
│   └── utils/            # Utility functions
│       └── utils.go
├── config/               # External config (legacy)
├── docs/                 # Swagger documentation
├── main.go               # Application entry point
├── go.mod
├── go.sum
└── .env
```

## Layers Architecture

### 1. Controller Layer
- Menerima HTTP requests
- Validasi input dasar (JSON parsing, required fields)
- Memanggil Service layer
- Mengembalikan HTTP responses

### 2. Service Layer (Business Logic)
- Mengandung business logic
- Validasi business rules
- Koordinasi antar repository
- Transaksi business

### 3. Repository Layer (Data Access)
- Mengakses database
- Query execution
- Transaction management

### 4. DTO Layer
- Request/Response objects
- Data validation structures

### 5. Model Layer
- Database entity structures
- Domain models

## Dependencies

```bash
go mod tidy
```

Dependencies yang digunakan:
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/joho/godotenv` - Load .env files
- `golang.org/x/crypto/bcrypt` - Password hashing

## Cara Menjalankan

1. Set environment variables di `.env`:
```
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
SERVER_PORT=3000
FRONTEND_ORIGIN=http://localhost:5173
MIDTRANS_SERVER_KEY=your_key
MIDTRANS_ENVIRONMENT=sandbox
SECRET_KEY=your_secret
```

2. Build aplikasi:
```bash
go build -o barber-backend.exe .
```

3. Run:
```bash
./barber-backend.exe
```

Server akan berjalan di `http://localhost:3000`

## API Endpoints

### Authentication
- `POST /api/register` - Register new user
- `POST /api/login` - Login user
- `GET /api/profile` - Get user profile (protected)
- `POST /api/logout` - Logout (protected)

### Services
- `GET /api/services` - List all services

### Barbers
- `GET /api/barbers` - List all barbers

### Bookings
- `GET /api/bookings` - List user bookings (protected)
- `POST /api/bookings` - Create new booking (protected)
- `GET /api/bookings/{id}` - Get booking detail (protected)
- `POST /api/bookings/{id}/payment` - Create payment (protected)
- `PATCH /api/bookings/{id}/cancel` - Cancel booking (protected)
- `PATCH /api/bookings/{id}/reschedule` - Reschedule booking (protected)
- `GET /api/availability` - Check availability

### Reviews
- `GET /api/reviews` - List reviews
- `POST /api/reviews` - Create review (protected)

### Cashier
- `GET /api/cashier/queue` - Get queue (protected - cashier only)
- `GET /api/cashier/summary` - Get summary (protected - cashier only)
- `POST /api/cashier/walk-in` - Create walk-in booking (protected - cashier only)
- `PATCH /api/cashier/bookings/{id}/finish` - Finish service (protected - cashier only)
- `PATCH /api/cashier/bookings/{id}/cancel` - Cancel service (protected - cashier only)

### Tickets
- `GET /api/my-tickets` - Get user tickets (protected)

### Payments
- `POST /api/payments/midtrans/notification` - Midtrans webhook

### System
- `GET /api/status` - API status
- `GET /` - API status

## Keuntungan Modular Architecture

1. **Separation of Concerns**: Setiap layer punya tanggung jawab yang jelas
2. **Testability**: Mudah untuk testing unit (mock repository)
3. **Maintainability**: Kode lebih mudah dipelihara dan dicari
4. **Reusability**: Business logic bisa dipakai ulang di berbagai controller
5. **Scalability**: Mudah untuk menambah fitur baru tanpa mempengaruhi kode yang ada
6. **Clear Dependencies**: Dependency injection antar layer

## Migration dari Monolithic

Jika migrasi dari struktur monolitik:
1. Backup semua file `.go` lama
2. Test setiap endpoint satu per satu
3. Update swagger documentation jika diperlukan
4. Jalankan database migration jika diperlukan
5. Update deployment configuration

## Development

### Menambah Fitur Baru

1. Buat DTO baru di `internal/dto/requests.go`
2. Buat service method di `internal/service/`
3. Buat repository method di `internal/repository/database.go`
4. Buat controller handler di `internal/controller/`
5. Tambahkan route di `main.go`

### Testing

Test service layer:
```bash
go test ./internal/service/...
```

Test repository layer:
```bash
go test ./internal/repository/...
```

## Troubleshooting

### Port sudah digunakan
```bash
# Ganti SERVER_PORT di .env
```

### Database connection failed
```bash
# Pastikan DATABASE_URL benar dan PostgreSQL running
```

### Build error
```bash
# Jalankan go mod tidy
go mod tidy
```

## License

MIT License
