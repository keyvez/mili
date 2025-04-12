package model

// Category represents a resource category
type Category struct {
	Name string
}

// Resource represents a disaster resource
type Resource struct {
	Name        string
	Description string
	Category    string
	Link        string
}
