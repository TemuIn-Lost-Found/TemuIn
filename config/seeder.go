package config

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"temuin/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) {
	log.Println("🌱 Starting database seeding...")
	seedCategories(db)
	seedUsers(db)
	seedLostItems(db)
	seedComments(db)
	CleanupVisitorStats(db)
	log.Println("✅ Database seeding completed")
}

func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count > 0 {
		return // Already seeded
	}

	categories := []models.Category{
		{Name: "Elektronik", Icon: "smartphone", SubCategories: []models.SubCategory{
			{Name: "Handphone"},
			{Name: "Laptop"},
			{Name: "Tablet"},
			{Name: "Kamera"},
			{Name: "Smartwatch"},
			{Name: "Headset/TWS"},
			{Name: "Powerbank"},
			{Name: "Lainnya"},
		}},
		{Name: "Dokumen", Icon: "description", SubCategories: []models.SubCategory{
			{Name: "KTP/Kartu Identitas"},
			{Name: "SIM"},
			{Name: "STNK/BPKB"},
			{Name: "Paspor"},
			{Name: "Kartu ATM/Kredit"},
			{Name: "Buku Tabungan"},
			{Name: "Ijazah/Sertifikat"},
			{Name: "Lainnya"},
		}},
		{Name: "Dompet & Tas", Icon: "account_balance_wallet", SubCategories: []models.SubCategory{
			{Name: "Dompet Pria"},
			{Name: "Dompet Wanita"},
			{Name: "Tas Ransel"},
			{Name: "Tas Selempang"},
			{Name: "Koper"},
			{Name: "Lainnya"},
		}},
		{Name: "Kunci", Icon: "vpn_key", SubCategories: []models.SubCategory{
			{Name: "Kunci Kendaraan"},
			{Name: "Kunci Rumah/Kos"},
			{Name: "Kunci Loker"},
			{Name: "Lainnya"},
		}},
		{Name: "Perhiasan", Icon: "diamond", SubCategories: []models.SubCategory{
			{Name: "Cincin"},
			{Name: "Kalung"},
			{Name: "Gelang"},
			{Name: "Jam Tangan (Analog/Digital)"},
			{Name: "Lainnya"},
		}},
		{Name: "Lainnya", Icon: "category", SubCategories: []models.SubCategory{
			{Name: "Pakaian/Aksesoris"},
			{Name: "Buku/Alat Tulis"},
			{Name: "Mainan"},
			{Name: "Alat Olahraga"},
			{Name: "Uncategorized"},
		}},
	}

	for _, cat := range categories {
		if err := db.Create(&cat).Error; err != nil {
			log.Printf("❌ Failed to seed category %s: %v", cat.Name, err)
		}
	}
	log.Println("✅ Categories seeded")
}

func seedUsers(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return // Users already exist
	}

	defaultPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	wargaPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	users := []models.User{
		{
			Username:    "admin",
			Email:       "admin@temuin.com",
			Password:    string(adminPassword),
			FirstName:   "Super",
			LastName:    "Admin",
			IsSuperuser: true,
			IsStaff:     true,
			IsActive:    true,
			CoinBalance: 1000,
		},
		{
			Username:    "warga_lokal",
			Email:       "warga@temuin.com",
			Password:    string(wargaPassword),
			FirstName:   "Warga",
			LastName:    "Lokal",
			IsActive:    true,
			CoinBalance: 500,
		},
		{
			Username:    "budi_santoso",
			Email:       "budi@student.univ.ac.id",
			Password:    string(defaultPassword),
			FirstName:   "Budi",
			LastName:    "Santoso",
			IsActive:    true,
			CoinBalance: 450,
		},
		{
			Username:    "siti_aminah",
			Email:       "siti@student.univ.ac.id",
			Password:    string(defaultPassword),
			FirstName:   "Siti",
			LastName:    "Aminah",
			IsActive:    true,
			CoinBalance: 600,
		},
		{
			Username:    "raka_pratama",
			Email:       "raka@student.univ.ac.id",
			Password:    string(defaultPassword),
			FirstName:   "Raka",
			LastName:    "Pratama",
			IsActive:    true,
			CoinBalance: 350,
		},
		{
			Username:    "dewi_lestari",
			Email:       "dewi@student.univ.ac.id",
			Password:    string(defaultPassword),
			FirstName:   "Dewi",
			LastName:    "Lestari",
			IsActive:    true,
			CoinBalance: 700,
		},
	}

	for _, u := range users {
		if err := db.Create(&u).Error; err != nil {
			log.Printf("❌ Failed to create user %s: %v", u.Username, err)
		}
	}
	log.Println("✅ Users seeded (admin/admin123, warga_lokal/password, dll)")
}

