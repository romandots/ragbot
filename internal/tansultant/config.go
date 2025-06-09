package tansultant

import "ragbot/internal/util"

type tc struct {
	token           string
	addressEndpoint string
	pricesEndpoint  string
}

var tansConfig *tc

func loadConfig() {
	tansConfig = &tc{
		token:           util.GetEnvString("TANSULTANT_API_ACCESS_TOKEN", ""),
		addressEndpoint: util.GetEnvString("TANSULTANT_API_ADDRESS_ENDPOINT", ""),
		pricesEndpoint:  util.GetEnvString("TANSULTANT_API_PRICES_ENDPOINT", ""),
	}
}
