package structs

type OrderRequest struct {
	// Customer Data
	Name         string `json:"name" validate:"required,min=2,max=100"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone" validate:"required,min=10,max=20"`
	CustomerNote string `json:"customer_note" validate:"omitempty,max=500"`

	// Address Data
	Street     string `json:"street" validate:"required,min=2,max=200"`
	HouseNo    string `json:"house_no" validate:"required,min=1,max=10"`
	PostalCode string `json:"postal_code" validate:"required,len=7,nl_postalcode"`
	City       string `json:"city" validate:"required,min=2,max=100"`
	Country    string `json:"country" validate:"omitempty,len=2"` // ISO country code

	// Order data
	Products      map[string]int `json:"products" validate:"required,min=1,dive,keys,uuid4,endkeys,required,min=1"` // productID -> quantity
	ShippingCents int            `json:"shipping_cents"`
}
