package structs

type OrderRequest struct {
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	Products map[string]int `json:"products"` // productID -> quantity
}
