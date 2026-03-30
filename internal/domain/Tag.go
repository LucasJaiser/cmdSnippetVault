package domain

// Tag represents a label that can be associated with snippets.
type Tag struct {
	ID   int64
	Name string
}

// TagWithCount pairs a tag name with the number of snippets it is linked to.
type TagWithCount struct {
	Name  string
	Count int
}
