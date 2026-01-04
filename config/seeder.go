package config

import (
	"log"
	"temuin/models"

	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) {
	seedCategories(db)
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
