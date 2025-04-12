package gdrive

// SheetConfig maps sheet IDs to their tab configurations
var SheetConfig = map[string]map[string]interface{}{
	"1L0dQpfj3c86mXRjADRrLshUCZrFzA3vcM_TfYxITjmc": {
		"Company List - Free Product": map[string]interface{}{
			"Component":           "FreeProductCard",
			"StructuredDataRange": "A6:G",
		},
		"Company List - Discount Codes": map[string]interface{}{
			"Component":           "DiscountCard",
			"StructuredDataRange": "A6:F",
		},
		"Company List - Free Product Pick-ups": map[string]interface{}{
			"Component":           "PickupCard",
			"StructuredDataRange": "A6:D",
		},
		"Company List - Free Services": map[string]interface{}{
			"Component":           "ServiceCard",
			"StructuredDataRange": "A6:F",
		},
	},
}
