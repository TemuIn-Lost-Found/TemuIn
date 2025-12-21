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

### 2. Konfigurasi Environment Variables

**Buat file `.env`** di root project:

```bash
# Copy dari template
cp .env.example .env
```

**Edit `.env`** dengan kredensial Google OAuth Anda:

```env
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

> **Catatan**: File `.env` sudah di-gitignore untuk keamanan. Jangan commit file ini!


### 3. Install Dependencies
Jalankan perintah ini di terminal (di dalam folder project) untuk mengunduh semua library yang dibutuhkan:

```bash
go mod tidy
```

### 4. Migrasi & Seeding Data (Wajib)
Project ini memiliki script khusus untuk membuat tabel dan mengisi data awal (dummy data). **Langkah ini wajib dijalankan pertama kali.**

Jalankan perintah:

```bash
go run cmd/reset_db/main.go
```

Jika berhasil, akan muncul pesan "Database reset complete and seeded!". Ini akan membuat user default:
- **Admin**: User `admin` / Pass `admin`
- **Warga**: User `warga_lokal` / Pass `password`

### 5. Menjalankan Aplikasi
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
=======
# TemuIn

**File Rules**
- Segala sesuatu yang berhubungan dengan gambar simpan di dalam folder `public/assets/images/`
- Folder `layouts` berfungsi untuk menyimpan file kerangka (layout utama) yang digunakan sebagai template dasar.
- Folder `review/app` berfungsi untuk menyimpan halaman inti aplikasi (fitur utama yang hanya bisa diakses setelah login atau otentikasi).
- Folder `review/components` berfungsi untuk menyimpan komponen UI pendukung yang dapat digunakan ulang di berbagai halaman.
- Folder `review/pages` berfungsi untuk menyimpan halaman umum atau public-facing yang bisa diakses tanpa login.

**Catatan Penamaan Branch**
- Silahkan membuat `branch` feat / fix sebelum mengerjakan fitur
- feat/... untuk push sebuah ftur
- fix/... untuk fix bug
- contoh : `feat/nama-fitur` = `feat/login`

# Warning
- Jangan pernah menghapus folder atau file apapun yg sudah ada atau bawaan dari laravel
- Jika ingin melakukan `git push` dan pull request pada project ini silahkan lakukan `git pull` ke branch `main` terlebih dahulu di lokal komputer
- Silahkan git push ke branch anda sendiri jangan langsung ke branch `main`
- Jika terjadi konflik silahkan perbaiki terlebih dahulu sebelum `push` ke branch anda
- Jika sudah selesai semua silahkan berikan `commit` yang jelas dan `pull request` ke branch `main`
