package wm

// Operand is a type for Boolean operands.
type Operand int

// Available Boolean operands.
const (
	OperandAnd Operand = iota
	OperandOr
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
)

// Filter defines a filter to be used in the queries.
type Filter struct {
	Field        Field
	Operand      Operand
	IsNot        bool
	IntValues    []int
	StringValues []string
	Range        [2]float64
}
