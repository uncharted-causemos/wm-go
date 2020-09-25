package storage

// Model names
const (
	APSIM                 = "apsim"
	AssetWealth           = "asset_wealth_model"
	CHIRPS                = "chirps"
	CLEM                  = "clem"
	Consumption           = "consumption_model"
	Cropland              = "cropland_model"
	DSSAT                 = "dssat"
	GRange                = "g-range"
	FloodIndex            = "flood_index_model"
	FSC                   = "fsc"
	LPJML                 = "lpjml"
	LPJMLHistoric         = "lpjml_historic"
	Malnutrition          = "malnutrition_model"
	MarketPrice           = "market_price_model"
	MultiTwist            = "multi_twist"
	PIHM                  = "pihm"
	Population            = "population_model"
	WorldPopulationAfrica = "world_population_africa"
	YieldAnomaliesLPJML   = "yield_anomalies_lpjml"
)

var modelMaxPrecision = map[string]uint32{
	GRange:       10,
	Consumption:  14,
	AssetWealth:  14,
	Malnutrition: 15,
	PIHM:         16,

	LPJML:         9,
	APSIM:         10,
	LPJMLHistoric: 10,
	FloodIndex:    11,
	MarketPrice:   11,
	DSSAT:         12,
	// Note: could not figure out or test max precision for following modles since we don't have them in our es currently
	// Update theses accordingly when models are available in the data bas:w
	// Might want to look at https://gitlab.uncharted.software/WM/wm-maas-ingest/blob/master/pkg/maas/model/resolutions.go for model resolution
	CHIRPS:                99,
	CLEM:                  99,
	Cropland:              99,
	FSC:                   99,
	MultiTwist:            99,
	Population:            99,
	WorldPopulationAfrica: 99,
	YieldAnomaliesLPJML:   99,
}
