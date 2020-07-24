# Resources
Preferably we want to have all the resources for metadata in ES for easier searching/querying. 
Following resources are new or updated ones that would be needed by Causmos in addition to the existing ones current Maas api covers (eg. `Model` `Concept Mapping` etc)
## Datacube (ES) 
Datacube is basically aggregated metadata for the model output / indicator useful for faceting/searching. 

#### Fields 

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `type`  | enum | type of data cube, 'model' or 'indicator'  | keyword |
| `model`  | string | output model name | keyword 
| `category`  | []string | list of model category eg. ["Agriculture", "Economic"] | keyword |
| `model_description` | string |  model description  | text |
| `label` | string |  model label  | text |
| `maintainer` | string |  model maintainer/source  | text |
| `source` | string |  model/indicator source (eg. FAO) | keyword |
| `output_name`  | string | output variable name  | keyword |
| `output_description`  | string | output description  | text | 
| `output_units`  | string | output units | keyword |
| `output_units_description`  | string | output units description (eg. meters)  | keyword | 
| `parameters` | []string | list of model parameter names for the output | keyword |
| `parameter_descriptions` | []string | list of model parameter descriptions to be used for text matching/searching | text |
| `concepts`  | []object | list of relevant concepts mapped to the indicator/model output, `[]{ name string, score number }` | nested |
| `concepts[].name`  | string | concept name | keyword |
| `concepts[].score`  | float | concept relevance score to this model/indicator output | float |
| `country`| []string | Countries covered by the output | keyword |
| `admin1`| []string | First level admin regions covered by the output | keyword |
| `admin2`| []string | Second level admin regions covered by the output | keyword |
| `periods` | []daterange | date ranges that's covered by the output runs, `[]{ gte, lte }` | date_range |

#### Example
```
 {
	"type": "model",

	"model": "DSSAT",
	"category": [
		"Agriculture",
		"Economic"
	],
	"model_description": "The Decision Support System for Agrotechnology Transfer (DSSAT) comprises dynamic crop growth simulation model for over 40 crops. The model simulates growth development; and yield as a function of the soil-plant-atmosphere dynamics.",
	"label": "Decision Support System for Agrotechnology Transfer",
	"maintainer": "Cheryl Porter, cporter@ufl.edu",


	"output_description": "Harvested weight at harvest (kg/ha)",
	"output_name": "HWAH",
	"output_units": "kg/ha",
	"output_units_description": "Kilogram per hectare"

	"parameters": ["season","crop","samples","management_practice","start_year","number_years","rainfall","fertilizer","planting_start", "planting_end","planting_window_shift"],

	"parameter_descriptions": [
		"The season for the given run. May supercede planting_start and planting_end.",
		"The crop for the given model run.",
		"The number of pixel predictions DSSAT will make. Setting samples to 0 returns the  entire geography (all Ethiopia) which is quite large.",
		"The management practice to model. maize_rf_highN corresponds to a high nitrogen management  practice. maize_irrig corresponds to a high nitrogen, irrigated management practice. maize_rf_0N  corresponds to a subsistence management practice. maize_rf_lowN corresponds to a low nitrogen  managemet practice. If set to combined, all practices are produced. ",
		"The year to begin the simulation. The earliest possible year to begin is 1984 and the latest is  2019.",
		"The number of years to run the simulation. If start_year + number_years - 1 > 2018 then this  will be set such that your simulation runs through 2018.",
		"The degree to perturb rainfall from the baseline model. This should be a real number,  which, if 0, would indicate no rainfall in any district. If 1 it would indicate rainfall matching baseline estimates. 1.25 would indicate a 25% increase in rainfall from off the baseline estimate.",
		"This a scalar between 0 and 200 which represents fertilizer in kg/ha. 100 is considered the  baseline amount (per management practice), so anything above 100 represents additional  fertilizer usage/availability and anything below 100 represents decreased fertilzer (per  management practice).",
		"This is the month and day in \"mm-dd\" format when planting should begin. This allows the modeler  to simulate various planting seasons (such as Belg and Maher).",
		"This is the month and day in \"mm-dd\" format when planting should end. This allows the modeler  to simulate various planting seasons (such as Belg and Maher). This must be after the  planting_start parameter.",
		"This is the number, in days, that the planting window was shifted"
	],
	"concepts": [{
		"name": "wm/concept/causal_factor/agriculture/crop_production",
		"score": "0.6544816493988037"
	}],

	"country": ["Ethiopia"],
	"admin1": ["Oromia", "Somali", "Afar"],
	"admin2": ["Borena", "Guji", "Bale", "Nogob", "... and more"],

  "periods": [
		{
			"gte": "2015-01",
			"lte": "2016-02"
		},
		{
			"gte": "2017-01",
			"lte": "2019-02"
		},
	]
}
```
#### Important Notes:
  * `period` may need to be a list of periods, if model output has multiple runs with different time intervals
  * Any other metadata fields that can be used for searching and faceting on would be useful. Such as  `metrics`, `items`, or `source` that we don't currently have.

