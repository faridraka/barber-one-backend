-- Waktu booking benar-benar masuk kursi. Dipakai untuk aturan tombol Selesai (30 menit)
-- dan Batal (10 menit) di dashboard kasir.
ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ;
