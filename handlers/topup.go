package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"temuin/config"
	"temuin/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

// TopUpRequest represents the request body for initiating a top-up
type TopUpRequest struct {
	Amount int `json:"amount" binding:"required"`
}

// InitiateTopUp creates a Midtrans Snap transaction and returns the token
func InitiateTopUp(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate amount (only 100, 500, 1000 allowed)
	validAmounts := map[int]int{
		100:  1000,  // 100 coins = Rp 1.000
		500:  5000,  // 500 coins = Rp 5.000
		1000: 10000, // 1000 coins = Rp 10.000
	}

	price, valid := validAmounts[req.Amount]
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin amount. Choose 100, 500, or 1000."})
		return
	}

	// Generate unique Order ID using timestamp and user ID
	orderID := fmt.Sprintf("TOPUP-%d-%d", user.ID, time.Now().Unix())

	// Create TopUpTransaction record
	transaction := models.TopUpTransaction{
		OrderID: orderID,
		UserID:  user.ID,
		Amount:  req.Amount,
		Price:   price,
		Status:  "pending",
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Create Snap transaction request
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: int64(price),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Username,
			Email: user.Email,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    fmt.Sprintf("COIN-%d", req.Amount),
				Name:  fmt.Sprintf("%d TemuIn Coins", req.Amount),
				Price: int64(price),
				Qty:   1,
			},
		},
	}

	// Get Snap token from Midtrans
	snapResp, err := config.SnapClient.CreateTransaction(snapReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snap_token": snapResp.Token,
		"order_id":   orderID,
	})
}

// MidtransNotification handles payment notification callbacks from Midtrans
func MidtransNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}

	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification"})
		return
	}

	// Verify signature
	orderID, _ := notificationPayload["order_id"].(string)
	statusCode, _ := notificationPayload["status_code"].(string)
	grossAmount, _ := notificationPayload["gross_amount"].(string)
	signatureKey, _ := notificationPayload["signature_key"].(string)

	// Calculate expected signature
	serverKey := config.MidtransServer
	expectedSignature := fmt.Sprintf("%s%s%s%s", orderID, statusCode, grossAmount, serverKey)
	hash := sha512.Sum512([]byte(expectedSignature))
	calculatedSignature := hex.EncodeToString(hash[:])

	if signatureKey != calculatedSignature {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Get transaction from database
	var transaction models.TopUpTransaction
	if err := config.DB.Where("order_id = ?", orderID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Parse transaction status from Midtrans
	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	fraudStatus, _ := notificationPayload["fraud_status"].(string)
	paymentType, _ := notificationPayload["payment_type"].(string)
	transactionTimeStr, _ := notificationPayload["transaction_time"].(string)

	// Parse transaction time
	var transactionTime *time.Time
	if transactionTimeStr != "" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", transactionTimeStr); err == nil {
			transactionTime = &parsedTime
		}
	}

	// Update transaction based on status
	var newStatus string
	shouldAddCoins := false

	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			newStatus = "success"
			shouldAddCoins = true
		} else {
			newStatus = "failed"
		}
	case "settlement":
		newStatus = "success"
		shouldAddCoins = true
	case "pending":
		newStatus = "pending"
	case "deny", "cancel", "expire":
		newStatus = "failed"
	case "failure":
		newStatus = "failed"
	default:
		newStatus = "pending"
	}

	// Update transaction record
	transaction.Status = newStatus
	transaction.PaymentType = paymentType
	transaction.TransactionTime = transactionTime

	if err := config.DB.Save(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Add coins to user balance if payment successful
	if shouldAddCoins && transaction.Status == "success" {
		var user models.User
		if err := config.DB.First(&user, transaction.UserID).Error; err == nil {
			user.CoinBalance += transaction.Amount
			config.DB.Save(&user)

			// Create coin transaction record for accounting
			coinTransaction := models.CoinTransaction{
				UserID:          user.ID,
				Amount:          transaction.Amount,
				TransactionType: "topup",
			}
			config.DB.Create(&coinTransaction)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetTopUpHistory returns the user's top-up transaction history
func GetTopUpHistory(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var transactions []models.TopUpTransaction
	config.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&transactions)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}

// CheckTopUpStatus checks the sta tus of a specific transaction with Midtrans
func CheckTopUpStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	user := c.MustGet("user").(*models.User)

	// Verify transaction belongs to user
	var transaction models.TopUpTransaction
	if err := config.DB.Where("order_id = ? AND user_id = ?", orderID, user.ID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Query Midtrans for transaction status
	coreClient := coreapi.Client{}
	coreClient.New(config.MidtransServer, config.MidtransEnv)

	transactionStatusResp, err := coreClient.CheckTransaction(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":           orderID,
		"transaction_status": transactionStatusResp.TransactionStatus,
		"payment_type":       transactionStatusResp.PaymentType,
		"gross_amount":       transactionStatusResp.GrossAmount,
		"local_status":       transaction.Status,
	})
}
