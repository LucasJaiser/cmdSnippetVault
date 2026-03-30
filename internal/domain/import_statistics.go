package domain

// ImportStatistics holds the results of a batch import operation.
type ImportStatistics struct {
	Created    int
	Duplicates int
	Rejected   int
}
