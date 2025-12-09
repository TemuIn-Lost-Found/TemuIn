# TemuIn - Lost & Found Application

Aplikasi Lost & Found yang dibangun menggunakan Go (Golang) dan MySQL.

## Prasyarat (Prerequisites)

Sebelum menjalankan aplikasi, pastikan teman Anda sudah menginstal:

1.  **Go (Golang)**: Versi 1.18 ke atas.
2.  **MySQL**: Database server.

## Cara Install & Menjalankan (Project Setup)

Berikut adalah langkah-langkah untuk teman Anda yang baru selesai melakukan clone:

### 1. Konfigurasi Database
Pastikan layanan MySQL sudah berjalan. Buat database kosong bernama `temuin_db` (atau sesuai konfigurasi di `config/database.go`).

```sql
CREATE DATABASE temuin_db;
```

*Catatan: Konfigurasi default di `config/database.go` menggunakan user `root` tanpa password. Jika konfigurasi MySQL teman Anda berbeda, minta mereka menyesuaikan file `config/database.go` baris 11.*

### 2. Install Dependencies
Jalankan perintah ini di terminal (di dalam folder project) untuk mengunduh semua library yang dibutuhkan:

```bash
go mod tidy
```

### 3. Migrasi & Seeding Data (Wajib)
Project ini memiliki script khusus untuk membuat tabel dan mengisi data awal (dummy data). **Langkah ini wajib dijalankan pertama kali.**

Jalankan perintah:

```bash
go run cmd/reset_db/main.go
```

Jika berhasil, akan muncul pesan "Database reset complete and seeded!". Ini akan membuat user default:
- **Admin**: User `admin` / Pass `admin`
- **Warga**: User `warga_lokal` / Pass `password`

### 4. Menjalankan Aplikasi
Setelah database siap, jalankan server utama:

```bash
go run main.go
```

Akses aplikasi di browser melalui: `http://localhost:8080`

## Pertanyaan Umum

**Q: Apakah perlu generate key?**
A: Untuk saat ini **TIDAK**. Secret key untuk session masih di-hardcode di `main.go`. Namun untuk production nanti, disarankan menggunakan environment variable.

**Q: Bagaimana kalau mau reset data lagi?**
A: Cukup jalankan ulang perintah `go run cmd/reset_db/main.go`. Hati-hati, semua data akan dihapus dan diganti data dummy baru.
