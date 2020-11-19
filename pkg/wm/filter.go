package wm

// Operand is a type for Boolean operands.
type Operand int

// Available Boolean operands.
const (
	OperandAnd Operand = iota
	OperandOr
)

// FilterContext is a type for availbel filter contexts
type FilterContext int

// Available filter context
const (
	ContextKB FilterContext = iota
	ContextDatacube
	ContextAnalysis
	ContextIndicator
)

// Field is a type for filterable fields.
type Field int

// Available field types.
const (
	FieldBeliefScore Field = iota
	FieldCause
	FieldConcept
	FieldEffect
	FieldEvidenceSource
	FieldGroundingScore
	FieldHedging
	FieldLocation
	FieldNumEvidence
	FieldOrganization
	FieldPolarity
	FieldPublicationYear
	FieldReader
	FieldRefutingEvidence
	FieldQuality

	// Datacubes fields
	FieldDatacubeID
	FieldDatacubeType
	FieldDatacubeModel
	FieldDatacubeCategory
	FieldDatacubeLabel
	FieldDatacubeMaintainer
	FieldDatacubeSource
	FieldDatacubeOutputName
	FieldDatacubeOutputUnits
	FieldDatacubeParameters
	FieldDatacubeConceptName
	FieldDatacubeConceptScore
	FieldDatacubeCountry
	FieldDatacubeAdmin1
	FieldDatacubeAdmin2
	FieldDatacubePeriod
	FieldDatacubeSearch

	// Analysis fields
	FieldAnalysisProjectID

	// Indicator fields
	FieldIndicatorVariable
)

// Filter defines a filter to be used in the queries.
type Filter struct {
	Field        Field
	Operand      Operand
	IsNot        bool
	IntValues    []int
	StringValues []string
	Range        Range
}

// Range defines a range
type Range struct {
	Minimum  float64
	Maximum  float64
	IsClosed bool
}