## Run (ES)
Model run with parameters/configs used for the run. (ie. Run results in current maas api)

#### Fields 

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `id`  | string | Run ID  | keyword |
| `model`  | string | Model name | keyword |
| `parameters`  | []object | Parameters for the run, `[]{ name, type, value}` | nested
| `parameters[].name`  | string | Parameter name | keyword
| `parameters[].type`  | string | Parameter type | keyword
| `parameters[].value`  | string | Parameter value | keyword
| `timestamp`  | timestamp | Epoch timestamp when the model run was initiated | date
| `country`| []string | Countries covered by the run output | keyword |
| `admin1`| []string | First level admin regions covered by the output | keyword |
| `admin2`| []string | Second level admin regions covered by the output | keyword |
| `period`| daterange | Date range covered by the output, `{ gte, lte }` | date_range |
| `status`  | string  | Run status eg. ["SUCCESS", "FAIL", "PENDING"] | string
| `output`  | string | URI for accessing raw output (eg. S3 uri) | string |
| `output_normalized`  | timestamp | URI for accessing normalized output (eg. s3 uri) | string

#### Example
```
{
	"id": "671e299cff0d6ee2e16d47c0e8f4ab633cb79525c8bb5e4f8f48a1c33ce757fa" 
	"model": "DSSAT"
	"parameters": [
		{
			"name": "season",
			"season": "Meher",
			"type": "ChoiceParameter"
		},
		{
			"name": "crop",
			"value": "teff, 
			"type": "ChoiceParameter"
		},
		{
			"name": "samples",
			"value": 0,
			"type": "NumberParameter"
		},
		{
			"name": "management_practice",
			"value": null,
			"type": "ChoiceParameter"
		},
		{
			"name": "number_years",
			"value": 10,
			"type": "TimeParameter"
		},
		{
			"name": "rainfall",
			"value": 1,
			"type": "NumberParameter"
		},
		{
			"name": "fertilizer",
			"value": 25,
			"type": "NumberParameter"
		},
		{
			"name": "planting_window_shift",
			"value": 0,
			"type": "NumberParameter"
		}
	],
	"timestamp": 0,

	"country": ["Ethiopia"],
	"admin1": ["Oromia", "Somali", "Afar"],
	"admin2": ["Borena", "Guji", "Bale", "Nogob", "... and more"],
  "period": {
		"gte": "2015-01",
		"lte": "2016-02"
	},
	"status": "SUCCESS",
	"output": "https://s3.amazonaws.com/world-modelers/results/DSSAT_results/pp_ETH_Oroima_Teff_Meher__rf_0N__fen_tot25__erain1.0__pfrst0.csv"
	"output_normalized: "https://s3.amazonaws.com/world-modelers/results_normalized/DSSAT/671e299cff0d6ee2e16d47c0e8f4ab633cb79525c8bb5e4f8f48a1c33ce757fa"
}
```
## Parameter
Similar to current parameters model in current maas api but add parameter `units` and `units_description` for applicable ones. 

