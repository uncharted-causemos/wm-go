package elastic

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
	MarketPrice:  15,
	PIHM:         16,

	// Note: could not figure out or test max precision for following modles since we don't have them in our es currently
	// Update theses accordingly when models are available in the data base
	APSIM:                 99,
	CHIRPS:                99,
	CLEM:                  99,
	Cropland:              99,
	DSSAT:                 99,
	FloodIndex:            99,
	FSC:                   99,
	LPJML:                 99,
	LPJMLHistoric:         99,
	MultiTwist:            99,
	Population:            99,
	WorldPopulationAfrica: 99,
	YieldAnomaliesLPJML:   99,
}