type seedItemData struct {
	Title         string
	Description   string
	Location      string
	BountyCoins   int
	CategoryName  string
	SubCategory   string
	ImageURL      string
	Username      string
	Status        string
	IsHighlighted bool
}

func seedLostItems(db *gorm.DB) {
	var count int64
	db.Model(&models.LostItem{}).Count(&count)
	if count > 0 {
		return // Items already seeded
	}

	log.Println("📥 Seeding realistic lost items with photos from internet...")

	items := []seedItemData{
		{
			Title:         "Dompet Kulit Cokelat Braun Buffel",
			Description:   "Telah hilang dompet kulit warna cokelat tua merek Braun Buffel. Hilang sekitar pukul 11 siang saat memarkirkan motor di Parkiran Gedung Utama. Isi dompet terdapat KTP, SIM C, kartu KTM, dan beberapa kartu ATM atas nama Budi Santoso. Bagi yang menemukan mohon sangat untuk dihubungi atau dikembalikan ke pos satpam, imbalan tersedia. Terima kasih.",
			Location:      "Parkiran Motor Gedung Utama",
			BountyCoins:   200,
			CategoryName:  "Dompet & Tas",
			SubCategory:   "Dompet Pria",
			ImageURL:      "https://images.unsplash.com/photo-1627123424574-724758594e93?w=600&auto=format&fit=crop&q=80",
			Username:      "budi_santoso",
			Status:        "LOST",
			IsHighlighted: true,
		},
		{
			Title:         "iPhone 13 128GB Hitam (Casing Bening MagSafe)",
			Description:   "Hilang iPhone 13 warna Midnight Black 128GB. Terakhir ditaruh di meja saat makan siang di kantin blok selatan sekitar pukul 12.30. Ciri-ciri menggunakan casing transparan bermagnet (MagSafe) dan ada stiker kucing kecil di belakang casing. Wallpaper layar kunci foto pemandangan gunung. Kondisi telepon masih berdering tapi tidak diangkat. Mohon bantuannya bagi yang menemukan!",
			Location:      "Kantin Blok Selatan, Meja dekat kasir",
			BountyCoins:   500,
			CategoryName:  "Elektronik",
			SubCategory:   "Handphone",
			ImageURL:      "https://images.unsplash.com/photo-1591337676887-a217a6970a8a?w=600&auto=format&fit=crop&q=80",
			Username:      "siti_aminah",
			Status:        "LOST",
			IsHighlighted: true,
		},
		{
			Title:         "MacBook Air M2 Silver + Sleeve Case Abu-abu",
			Description:   "Ketinggalan laptop MacBook Air M2 warna Silver yang disimpan di dalam sleeve case bahan kanvas abu-abu. Di dalam tas laptop juga ada charger MagSafe 3 dan mouse wireless Logitech M170. Terakhir dipakai nugas di perpustakaan lantai 2 dekat jendela barat jam 15.00 WIB. Sangat penting karena berisi data skripsi dan pekerjaan!",
			Location:      "Ruang Baca Lantai 2, Perpustakaan Pusat",
			BountyCoins:   500,
			CategoryName:  "Elektronik",
			SubCategory:   "Laptop",
			ImageURL:      "https://images.unsplash.com/photo-1611186871348-b1ce696e52c9?w=600&auto=format&fit=crop&q=80",
			Username:      "raka_pratama",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Kunci Motor Honda Scoopy & Remote Keyless (Gantungan Boneka)",
			Description:   "Hilang satu set kunci motor Honda Scoopy beserta remote smart key (keyless). Ada gantungan kunci bentuk boneka warna biru dan tali hitam merek Honda. Perkiraan jatuh saat jalan kaki usai sholat Zuhur dari Masjid kembali ke fakultas. Motor tidak bisa diparkir pulang tanpa remote ini, mohon infonya!",
			Location:      "Jalur dari Gedung Fakultas Teknik ke Masjid Baiturrahim",
			BountyCoins:   150,
			CategoryName:  "Kunci",
			SubCategory:   "Kunci Kendaraan",
			ImageURL:      "https://images.unsplash.com/photo-1582139329536-e7284fece509?w=600&auto=format&fit=crop&q=80",
			Username:      "dewi_lestari",
			Status:        "LOST",
			IsHighlighted: true,
		},
		{
			Title:         "KTP dan SIM A/C atas nama Raka Pratama",
			Description:   "Hilang dompet kecil khusus kartu yang berisi KTP, SIM A, SIM C, dan Kartu BPJS atas nama Raka Pratama. Dompet kartunya terbuat dari bahan plastik hitam transparan. Hilang waktu lari jogging pagi di sekitar lapangan basket Sport Center. Buat yang menemukan tolong dikabari ya.",
			Location:      "Lapangan Basket Sport Center",
			BountyCoins:   100,
			CategoryName:  "Dokumen",
			SubCategory:   "KTP/Kartu Identitas",
			ImageURL:      "https://images.unsplash.com/photo-1589578527966-fdac0f44566c?w=600&auto=format&fit=crop&q=80",
			Username:      "raka_pratama",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Tas Ransel Kanvas Hitam Merk Exsport",
			Description:   "Tertinggal tas ransel kanvas warna hitam polos merk Exsport setelah acara seminar kewirausahaan pagi tadi. Ciri-ciri ada gantungan kunci pin akrilik logo universitas di seleting depan. Isi dalam tas ada catatan kuliah, kalkulator scientific, dan jaket hoodie biru. Kalau ada yang mengamankan mohon kabari ya.",
			Location:      "Ruang Seminar R-304 Gedung Administrasi",
			BountyCoins:   250,
			CategoryName:  "Dompet & Tas",
			SubCategory:   "Tas Ransel",
			ImageURL:      "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=600&auto=format&fit=crop&q=80",
			Username:      "budi_santoso",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Headset Bluetooth Sony WH-1000XM4 Warna Hitam",
			Description:   "Hilang headphone wireless Sony WH-1000XM4 warna hitam beserta hardcase bawaannya. Tertinggal di kursi tunggu halte saat menunggu jemputan online sekitar jam 17.00 sore kemarin. Kondisi masih sangat mulus dan ada stiker nama 'Siti A.' di bagian dalam case.",
			Location:      "Halte Bus Barat dekat gerbang keluar",
			BountyCoins:   350,
			CategoryName:  "Elektronik",
			SubCategory:   "Headset/TWS",
			ImageURL:      "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=600&auto=format&fit=crop&q=80",
			Username:      "siti_aminah",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Jam Tangan Seiko 5 Automatic Strap Kanvas Hijau",
			Description:   "Kelupaan mengambil jam tangan analog Seiko 5 Automatic berstrap kain NATO warna hijau army (green canvas). Dilepas saat wudhu di keran wastafel kamar mandi gedung B lantai 1. Jam ini punya nilai kenangan tinggi pemberian orang tua, mohon dengan sangat bagi yang mengamankan untuk dikembalikan.",
			Location:      "Kamar Mandi Lantai 1 Gedung B",
			BountyCoins:   300,
			CategoryName:  "Perhiasan",
			SubCategory:   "Jam Tangan (Analog/Digital)",
			ImageURL:      "https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=600&auto=format&fit=crop&q=80",
			Username:      "warga_lokal",
			Status:        "LOST",
			IsHighlighted: true,
		},
		{
			Title:         "Tablet iPad Air 5 Blue + Apple Pencil 2",
			Description:   "Hilang 1 unit iPad Air generasi ke-5 warna biru (Blue) beserta Apple Pencil 2 yang menempel pada casing lipat warna Navy. Ketinggalan di sofa lobby utama saat menelepon orang tua kemarin sore. Layar terpasang tempered glass matte. Ada imbalan pantas bagi penemu jujur.",
			Location:      "Lobby Utama Gedung Rektorat di sofa tunggu",
			BountyCoins:   500,
			CategoryName:  "Elektronik",
			SubCategory:   "Tablet",
			ImageURL:      "https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?w=600&auto=format&fit=crop&q=80",
			Username:      "dewi_lestari",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Kunci Rumah dan Kos (3 Anak Kunci + Gantungan Akrilik)",
			Description:   "Hilang sebentang kunci berisi 3 anak kunci silinder dan 1 kunci gembok kecil merk Kenko. Diciri-irikan dengan gantungan akrilik bertuliskan 'KOS NO 12B'. Kemungkinan jatuh pas istirahat jam 12 di kantin outdoor. Belum bisa masuk kamar kos dari siang.",
			Location:      "Kantin Fakultas Hukum / Area Tempat Duduk Outdoor",
			BountyCoins:   100,
			CategoryName:  "Kunci",
			SubCategory:   "Kunci Rumah/Kos",
			ImageURL:      "https://images.unsplash.com/photo-1558036117-15d82a90b9b1?w=600&auto=format&fit=crop&q=80",
			Username:      "budi_santoso",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Kacamata Minus Frame Besi Gold (Merk Ray-Ban)",
			Description:   "Ketinggalan kacamata minus (-3.00 dan silinder) frame besi tipis warna gold merk Ray-Ban di atas meja lab komputer nomor 14. Ditaruh di atas kotak kacamata warna cokelat bermerk Optik Seis. Sangat dibutuhkan untuk aktivitas belajar setiap hari.",
			Location:      "Ruang Laboratorium Komputer Lantai 3",
			BountyCoins:   150,
			CategoryName:  "Lainnya",
			SubCategory:   "Pakaian/Aksesoris",
			ImageURL:      "https://images.unsplash.com/photo-1572635196237-14b3f281503f?w=600&auto=format&fit=crop&q=80",
			Username:      "warga_lokal",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Buku Tabungan Bank BCA & Kartu ATM atas nama Dewi Lestari",
			Description:   "Jatuh amplop cokelat berisikan Buku Tabungan BCA Tahapan dan Kartu ATM pasca print mutasi dari bank kampus. Perkiraan jatuh di jalur pejalan kaki antara Gedung E menuju Gedung F. Rekening sudah diblokir, tapi buku fisiknya perlu untuk pembukaan kembali.",
			Location:      "Jalur Pejalan Kaki Antara Gedung E dan Gedung F",
			BountyCoins:   100,
			CategoryName:  "Dokumen",
			SubCategory:   "Buku Tabungan",
			ImageURL:      "https://images.unsplash.com/photo-1559526324-4b87b5e36e44?w=600&auto=format&fit=crop&q=80",
			Username:      "dewi_lestari",
			Status:        "FOUND",
			IsHighlighted: false,
		},
		{
			Title:         "Dompet Wanita Merah Maroon Merk Fossil",
			Description:   "Telah hilang dompet panjang wanita warna merah maroon merk Fossil bermotif kulit jeruk. Berisi uang tunai, kartu Flazz BCA, dan beberapa tanda pengenal atas nama Dewi Lestari. Bagi yang menemukan, uang tunainya ikhlas buat penemu, tolong kembalikan dokumen dan kartunya saja.",
			Location:      "Taman Baca Depan Fakultas Ilmu Sosial",
			BountyCoins:   200,
			CategoryName:  "Dompet & Tas",
			SubCategory:   "Dompet Wanita",
			ImageURL:      "https://images.unsplash.com/photo-1590874103328-eac38a683ce7?w=600&auto=format&fit=crop&q=80",
			Username:      "dewi_lestari",
			Status:        "FOUND",
			IsHighlighted: false,
		},
		{
			Title:         "Powerbank Anker PowerCore 20000mAh Hitam",
			Description:   "Hilang powerbank Anker warna hitam doff ukuran besar 20.000mAh beserta kabel pendek braided USB-C merk Baseus. Ketinggalan saat cas HP di kursi pojokan UKM Center semalam pas musyawarah rutin.",
			Location:      "Ruang Tunggu UKM Center",
			BountyCoins:   100,
			CategoryName:  "Elektronik",
			SubCategory:   "Powerbank",
			ImageURL:      "https://images.unsplash.com/photo-1609592424209-27d490c37bb7?w=600&auto=format&fit=crop&q=80",
			Username:      "raka_pratama",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Botol Minum Thermos Stainless Corkcicle Hitam 16oz",
			Description:   "Tertinggal tumbler minuman besi stainless merk Corkcicle warna hitam matte ukuran 16oz di bawah kursi baris D nomor 12 saat mengikuti kuliah tamu kemarin. Ada stiker logo Spotify kecil di bagian dasar botol.",
			Location:      "Ruang Auditorium Lantai Dasar (Kursi Baris D)",
			BountyCoins:   150,
			CategoryName:  "Lainnya",
			SubCategory:   "Uncategorized",
			ImageURL:      "https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=600&auto=format&fit=crop&q=80",
			Username:      "siti_aminah",
			Status:        "LOST",
			IsHighlighted: false,
		},
		{
			Title:         "Cincin Emas Polos 2 Gram (Cincin Tunangan)",
			Description:   "Telah jatuh atau terlepas cincin tunangan bahan emas putih polos tanpa mata sekitar 2 gram. Terakhir mencuci tangan di wastafel toilet wanita Student Center jam 14.00 WIB. Bagi penemu yang baik hati mohon kabari, ada imbalan sepantasnya dari saya.",
			Location:      "Toilet Wanita Lantai 2 Gedung Student Center",
			BountyCoins:   400,
			CategoryName:  "Perhiasan",
			SubCategory:   "Cincin",
			ImageURL:      "https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=600&auto=format&fit=crop&q=80",
			Username:      "dewi_lestari",
			Status:        "LOST",
			IsHighlighted: false,
		},
	}

	for _, data := range items {
		var user models.User
		if err := db.Where("username = ?", data.Username).First(&user).Error; err != nil {
			db.Where("username = ?", "admin").First(&user)
		}

		var cat models.Category
		if err := db.Where("name = ?", data.CategoryName).First(&cat).Error; err != nil {
			log.Printf("⚠️ Category %s not found for item %s", data.CategoryName, data.Title)
			continue
		}

		var subCat models.SubCategory
		if err := db.Where("name = ? AND category_id = ?", data.SubCategory, cat.ID).First(&subCat).Error; err != nil {
			db.Where("category_id = ?", cat.ID).First(&subCat)
		}

		log.Printf("🌐 Downloading image for: %s...", data.Title)
		imgBytes, contentType := fetchOrFallbackImage(data.ImageURL)

		item := models.LostItem{
			Title:         data.Title,
			Description:   data.Description,
			Location:      data.Location,
			BountyCoins:   data.BountyCoins,
			UserID:        user.ID,
			CategoryID:    cat.ID,
			SubCategoryID: &subCat.ID,
			Status:        data.Status,
			Image:         "stored",
			IsHighlighted: data.IsHighlighted,
		}

		if data.IsHighlighted {
			expiry := time.Now().Add(72 * time.Hour)
			item.HighlightExpiry = &expiry
		}

		if data.Status == "FOUND" {
			var finder models.User
			db.Where("username = ?", "warga_lokal").First(&finder)
			item.FinderID = &finder.ID
			item.FinderConfirmed = true
			item.OwnerConfirmed = true
		}

		if err := db.Create(&item).Error; err != nil {
			log.Printf("❌ Failed to seed item: %s: %v", data.Title, err)
			continue
		}

		if len(imgBytes) > 0 {
			imgRec := models.LostItemImage{
				ItemID:      item.ID,
				ImageData:   imgBytes,
				ContentType: contentType,
			}
			db.Create(&imgRec)
		}
	}
	log.Println("✅ Lost items & images seeded successfully!")
}

