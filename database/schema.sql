-- Schema Barber One: Tabel users, services, bookings, transactions, reviews

-- 1. Tabel users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- 'user', 'kasir', 'admin'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabel services (Layanan Potong Rambut & Grooming)
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price NUMERIC(12, 2) NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tabel barber
CREATE TABLE IF NOT EXISTS barber (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    rating NUMERIC(2, 1) NOT NULL DEFAULT 0.0 CHECK (rating >= 0 AND rating <= 5),
    available BOOLEAN NOT NULL DEFAULT TRUE
);

-- 4. Tabel bookings (Reservasi & Antrean)
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    ticket_code VARCHAR(20) UNIQUE NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    barber_id INT NOT NULL REFERENCES barber(id) ON DELETE RESTRICT,
    chair_number INT NOT NULL DEFAULT 1,
    booking_date DATE NOT NULL DEFAULT CURRENT_DATE,
    booking_time VARCHAR(10) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'Menunggu', -- 'Menunggu', 'Sedang Dilayani', 'Dikonfirmasi', 'Selesai', 'Batal'
    notes TEXT,
    started_at TIMESTAMPTZ, -- waktu booking masuk kursi (Sedang Dilayani)
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Tabel transactions (Transaksi Pembayaran Kasir / POS)
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    cashier_id INT REFERENCES users(id) ON DELETE SET NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL, -- 'Tunai', 'QRIS', 'Debit', 'Transfer'
    payment_status VARCHAR(30) NOT NULL DEFAULT 'Lunas',
    gateway_order_id VARCHAR(100) UNIQUE,
    gateway_token TEXT,
    gateway_redirect_url TEXT,
    expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indeks performa
CREATE INDEX IF NOT EXISTS idx_bookings_date_status ON bookings(booking_date, status);
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
-- Barber hanya dapat memiliki satu booking aktif pada tanggal dan jam yang sama.
-- Booking berstatus 'Batal' tidak ikut membatasi pemesanan ulang.
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_barber_date_time_active
    ON bookings(barber_id, booking_date, booking_time)
    WHERE status <> 'Batal';
CREATE INDEX IF NOT EXISTS idx_transactions_booking_id ON transactions(booking_id);
CREATE INDEX IF NOT EXISTS idx_transactions_paid_at ON transactions(paid_at);

-- 6. Tabel reviews (Ulasan pelanggan untuk booking yang selesai)
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (booking_id)
);
