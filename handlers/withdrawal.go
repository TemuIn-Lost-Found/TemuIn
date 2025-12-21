// handlers/withdrawal.go
package handlers

import (
	"net/http"

	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/gin-gonic/gin"
)

const rateRpPerCoin = 10 // 1 coin = Rp10 (sesuaikan ke kebutuhan)

// WithdrawalRequestBody simple struct
type WithdrawalRequestBody struct {
	Coins         int    `json:"coins" binding:"required"`
	Method        string `json:"method" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	Note          string `json:"note"`
}

const RATE_RP_PER_COIN = 10

func WithdrawalPage(c *gin.Context) {
	utils.RenderTemplate(c, "templates/core/withdrawal.html", nil)
}

func RequestWithdrawal(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var body WithdrawalRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if body.Coins < 100 || body.Coins%100 != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minimum 100 coins & multiple of 100"})
		return
	}

	amountIDR := body.Coins * RATE_RP_PER_COIN

	tx := config.DB.Begin()

	var u models.User
	if err := tx.First(&u, user.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	if u.CoinBalance < body.Coins {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient coin balance"})
		return
	}

	u.CoinBalance -= body.Coins
	if err := tx.Save(&u).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct coins"})
		return
	}

	tx.Create(&models.CoinTransaction{
		UserID:          u.ID,
		Amount:          -body.Coins,
		TransactionType: "withdraw_request",
	})

	wr := models.WithdrawalRequest{
		UserID:        u.ID,
		Coins:         body.Coins,
		Amount:        amountIDR,
		Method:        body.Method,
		AccountName:   body.AccountName,
		AccountNumber: body.AccountNumber,
		Status:        "pending",
		Note:          body.Note,
	}

	if err := tx.Create(&wr).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create withdrawal"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetWithdrawalHistory user endpoint
func GetWithdrawalHistory(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var list []models.WithdrawalRequest
	config.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"withdrawals": list})
}
