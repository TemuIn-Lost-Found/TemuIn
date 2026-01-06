package config

import (
	"log"
	"temuin/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) {
	seedCategories(db)
	seedAdmin(db)
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

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		return // Admin already exists
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	admin := models.User{
		Username:    "admin",
		Email:       "admin@temuin.com",
		Password:    string(hashedPassword),
		FirstName:   "Super",
		LastName:    "Admin",
		IsSuperuser: true,
		IsStaff:     true,
		IsActive:    true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to create admin user: %v", err)
	} else {
		log.Println("✅ Admin user created (username: admin, password: admin123)")
	}
}

// CleanupVisitorStats removes duplicate visits (same IP, same Day), keeping only the earliest one.
func CleanupVisitorStats(db *gorm.DB) {
	log.Println("🧹 Starting Visitor Stats Cleanup...")
	
	// SQL to delete duplicates in MySQL:
	// Delete t1 (the duplicate with higher ID)
	// From core_sitevisit t1
	// Inner Join core_sitevisit t2
	// Where they have same IP and same Date
	// And t1.id > t2.id (so we keep t2, the earlier one)
	
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
