package storage

type UserFilter struct {
	Name   string
	MinAge int
	MaxAge int
	Limit  int
	Offset int
}
