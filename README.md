# TemuIn - Lost & Found Application

Aplikasi Lost & Found berbasis web yang dibangun menggunakan **Go (Golang)**, **MySQL**, dan **Pongo2 Template Engine**.

## 📋 Fitur Utama

- 🔐 Autentikasi (Login/Register) dengan validasi form
- 🔑 Google OAuth 2.0 Sign-In
- 📝 Lapor barang hilang dengan upload gambar
- 🔍 Pencarian dan filter barang
- 💬 Sistem komentar pada item
- 🏆 Sistem bounty dengan koin
- ⭐ Highlight item (24 jam)
- 👁️ Password visibility toggle
- ✅ Form validation dengan error messages

## 🔧 Prasyarat (Prerequisites)

Pastikan sudah menginstal:

1. **Go (Golang)**: Versi 1.18 atau lebih baru
   - Download: https://go.dev/dl/
   - Verifikasi: `go version`

2. **MySQL**: Database server versi 5.7+ atau MariaDB
   - Windows: XAMPP, Laragon, atau standalone MySQL
   - Verifikasi: `mysql --version`

3. **Git**: Untuk clone repository
   - Download: https://git-scm.com/
   - Verifikasi: `git --version`

## 🚀 Setup Project untuk Kontributor

### 1. Clone Repository

```bash
git clone <repository-url>
cd TemuIn
```

### 2. Checkout ke Branch yang Sesuai

```bash
# Jika ingin bekerja pada fitur login/register
git checkout feat/update-login

# Atau buat branch baru untuk fitur Anda
git checkout -b feat/nama-fitur-anda
```

### 3. Konfigurasi Database

**a. Buat Database**

Buka MySQL client (phpMyAdmin, MySQL Workbench, atau command line):

```sql
CREATE DATABASE temuin_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**b. Sesuaikan Koneksi Database** (jika perlu)

Edit file `config/database.go` baris 15:

```go
dsn := "root:@tcp(127.0.0.1:3306)/temuin_db?charset=utf8mb4&parseTime=True&loc=Local"
```

Sesuaikan `root:` dengan kredensial MySQL Anda (`username:password`).

### 4. Konfigurasi Environment Variables

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

**Cara mendapatkan Google OAuth credentials:**
1. Buka [Google Cloud Console](https://console.cloud.google.com/)
2. Buat project baru atau pilih existing project
3. Enable Google+ API
4. Buat OAuth 2.0 credentials
5. Set redirect URI: `http://localhost:8080/auth/google/callback`
6. Copy Client ID dan Client Secret ke `.env`

### 5. Install Dependencies

```bash
go mod tidy
```

Perintah ini akan download semua library yang dibutuhkan seperti:
- Gin (web framework)
- GORM (ORM)
- Pongo2 (template engine)
- OAuth2 libraries
- dll.

### 6. Database Migration & Seeding

**⚠️ WAJIB** - Jalankan script ini untuk setup tabel dan data awal:

```bash
go run cmd/reset_db/main.go
```

**Output yang diharapkan:**
```
Database reset complete and seeded!
```

**User default yang dibuat:**
- **Admin**: username `admin` / password `admin`
- **Warga**: username `warga_lokal` / password `password`

> **Note**: Script ini akan **DROP** semua tabel existing dan buat ulang dengan data dummy. Hati-hati saat menjalankan!

### 7. Jalankan Server

```bash
go run main.go
```

**Output yang diharapkan:**
```
Database connection established
[GIN-debug] Listening and serving HTTP on :8080
```

Akses aplikasi di browser: **http://localhost:8080**

## 🎯 Development Workflow

### Branch Naming Convention

- `feat/nama-fitur` - Untuk fitur baru (contoh: `feat/login`, `feat/comment-system`)
- `fix/nama-bug` - Untuk bug fix (contoh: `fix/login-error`, `fix/image-upload`)
- `docs/nama-doc` - Untuk dokumentasi

### Commit Guidelines

Gunakan commit message yang jelas dan deskriptif:

```bash
# Good commits
git commit -m "feat: add Google OAuth login integration"
git commit -m "fix: resolve password validation error"
git commit -m "refactor: extract validation logic to utils"

# Bad commits  
git commit -m "update"
git commit -m "fix bug"
git commit -m "changes"
```

