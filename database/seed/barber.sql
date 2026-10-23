-- Data awal barber. Aman dijalankan berulang kali: nama yang sama tidak digandakan.
INSERT INTO barber (name, rating, available)
SELECT seed.name, seed.rating, seed.available
FROM (VALUES
    ('Andi', 4.8::NUMERIC(2, 1), TRUE),
    ('Budi', 4.6::NUMERIC(2, 1), TRUE),
    ('Rizky', 4.9::NUMERIC(2, 1), FALSE)
) AS seed(name, rating, available)
WHERE NOT EXISTS (
    SELECT 1 FROM barber existing WHERE lower(existing.name) = lower(seed.name)
);
