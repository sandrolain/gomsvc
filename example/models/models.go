package models

type HelloBody struct {
	Type   string `json:"ty" validate:"required,min=3,max=32"`
	Salary int    `json:"sa" validate:"required,number"`
}

type HelloData struct {
	XToken string    `reqHeader:"X-Token"`
	Body   HelloBody `req:"body"`
	Num    float64   `query:"num" validate:"min=2"`
}