### Pull Request Workflow

1. **Pull latest dari main**
   ```bash
   git checkout main
   git pull origin main
   ```

2. **Merge main ke branch Anda**
   ```bash
   git checkout feat/your-feature
   git merge main
   ```

3. **Resolve conflicts** jika ada

4. **Test aplikasi** untuk pastikan tidak ada yang rusak

5. **Commit dan push**
   ```bash
   git add .
   git commit -m "feat: your clear commit message"
   git push origin feat/your-feature
   ```

6. **Buat Pull Request** ke branch `main` di GitHub

## 📁 Struktur Project

```
TemuIn/
├── cmd/                    # Command-line utilities
│   └── reset_db/          # Database reset & seeding
├── config/                 # Configuration files
│   ├── database.go        # Database connection
│   └── oauth.go           # Google OAuth config
├── handlers/               # HTTP request handlers
│   ├── auth.go            # Authentication handlers
│   ├── browse.go          # Category/subcategory browsing
│   ├── home.go            # Home page handler
│   └── items.go           # Item CRUD handlers
├── middleware/             # Gin middleware
│   └── auth.go            # Auth middleware
├── models/                 # Database models (GORM)
│   └── models.go          # User, LostItem, Category, etc.
├── routes/                 # Route definitions
│   └── routes.go          # All app routes
├── static/                 # Static assets (CSS, JS, images)
│   ├── css/
│   ├── images/
│   └── js/
├── templates/              # Pongo2 HTML templates
│   ├── base.html          # Base layout
│   ├── core/              # Core pages (login, register, home)
│   └── partials/          # Reusable components (sidebar, navbar)
├── utils/                  # Utility functions
│   ├── context.go         # Global template context
│   ├── filters.go         # Pongo2 custom filters
│   └── validation.go      # Form validation helpers
├── .env                    # Environment variables (gitignored)
├── .env.example           # Environment template
├── .gitignore             # Git ignore rules
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
├── main.go                # Application entry point
└── README.md              # This file
```

## 🔐 Security Notes

- ✅ Passwords di-hash menggunakan bcrypt
- ✅ Google OAuth credentials disimpan di `.env` (tidak di-commit)
- ✅ Session management dengan secure cookies
- ✅ Form validation di backend
- ⚠️ **Development only**: Session secret masih hardcoded di `main.go` (line 26)
  - Untuk production: gunakan environment variable

## 🐛 Troubleshooting

### Error: "Database connection failed"

**Solusi:**
1. Pastikan MySQL service running
2. Check kredensial di `config/database.go`
3. Pastikan database `temuin_db` sudah dibuat

### Error: "Template not found"

**Solusi:**
- Pastikan running `go run main.go` dari **root folder project**
- Path template harus relative dari root

### Error: "OAuth error" atau "Invalid client"

**Solusi:**
1. Pastikan `.env` file ada dan terisi
2. Verify Google OAuth credentials benar
3. Check redirect URI di Google Console sama dengan `GOOGLE_REDIRECT_URL` di `.env`
4. Restart server setelah update `.env`

### Error: "Port 8080 already in use"

**Solusi:**
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <process-id> /F

# Linux/Mac
lsof -i :8080
kill -9 <process-id>
```

### Migration Error: "Table already exists"

**Solusi:**
- Drop database manual dan buat ulang:
  ```sql
  DROP DATABASE temuin_db;
  CREATE DATABASE temuin_db;
  ```
- Jalankan ulang: `go run cmd/reset_db/main.go`

## 📚 Resources & Documentation

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM](https://gorm.io/docs/)
- [Pongo2 Template](https://github.com/flosch/pongo2)
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)

## 👥 Contributors

Contributions are welcome! Please follow the development workflow above.

## ⚠️ Important Rules

1. ❌ **Jangan hapus** file atau folder apapun tanpa diskusi dengan tim
2. ❌ **Jangan push** langsung ke branch `main`
3. ✅ **Selalu pull** dari `main` sebelum push ke branch Anda
4. ✅ **Test** perubahan Anda sebelum commit
5. ✅ **Write clear** commit messages
6. ✅ **Resolve conflicts** sebelum push

## 📝 License

This project is for educational purposes.