func seedComments(db *gorm.DB) {
	var count int64
	db.Model(&models.Comment{}).Count(&count)
	if count > 0 {
		return
	}

	var item models.LostItem
	if err := db.Where("status = ?", "LOST").First(&item).Error; err != nil {
		return
	}

	var u1, u2 models.User
	db.Where("username = ?", "warga_lokal").First(&u1)
	db.Where("username = ?", "admin").First(&u2)

	if u1.ID != 0 && u2.ID != 0 {
		comments := []models.Comment{
			{
				ItemID:  item.ID,
				UserID:  u1.ID,
				Content: "Wah saya sempat lihat di dekat pos security kemarin sore mas, coba dicek kesana.",
			},
			{
				ItemID:  item.ID,
				UserID:  u2.ID,
				Content: "Terima kasih infonya! Untuk penemu silakan laporkan lewat fitur klaim ya agar aman.",
			},
		}
		for _, c := range comments {
			db.Create(&c)
		}
		log.Println("✅ Sample comments seeded")
	}
}

func fetchOrFallbackImage(url string) ([]byte, string) {
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			contentType := resp.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}
			return data, contentType
		}
	}
	log.Printf("⚠️ Could not fetch image from URL, generating fallback graphic for %s...", url)
	return createFallbackImage(), "image/png"
}

func createFallbackImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	col := color.RGBA{R: 88, G: 101, B: 242, A: 255}
	for x := 0; x < 600; x++ {
		for y := 0; y < 400; y++ {
			img.Set(x, y, col)
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// CleanupVisitorStats removes duplicate visits (same IP, same Day), keeping only the earliest one.
func CleanupVisitorStats(db *gorm.DB) {
	log.Println("🧹 Starting Visitor Stats Cleanup...")

	query := `
		DELETE t1 FROM core_sitevisit t1
		INNER JOIN core_sitevisit t2 
		WHERE 
			t1.id > t2.id AND 
			t1.ip_address = t2.ip_address AND 
			DATE(t1.visited_at) = DATE(t2.visited_at)
	`

	if err := db.Exec(query).Error; err != nil {
		log.Printf("⚠️ Failed to clean up visitor stats: %v", err)
	} else {
		log.Println("✅ Visitor Stats cleaned up! (Duplicates removed)")
	}
}
