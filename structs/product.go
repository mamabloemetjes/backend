package structs

// Size of the product
type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

// ProductType enum
type ProductType string

const (
	Flower  ProductType = "flower"
	Bouquet ProductType = "bouquet"
)

// Color enum
type Color string

const (
	ColorRed    Color = "red"
	ColorBlue   Color = "blue"
	ColorGreen  Color = "green"
	ColorYellow Color = "yellow"
	ColorBlack  Color = "black"
	ColorWhite  Color = "white"
	ColorPurple Color = "purple"
	ColorOrange Color = "orange"
	ColorPink   Color = "pink"
)
