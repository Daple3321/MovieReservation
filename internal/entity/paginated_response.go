package entity

type PaginatedResponse struct {
	Items      interface{}
	Page       int
	Limit      int
	TotalItems int
	TotalPages int
}