### Fields 

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `model`  | string | Model | keyword |
| `name`  | string | Parameter type | keyword
| `description`  | string | Parameter description | keyword
| `units`  | string | Parameter unit | keyword
| `units_description`  | string | Parameter unit description | keyword
| `type`  | enum | Parameter type (ie. ChoiceParameter, NumberParameter) | keyword
| `maximum`  | number | Maximum number (if `type` is NumberParameter) | double
| `minimum`  | number | Minimum number (if `type` is NumberParameter) | double
| `choices`  | []string | Set of choices (if `type` is ChoiceParameter) | keyword |
| `default`  | string | Default choice | keyword

### Example 
````
  {
		"model": "DSSAT",
    "choices": [
      "Meher",
      "Belg"
    ],
    "default": "meher",
    "description": "The season for the given run. May supercede planting_start and planting_end.",
    "name": "season",
    "type": "ChoiceParameter"
  },
  {
		"model": "DSSAT",
    "default": 100,
    "description": "This a scalar between 0 and 200 which represents fertilizer in kg/ha. 100 is considered the  baseline amount (per management practice), so anything above 100 represents additional  fertilizer usage/availability and anything below 100 represents decreased fertilzer (per  management practice).",
    "maximum": 200,
    "minimum": 0,
		"units": "kg/ha",
		"units_description: "Kilogram per hectare"
    "name": "fertilizer",
    "type": "NumberParameter"
  },

````

## Other Resources (ES)
Preferably have other resources in ES that are not mentioned above like `Model`, `Concept` or `Concept Mappings` that existing maas api provides.

### Fields Concept Mapping

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `concept`  | string | Concept name | keyword |
| `score`  | number | Mapping score | float
| `type`  | enum | Target type (ie. model, parameter, output) | keyword
| `target`  | string | Target of mapping (ie. a model name, output id, or parameter id) | keyword

### Example 
````
  Concept mapping example
	One example uses case would be querying all available concept names for certain type, model, output etc.  
  {
		"concept": "wm/concept/causal_factor/agriculture/crop_production"
		"score": 0.7293715476989746
		"type": "output"
		"target": "DSSAT-HWAH"
  },
  {
		"concept": "<concept name>"
		"score": "<relevance score>"
		"type": "model|parameter|output"
		"target": "model name | output id | parameter id"
  },
	... 

````

## Output (AWS S3)
Normalized model output data. Preferably in S3 bucket and partitioned using parquet format. (eg. `/DSSAT/062d9473d76a01db9f255e0807ce91b1f3ca6caba81b92a53ae530da9b6e2d78/{partitioned_filename}.parquet`). 

#### Fields 

| Field  | Type | Description |
| ------------- | ------------- | ------------- |
| `run_id`  | string | Model run Id |
| `model`  | string | Model name |
| `feature` (or output)  | string | model output variable name |
| `value` (or output_value)  | float | model output value |
| `admin0`  | string  | 0 level admin region ie. Country or nation level |
| `admin1`  | string  | First level admin region (eg. state, province etc) |
| `admin2`  | string  | Second level admin region (eg. county, district etc) |
| `adminN`  | string  | 1-n level admin region |
| `lat`  | float | Latitude |
| `lng`  | float | Longitude  |
| `timestamp`  | timestamp | Timestamp |

#### Example

```
{
	"run_id": 	"df3f4f29f433ca66ca71cf5764c757559a1f1268a53aba44255e329c128cb263",
	"model": "cropland_model",
	"admin0": "Ethiopia",
	"admin1": "Oromia",
	"admin2": "Borena",
	"feature": "cropland",
	"value": 0.004375,
	"lat": 3.92128,
	"lon": 38.057093
	"timestamp": "2012-01-01T00:00:00Z",
}
```
#### Important Notes:
  * `timestamp` - In order to enable comparison between model output, It's ideal to have this to be normalized and aggregated (preferably using agg function set by expert modeller) up to certain resolution across all model outputs.
* `runId` or `model` may be omitted since s3 bucket file path likely includes them.

# Causemos REST API for new Data view

### GET /datacubes

