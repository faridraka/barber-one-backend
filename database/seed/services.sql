-- Data awal layanan Barber One.
-- Aman dijalankan berulang kali: layanan bernama sama tidak digandakan.
INSERT INTO services (name, description, price, duration_minutes)
SELECT seed.name, seed.description, seed.price, seed.duration_minutes
FROM (VALUES
    ('Haircut & Wash', 'Layanan potong rambut lengkap dengan teknik precision cutting disesuaikan bentuk wajah dan preferensi gaya Anda.', 55000::NUMERIC(12, 2), 60),
    ('Shave & Trim', 'Cukuran rapi dan presisi menggunakan alat terbaik dan teknik barbershop klasik.', 35000::NUMERIC(12, 2), 40),
    ('Hair Darkening', 'Perawatan warna rambut untuk mengembalikan pigmen alami agar rambut tampak lebih gelap dan berkilau.', 250000::NUMERIC(12, 2), 60),
    ('Hair Caviar', 'Perawatan intensif berbasis protein dan keratin untuk rambut yang lembut dan berkilau.', 150000::NUMERIC(12, 2), 60),
    ('Perm', 'Teknik pengeritingan permanen untuk menciptakan volume dan tekstur gelombang natural.', 445000::NUMERIC(12, 2), 120),
    ('Down Perm', 'Teknik pelurusan semi-permanen untuk hasil rambut yang lebih rapi dan terkontrol.', 225000::NUMERIC(12, 2), 80)
) AS seed(name, description, price, duration_minutes)
WHERE NOT EXISTS (
    SELECT 1 FROM services existing WHERE lower(existing.name) = lower(seed.name)
);
