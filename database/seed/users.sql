-- Akun contoh untuk pengembangan lokal.
-- Kata sandi seluruh akun: BarberOne123!
-- Jangan gunakan akun ini di produksi.
INSERT INTO users (name, phone, email, password_hash, role)
SELECT seed.name, seed.phone, seed.email, seed.password_hash, seed.role
FROM (VALUES
    ('Kasir Barber One', '081234567891', 'kasir@barberone.local', '$2a$10$NjNDo/L52gM1xd8Pt2rbPe0wLboWAbqJ6jn8jHbyJ16qcLAV3q8uO', 'kasir'),
    ('Pelanggan Contoh', '081234567890', 'user@barberone.local', '$2a$10$NjNDo/L52gM1xd8Pt2rbPe0wLboWAbqJ6jn8jHbyJ16qcLAV3q8uO', 'user')
) AS seed(name, phone, email, password_hash, role)
WHERE NOT EXISTS (
    SELECT 1
    FROM users existing
    WHERE existing.phone = seed.phone
       OR lower(existing.email) = lower(seed.email)
);
