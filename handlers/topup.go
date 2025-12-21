package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
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
		log.Printf("error creating topup transaction: %v", err)
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
		log.Printf("snap create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"snap_token": snapResp.Token,
		"order_id":   orderID,
	})
}

// MidtransNotification handles payment notification callbacks from Midtrans
// (This endpoint must be configured in Midtrans dashboard as "notification URL")
func MidtransNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}

	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification"})
		return
	}

	// Extract values safely and stringify them for signature calculation
	rawOrderID := notificationPayload["order_id"]
	rawStatusCode := notificationPayload["status_code"]
	rawGross := notificationPayload["gross_amount"]
	signatureKey, _ := notificationPayload["signature_key"].(string)

	orderID := fmt.Sprintf("%v", rawOrderID)
	statusCode := fmt.Sprintf("%v", rawStatusCode)
	grossAmount := fmt.Sprintf("%v", rawGross)

	// Calculate expected signature
	serverKey := config.MidtransServer
	expectedSignature := fmt.Sprintf("%s%s%s%s", orderID, statusCode, grossAmount, serverKey)
	hash := sha512.Sum512([]byte(expectedSignature))
	calculatedSignature := hex.EncodeToString(hash[:])

	if signatureKey != calculatedSignature {
		log.Printf("invalid signature: got=%s want=%s", signatureKey, calculatedSignature)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Get transaction from database
	var transaction models.TopUpTransaction
	if err := config.DB.Where("order_id = ?", orderID).First(&transaction).Error; err != nil {
		log.Printf("transaction not found for order_id=%s", orderID)
		// still return 200 to avoid Midtrans repeated retries? Better return 404 so you see missing mapping.
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	prevStatus := transaction.Status

	// Parse transaction status from Midtrans
	transactionStatus := fmt.Sprintf("%v", notificationPayload["transaction_status"])
	fraudStatus := fmt.Sprintf("%v", notificationPayload["fraud_status"])
	paymentType := fmt.Sprintf("%v", notificationPayload["payment_type"])
	transactionTimeStr := fmt.Sprintf("%v", notificationPayload["transaction_time"])

	// Parse transaction time (Midtrans returns "2006-01-02 15:04:05")
	var transactionTime *time.Time
	if transactionTimeStr != "" && transactionTimeStr != "<nil>" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", transactionTimeStr); err == nil {
			transactionTime = &parsedTime
		}
	}

	// Update transaction based on status
	var newStatus string
	shouldAddCoins := false

	switch transactionStatus {
	case "capture":
		// card capture -> check fraud status
		if fraudStatus == "accept" || fraudStatus == "challenge" {
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
	case "deny", "cancel", "expire", "failure":
		newStatus = "failed"
	default:
		newStatus = "pending"
	}

	// Update transaction record
	transaction.Status = newStatus
	transaction.PaymentType = paymentType
	transaction.TransactionTime = transactionTime

	if err := config.DB.Save(&transaction).Error; err != nil {
		log.Printf("failed saving transaction update: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Idempotent coin add: only add if newly transitioned to success
	if shouldAddCoins && transaction.Status == "success" && prevStatus != "success" {
		var user models.User
		if err := config.DB.First(&user, transaction.UserID).Error; err == nil {
			// Add coins (Amount is number of coins)
			user.CoinBalance += transaction.Amount
			if err := config.DB.Save(&user).Error; err != nil {
				log.Printf("failed updating user balance: %v", err)
			} else {
				// Create coin transaction record for accounting
				coinTransaction := models.CoinTransaction{
					UserID:          user.ID,
					Amount:          transaction.Amount,
					TransactionType: "topup",
				}
				if err := config.DB.Create(&coinTransaction).Error; err != nil {
					log.Printf("failed creating coin transaction record: %v", err)
				}
			}
		} else {
			log.Printf("user not found while adding coins: id=%d", transaction.UserID)
		}
	}

	// Respond OK to Midtrans
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

// CheckTopUpStatus checks the status of a specific transaction with Midtrans
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
		log.Printf("coreapi check error: %v", err)
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

// ConfirmTopUp allows frontend to ask server to re-check Midtrans and apply coins if settled.
// This helps sandbox flows where the Snap popup returns but notification from Midtrans may arrive later.
func ConfirmTopUp(c *gin.Context) {
	type reqBody struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id required"})
		return
	}

	orderID := body.OrderID
	coreClient := coreapi.Client{}
	coreClient.New(config.MidtransServer, config.MidtransEnv)

	resp, err := coreClient.CheckTransaction(orderID)
	if err != nil {
		log.Printf("coreapi check error confirm: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check Midtrans: " + err.Error()})
		return
	}

	var tx models.TopUpTransaction
	if err := config.DB.Where("order_id = ?", orderID).First(&tx).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	prevStatus := tx.Status
	transactionStatus := resp.TransactionStatus
	fraudStatus := resp.FraudStatus
	paymentType := resp.PaymentType

	newStatus := tx.Status
	shouldAddCoins := false

	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" || fraudStatus == "challenge" {
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
	case "deny", "cancel", "expire", "failure":
		newStatus = "failed"
	default:
		newStatus = "pending"
	}

	tx.Status = newStatus
	tx.PaymentType = paymentType
	if err := config.DB.Save(&tx).Error; err != nil {
		log.Printf("failed saving tx on confirm: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Add coins idempotently
	if shouldAddCoins && tx.Status == "success" && prevStatus != "success" {
		var user models.User
		if err := config.DB.First(&user, tx.UserID).Error; err == nil {
			user.CoinBalance += tx.Amount
			if err := config.DB.Save(&user).Error; err != nil {
				log.Printf("failed updating user balance on confirm: %v", err)
			} else {
				coinTransaction := models.CoinTransaction{
					UserID:          user.ID,
					Amount:          tx.Amount,
					TransactionType: "topup",
				}
				if err := config.DB.Create(&coinTransaction).Error; err != nil {
					log.Printf("failed creating coin transaction on confirm: %v", err)
				}
			}
		} else {
			log.Printf("user not found on confirm: id=%d", tx.UserID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "ok",
		"midtrans_status":  transactionStatus,
		"midtrans_fraud":   fraudStatus,
		"local_status":     tx.Status,
		"should_add_coins": shouldAddCoins,
	})
}
