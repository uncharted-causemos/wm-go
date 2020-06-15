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
    "minumum": 0,
		"units": "kg/ha",
		"units_description: "Kilogram per hectare"
    "name": "fertilizer",
    "type": "NumberParameter"
  },

````

## Other Resources (ES)
Preferably have other resources in ES that are not mentioned above like `Model`, `Concept` or `Concept Mappings` that existing maas api provides.

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
| `feature_value` (or output_value)  | float | model output value |
| `country`  | string  | Country where the point belong to |
| `state`  | string  | State where the point belong to |
| `city`  | string  | City where the point belong to |
| `admin1`  | string  | First level admin region (eg. state, province etc) |
| `admin2`  | string  | Second level admin region (eg. county, district etc) |
| `adminN`  | string  | 1-n level admin region |
| `lat`  | float | Latitude |
| `lng`  | float | Longitude  |
| `timestamp`  | timestamp | Timestamp |

#### Example

```
{
	"id": "1048401765-0-0",
	"run_id": "df3f4f29f433ca66ca71cf5764c757559a1f1268a53aba44255e329c128cb263",
	"model": "cropland_model",
	"country": "Ethiopia",
	"state": "Oromia",
	"admin1": "Oromia",
	"admin2": "Borena",
	"city": "",
	"feature_name": "cropland",
	"feature_value": 0.004375,
	"lat": 3.92128,
	"lon": 38.057093
	"timestamp": "2012-01-01T00:00:00Z",
	"updated_at": "0001-01-01T00:00:00Z"
	"created_at": "2020-01-16T04:27:56Z",
}
```
#### Important Notes:
  * `timestamp` - In order to enable comparison between model output, It's ideal to have this to be normalized and aggregated (preferably using agg function set by expert modeller) up to certain resolution across all model outputs.
  * `state` - We may not need `state` field if `admin1[1-n]` covers. It might be better to have division names normalized just using admin[1-n] since different country uses different name for division levels.
# Causemos REST API for new Data view

### GET /datacubes

#### Parameters
 - **search** search term used for text matching on text type fields in addition to filters
 - **filters** fliters object eg. `filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}`

#### Example

```
Request: 
  /datacubes?search=crop&filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}}

Response: 
 [
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

      "parameters": [{
          "description": "The degree to perturb rainfall from the baseline model. This should be a real number,  which, if 0, would indicate no rainfall in any district. If 1 it would indicate rainfall matching baseline estimates. 1.25 would indicate a 25% increase in rainfall from off the baseline estimate.",
          "name": "rainfall",
          "type": "NumberParameter"
        },
        {
          "description": "This a scalar between 0 and 200 which represents fertilizer in kg/ha. 100 is considered the  baseline amount (per management practice), so anything above 100 represents additional  fertilizer usage/availability and anything below 100 represents decreased fertilzer (per  management practice).",
          "name": "fertilizer",
          "type": "NumberParameter"
        },
        ...
      ],
      "concepts": [{
        "name": "wm/concept/causal_factor/agriculture/crop_production",
        "score": "0.6544816493988037"
      }],

      "region": "Ethiopia",

      "period": {
        "gte": "2015-01",
        "lte": "2016-02"
      }
    }
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


### GET /models/{model}/parameters
Mirrors `https://model-service.worldmodelers.com/model_parameters/{ModelName}`

#### Path
 - **model** model name


#### Example
```
Request:

GET /model/DSSAT/parameters

Response: 
[
  {
    "default": "05-20",
    "description": "This is the month and day in \"mm-dd\" format when planting should end. This allows the modeler  to simulate various planting seasons (such as Belg and Maher). This must be after the  planting_start parameter.",
    "maximum": "12-31",
    "minumum": "01-01",
    "name": "planting_end",
    "type": "TimeParameter"
  },
  {
    "default": 0,
    "description": "This is the number, in days, that the planting window was shifted",
    "maximum": 30,
    "minumum": -30,
    "name": "planting_window_shift",
    "type": "NumberParameter"
  }
	...
]
```

### GET /model/{modelId}/runs
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
