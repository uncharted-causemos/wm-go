# ES Resources

## Datacube 
Data cube is basically aggregated metadata for model output / indicator useful for faceting/searching. 

#### Fields 

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `type`  | enum | type of data cube, 'model' or 'indicator'  | keyword |
| `model`  | string | model name | keyword 
| `category`  | []string | list of model category eg. ["Agriculture", "Economic"] | keyword |
| `model_description` | string |  model description  | text |
| `label` | string |  model label  | text |
| `maintainer` | string |  model maintainer/source  | text |
| `output_name`  | string | output variable name  | keyword |
| `output_description`  | string | output description  | text | 
| `output_units`  | string | output units | keyword |
| `parameters` | []object | list of model parameters, `[]{ name string, type string, description string }` | nested |
| `parameters[].name` | string | parameter name | keyword |
| `parameters[].type` | string | parameter type | keyword |
| `parameters[].description` | string | parameter description | text |
| `concepts`  | []object | list of relevant concepts mapped to the output, `[]{ name string, score number }` | nested |
| `concepts[].name`  | string | concept name | keyword or text? |
| `concepts[].score`  | float | concept relevance score to this model output | float |
| `region` | string | name of the region that the data cube (model output) belongs to | keyword |
| `period` | object | date range that's covered by the output, `{ gte, lte }` | date_range |

***TODO:*** Add indicator metadata and update fields

#### Example: 
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

	"parameters": [{
			"description": "The season for the given run. May supercede planting_start and planting_end.",
			"name": "season",
			"type": "ChoiceParameter"
		},
		{
			"description": "The crop for the given model run.",
			"name": "crop",
			"type": "ChoiceParameter"
		},
		{
			"description": "The number of pixel predictions DSSAT will make. Setting samples to 0 returns the  entire geography (all Ethiopia) which is quite large.",
			"name": "samples",
			"type": "NumberParameter"
		},
		{
			"description": "The management practice to model. maize_rf_highN corresponds to a high nitrogen management  practice. maize_irrig corresponds to a high nitrogen, irrigated management practice. maize_rf_0N  corresponds to a subsistence management practice. maize_rf_lowN corresponds to a low nitrogen  managemet practice. If set to combined, all practices are produced. ",
			"name": "management_practice",
			"type": "ChoiceParameter"
		},
		{
			"description": "The year to begin the simulation. The earliest possible year to begin is 1984 and the latest is  2019.",
			"name": "start_year",
			"type": "TimeParameter"
		},
		{
			"description": "The number of years to run the simulation. If start_year + number_years - 1 > 2018 then this  will be set such that your simulation runs through 2018.",
			"name": "number_years",
			"type": "TimeParameter"
		},
		{
			"description": "The degree to perturb rainfall from the baseline model. This should be a real number,  which, if 0, would indicate no rainfall in any district. If 1 it would indicate rainfall matching baseline estimates. 1.25 would indicate a 25% increase in rainfall from off the baseline estimate.",
			"name": "rainfall",
			"type": "NumberParameter"
		},
		{
			"description": "This a scalar between 0 and 200 which represents fertilizer in kg/ha. 100 is considered the  baseline amount (per management practice), so anything above 100 represents additional  fertilizer usage/availability and anything below 100 represents decreased fertilzer (per  management practice).",
			"name": "fertilizer",
			"type": "NumberParameter"
		},
		{
			"description": "This is the month and day in \"mm-dd\" format when planting should begin. This allows the modeler  to simulate various planting seasons (such as Belg and Maher).",
			"name": "planting_start",
			"type": "TimeParameter"
		},
		{
			"description": "This is the month and day in \"mm-dd\" format when planting should end. This allows the modeler  to simulate various planting seasons (such as Belg and Maher). This must be after the  planting_start parameter.",
			"name": "planting_end",
			"type": "TimeParameter"
		},
		{
			"description": "This is the number, in days, that the planting window was shifted",
			"name": "planting_window_shift",
			"type": "NumberParameter"
		}

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
```
#### Important Notes:
  * `region` - We may want to have multiple fields for every regional levels like `country`, `state`. Or maybe consider a list of regions (countries, states, etc). eg. `["Ethiopia", "South Sudan"]` if output covers multiple regions.
  * `period` may need to be a list of periods, if model output has multiple runs with different time intervals
  * Having more fields that can be used for searching and faceting on would be nice. eg.  `metrics`, `items`, `source` that we don't currently have or not able to retrieve. 

## Run
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
	"timestamp": 0
}
```

## Output
Model output

***TODO:*** Define model output schema here
#### Fields 

| Field  | Type | Description | ES Mapping
| ------------- | ------------- | ------------- | ------------- |
| `geo`  | string | Lat lng, `{lat, lon}`  | geo_point |
| `model`  | string | Model name | keyword |
| `run_id`  | string | Model run Id | keyword |
| `timestamp`  | timestamp | Timestamp | date |
| `region`  | string  | Region where the points belong to | keyword
... More

#### Example

```
```
#### Important Notes:
  * `timestamp` - In order to enable comparison between model output, It's ideal to have this to be normalized and aggregated to certain resolution across all model outputs. Currently we aggregate the values to monthly timestamps using average but it would be ideal to use the default agg function set by modellers.
  * `region` - We may want to have multiple fields for different level of geographical regions, like, `county`, `state`, etc.


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
  /datacubes/facets?facets=["parameters.name", "region"]&search=crop&filters={ clauses: [ { field: "category", isNot: false, operand: "or", values: ["Economic"] }, { field: "parameters.name", isNot: false, operand: "or", values: ["rainfall", "fertilizer" ] } }]}}

Response:

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
 - **sort_by** sort by provided field
 - **limit** limit

#### Example

### GET /output/{runId}/timeseries
Temporal timeseries aggregation of the ouput with given run ID

### GET /output/tiles/{z}/{x}/{y}
MVT tile representation of the model output

#### Parameters
 - **specs** list of tile specs. eg. `specs=[
						{"model":"population", "runId":"rid", "feature":"f1", "date":"2020-01", "valueProp": "v1"},
					  {"model":"DSSAT", "runId":"rid2", "feature":"f2", "date":"2020-02", "valueProp": "v2"}
					]`
