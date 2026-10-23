# Menyiapkan database

Jalankan skema terlebih dahulu, lalu pilih seeder yang dibutuhkan dari folder
`backend`:

```bash
cd backend
psql "$DATABASE_URL" -f database/schema.sql
psql "$DATABASE_URL" -f database/seed/services.sql
psql "$DATABASE_URL" -f database/seed/barber.sql
psql "$DATABASE_URL" -f database/seed/users.sql
```

Atau jalankan semua seeder sekaligus:

```bash
psql "$DATABASE_URL" -f database/seed/run.sql
```

Pastikan `DATABASE_URL` sudah tersedia di environment shell. Jika nilainya hanya
ada di `backend/.env`, muat variabelnya ke shell terlebih dahulu sebelum
menjalankan `psql`.

Seeder layanan, barber, dan pengguna dapat dijalankan terpisah atau berulang kali.
Data yang sudah ada tidak akan digandakan atau ditimpa. Akun contoh di
`seed/users.sql` memakai kata sandi `BarberOne123!` dan hanya untuk pengembangan
lokal.
