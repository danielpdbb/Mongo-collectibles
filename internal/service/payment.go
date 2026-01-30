package service

type PaymentRequest struct {
	Amount int
	Method string
	Name   string
	Email  string
}

func CreatePayMongoPayment(req PaymentRequest) string {
	// Placeholder for real PayMongo API integration
	return "paymongo_reference_id"
}