#### Parameters
 - **search** search term used for text matching on text type fields in addition to filters
 - **filters** fliters object eg. `filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}`

#### Example

```
Request: 
  /datacubes?search=rainfall&filters={"clauses":[{"field":"model","operand":"or","isNot":false,"values":["G-Range", "PIHM", "malnutrition_model"]}]}

Response: 
 [
   {
        "id": "6cfd6f41-21dc-4f84-85a5-da6a8f4707d4",
        "type": "model",
        "model": "G-Range",
        "category": [
            "Agriculture"
        ],
        "model_description": "G-Range is a global rangeland model that simulates generalised changes in rangelands through time. Spatial data and a set of parameters that describe plant growth in landscape units combine with computer code representing ecological processes to represent soil nutrient and water dynamics, vegetation growth, fire, and wild and domestic animal offtake. The model is spatial, with areas of the world divided into square cells. Those cells that are rangelands have ecosystem dynamics simulated. A graphical user interface allows users to explore model output.",
        "label": "APSIMx-G-Range",
        "maintainer": "Andrew Moore, Andrew.Moore@csiro.au",
        "source": "Andrew Moore, Andrew.Moore@csiro.au",
        "output_name": "total_anomaly_herbage_prodn",
        "output_description": "Difference between herbage aboveground net primary production from rangelands and its long-term average value",
        "output_units": "quintal",
        "output_units_description": "",
        "parameters": [
            "climate_anomalies",
            "cereal_prodn_pctile",
            "cereal_prodn_tercile",
            "irrigation",
            "additional_extension",
            "temperature",
            "sowing_window_shift",
            "fertilizer",
            "rainfall"
        ],
        "parameter_descriptions": [
            "One of 5 classes based on the mean 2018-19 cropping-year (March-February) rainfall and temperature anomalies in the climate ensemble member. Ensemble members where the root-mean-square anomaly of temperature and precipitation are within 0.9 standard deviations are \"midrange\"; otherwise ensemble members are classified according to the quadrant in which they fall.",
            "Ranking of total national production of the 5 cereals as modelled under 2018 land use and practices, expressed as a percentile (the zero percentile is lowest)",
            "Grouping of climate ensemble members according to terciles of total national production of the 5 cereals as modelled under 2018 land use and practices",
            "Average proportion of cereal area that is irrigated across Ethiopia. Local proportions vary spatially and with the type of crop",
            "For this scenario, an \"extension package\" means the adoption of both improved crop cultivars and chemical fertilizer application. The value is the proportion of land **not already using \"extension package\"** that is converted to management under the \"extension package\". For example, if 20% of maize crops in a grid-cell already use improved cultivars plus fertilizer, then 40% \"additional extension package\" will increase the overall level to (20% + 40% x (100%-20%)) = 52%",
            "Change applied to maximum and minimum air temperature in every day of the climate record in the counterfactual",
            "Shift (measured in days) in the date range over which crops are sown in response to a sufficiently large rainfall event",
            "Additional N fertilizer applied at sowing, over and above the rate that is specific to a location, crop and management system",
            "Multiplier applied to daily rainfall in every day of the climate record in the counterfactual"
        ],
        "concepts": [
            {
                "name": "wm/concept/causal_factor/crisis_and_disaster/environmental_disasters/crop_failure",
                "score": 0.5259248614311218
            },
            {
                "name": "wm/concept/causal_factor/economic_and_commerce/economic_activity/market/price_or_cost/food_price",
                "score": 0.5313275456428528
            },
            {
                "name": "wm/concept/causal_factor/access/water_access",
                "score": 0.5188208222389221
            },
            {
                "name": "wm/concept/causal_factor/trend",
                "score": 0.5317885875701904
            },
            {
                "name": "wm/concept/causal_factor/economic_and_commerce/economic_activity/market/price_or_cost/cost_of_living",
                "score": 0.5160562992095947
            },
            {
                "name": "wm/concept/causal_factor/economic_and_commerce/economic_activity/market/revenue/farmer_income",
                "score": 0.5530139207839966
            },
            {
                "name": "wm/concept/causal_factor/interventions/provision_of_goods_and_services/household_solid_waste_management",
                "score": 0.590152382850647
            },
            {
                "name": "wm/concept/causal_factor/interventions/provision_of_goods_and_services/provision_of_credit_and_training_for_income_generation",
                "score": 0.5345229506492615
            },
            {
                "name": "wm/concept/causal_factor/economic_and_commerce/economic_activity/market/price_or_cost/cost_of_transportation",
                "score": 0.5599560737609863
            },
            {
                "name": "wm/concept/causal_factor/interventions/provision_of_goods_and_services/point_of_use_water_treatment_at_household_level",
                "score": 0.5954615473747253
            }
        ],
        "country": [
            "Ethiopia"
        ],
        "admin1": [
            "Oromia",
            "Somali",
            "Amhara",
            "Southern Nations, Nationalities and Peoples",
            "Afar",
            "Tigray",
            "Benshangul-Gumaz",
            "Gambela Peoples",
            "Dire Dawa",
            "Harari People"
        ],
        "admin2": [
            "Afder",
            "Bale",
            "Borena",
            "Doolo",
            "Liben",
            "Jarar",
            "Afar Zone 1",
            "Semen Gondar",
            "Siti",
            "Shabelle"
        ],
        "period": [
            {
                "gte": "1525132800000",
                "lte": "1554076800000"
            }
        ]
    },
  ...
 ]
```


