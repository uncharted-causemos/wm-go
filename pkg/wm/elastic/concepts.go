package elastic

// GetConcepts returns concepts
func (es *ES) GetConcepts() ([]string, error) {
	// TODO: Get this from ES instead of old maas api when Galois ES is up.
	return es.modelService.GetConcepts()
}
