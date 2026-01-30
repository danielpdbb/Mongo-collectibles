package service

func CalculateRentalPrice(size string, days int) int {
	var dailyRate int

	switch size {
	case "S":
		dailyRate = 1000
	case "M":
		dailyRate = 5000
	case "L":
		dailyRate = 10000
	default:
		return 0
	}

	if days < 7 {
		dailyRate *= 2
	}

	return dailyRate * days
}