### GET /datacubes/facets

#### Parameters
 - **search** search term used for text matching on text type fields in addition to filters
 - **filters** fliters object eg. `filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}`
 - **facets** list of facet(attribute) names

#### Example

```
Request:
  /datacubes/facets?facets=["parameters.name", "country"]&search=crop&filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}}

Response:
{
	"parameters.name": [
		{
			"key": "rainfall",
			"count" 12
		},
		{
			"key": "fertilizer",
			"count" 4
		},
		...
	],
	"country": [
		{
			"key": "Ethiopia",
			"count": 43
		},
		{
			"key": "South Sudan",
			"count": 2
		}
	]
}
```


### GET /models/{modelId}/parameters
Mirrors `https://model-service.worldmodelers.com/model_parameters/{ModelName}`

#### Path
 - **modelId** model name


#### Example
```
Request:

GET /models/DSSAT/parameters

Response: 
[
  {
    "default": "05-20",
    "description": "This is the month and day in \"mm-dd\" format when planting should end. This allows the modeler  to simulate various planting seasons (such as Belg and Maher). This must be after the  planting_start parameter.",
    "maximum": "12-31",
    "minimum": "01-01",
    "name": "planting_end",
    "type": "TimeParameter"
  },
  {
    "default": 0,
    "description": "This is the number, in days, that the planting window was shifted",
    "maximum": 30,
    "minimum": -30,
    "name": "planting_window_shift",
    "type": "NumberParameter"
  }
	...
]
```

### GET /models/{modelId}/runs
Get all runs for the model

#### Parameters
 - **sort_by** Sort the runs by the provided sort_by field
 - **limit** Limits the # of results

#### Example

### GET /output/{runId}/timeseries
Temporal timeseries aggregation of the ouput with given run ID

### GET /output/tiles/{z}/{x}/{y}
MVT tile representation of the model output

#### Parameters
 - **specs** (required) List of output selection specs for the output to be included in the tile. eg. `specs=[{"model":"G-Range","runId":"062d9473d76a01db9f255e0807ce91b1f3ca6caba81b92a53ae530da9b6e2d78","feature":"total_anomaly_herbage_prodn","date":"2019-04-01T00:00:00.000Z","valueProp":"G-Range:total_anomaly_herbage_prodn"},{"model":"malnutrition_model","runId":"8e62caa28c3132c4a8e6042a83a3ce0c03c86d94a764e2a13b55b484d985eecb","feature":"malnutrition cases","date":"2018-05-01T00:00:00.000Z","valueProp":"malnutrition_model:malnutrition cases"}]`
