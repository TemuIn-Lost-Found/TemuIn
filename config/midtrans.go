package config

import (
	"log"
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

func InitMidtrans() {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	clientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	env := os.Getenv("MIDTRANS_ENV")

	if serverKey == "" || clientKey == "" {
		log.Println("Warning: MIDTRANS keys not set. Midtrans integration will not work properly.")
	}

	MidtransServer = serverKey
	MidtransClient = clientKey

	if env == "production" {
		MidtransEnv = midtrans.Production
	} else {
		MidtransEnv = midtrans.Sandbox
	}

	SnapClient.New(serverKey, MidtransEnv)
}
