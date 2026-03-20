package domain

type Tag struct {
	ID   int64
	Name string
}

type TagWithCount struct {
	Name  string
	Count int
}
