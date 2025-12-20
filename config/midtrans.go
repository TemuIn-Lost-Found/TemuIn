package config

import (
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var (
	SnapClient     snap.Client
	MidtransEnv    midtrans.EnvironmentType
	MidtransServer string
	MidtransClient string
)

// InitMidtrans initializes the Midtrans Snap client
func InitMidtrans() {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	clientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	env := os.Getenv("MIDTRANS_ENVIRONMENT")

	// Store keys for later use
	MidtransServer = serverKey
	MidtransClient = clientKey

	// Set environment
	if env == "production" {
		MidtransEnv = midtrans.Production
	} else {
		MidtransEnv = midtrans.Sandbox
	}

	// Initialize Snap client
	SnapClient.New(serverKey, MidtransEnv)
}
