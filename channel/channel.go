package channel

import (
	"RappaMaster/database"
	fisco_bcos_client "RappaMaster/fisco-bcos-client"
	"RappaMaster/transaction"
)

type RappaChannel struct {
	*database.DatabaseService // shared db
	*fisco_bcos_client.RappaFBClient
	UpchainBuffer chan transaction.Transaction
}
