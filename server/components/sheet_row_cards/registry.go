package sheet_row_cards

import (
	"fmt"
	"reflect"
	"time"

	"github.com/a-h/templ"
)

// CardRenderer is a function type that renders a card template
type CardRenderer func(row any) templ.Component

// CardType represents a type of card with its row type and render function
type CardType struct {
	RowType    reflect.Type
	RenderFunc CardRenderer
}

// cardTypes stores mappings between card type names and their definitions
var cardTypes = make(map[string]CardType)

// RegisterCardType registers a new card type with the given name
func RegisterCardType[T any](name string, renderFunc func(row any) templ.Component) {
	var t T
	cardTypes[name] = CardType{
		RowType:    reflect.TypeOf(t),
		RenderFunc: renderFunc,
	}
}

// GetCardType returns the card type for the given name
func GetCardType(name string) (CardType, bool) {
	ct, ok := cardTypes[name]
	return ct, ok
}

// CompanyField represents a company name that may have a hyperlink
type CompanyField struct {
	Text string
	Link string
}

// convertValue converts a raw value from the sheet to the appropriate Go type
func convertValue(value interface{}, fieldType reflect.Type) (interface{}, error) {
	if value == nil {
		return reflect.Zero(fieldType).Interface(), nil
	}

	switch fieldType.Name() {
	case "Time":
		dateStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for time.Time, got %T", value)
		}
		// Handle empty dates by returning zero time
		if dateStr == "" {
			return time.Time{}, nil
		}
		
		// Try parsing with different formats
		formats := []string{
			"1/2/06",      // M/D/YY (2-digit year)
			"01/02/06",    // MM/DD/YY (2-digit year)
			"1/2/2006",    // M/D/YYYY
			"01/02/2006",  // MM/DD/YYYY
			"2006-01-02",  // YYYY-MM-DD
			time.RFC3339,  // ISO format
			"January 2, 2006", // Month D, YYYY
			"Jan 2, 2006",    // Mon D, YYYY
		}

		var parseErr error
		for _, format := range formats {
			if t, err := time.Parse(format, dateStr); err == nil {
				return t, nil
			} else {
				parseErr = err
			}
		}
		return nil, fmt.Errorf("failed to parse date with any format: %w", parseErr)
	case "CompanyField":
		// Handle CompanyField type
		switch v := value.(type) {
		case map[string]interface{}:
			return CompanyField{
				Text: v["text"].(string),
				Link: v["link"].(string),
			}, nil
		case string:
			return CompanyField{
				Text: v,
				Link: "",
			}, nil
		default:
			return nil, fmt.Errorf("unexpected type for CompanyField: %T", value)
		}
	default:
		// For other types, handle both string and map values
		switch v := value.(type) {
		case string:
			return v, nil
		case map[string]interface{}:
			// If it's a map with a "text" field, use that
			if text, ok := v["text"].(string); ok {
				return text, nil
			}
			// Otherwise try to convert the whole map to JSON
			return fmt.Sprintf("%v", v), nil
		case map[string]string:
			// If it's a map with a "text" field, use that
			if text, ok := v["text"]; ok {
				return text, nil
			}
			// Otherwise try to convert the whole map to JSON
			return fmt.Sprintf("%v", v), nil
		default:
			return nil, fmt.Errorf("expected string or map, got %T", value)
		}
	}
}

// CreateRowFromData creates a struct of type T from raw sheet data
func CreateRowFromData[T any](data []interface{}) (T, error) {
	var result T
	resultValue := reflect.ValueOf(&result).Elem()
	resultType := resultValue.Type()

	// Create a map of column names to field indices
	colToField := make(map[string]int)
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		if colName, ok := field.Tag.Lookup("col"); ok {
			colToField[colName] = i
		}
	}

	// Fill in the struct fields
	for colName, fieldIdx := range colToField {
		// Find the column index in the data
		colIdx := -1
		for i, col := range data {
			if str, ok := col.(string); ok && str == colName {
				colIdx = i
				break
			}
		}

		if colIdx == -1 || colIdx+1 >= len(data) {
			continue // Skip if column not found or no value
		}

		field := resultValue.Field(fieldIdx)
		convertedValue, err := convertValue(data[colIdx+1], field.Type())
		if err != nil {
			return result, fmt.Errorf("failed to convert value for field %s: %w", colName, err)
		}

		field.Set(reflect.ValueOf(convertedValue))
	}

	return result, nil
}

// ParseRowData parses a row of data into the appropriate struct type using reflection
func ParseRowData(cardType CardType, row []interface{}, colMap map[string]int) (any, error) {
	// Create a new instance of the row type
	rowValue := reflect.New(cardType.RowType).Elem()

	// For each field in the struct
	for i := 0; i < cardType.RowType.NumField(); i++ {
		field := cardType.RowType.Field(i)
		colName, ok := field.Tag.Lookup("col")
		if !ok {
			continue // Skip fields without a col tag
		}

		colIdx, ok := colMap[colName]
		if !ok || colIdx >= len(row) {
			continue // Skip if column not found or index out of range
		}

		// Get the value from the row
		value := row[colIdx]

		// Convert and set the field value
		convertedValue, err := convertValue(value, field.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for field %s: %w", field.Name, err)
		}

		rowValue.Field(i).Set(reflect.ValueOf(convertedValue))
	}

	return rowValue.Interface(), nil
}

// RegisterCardRenderer registers a card renderer for a specific row type
func RegisterCardRenderer[T any](name string, renderer func(row T) templ.Component) {
	cardTypes[name] = CardType{
		RowType: reflect.TypeOf((*T)(nil)).Elem(),
		RenderFunc: func(row any) templ.Component {
			return renderer(row.(T))
		},
	}
}

// GetCardRenderer returns the card renderer for a given row type
func GetCardRenderer(name string) (CardRenderer, error) {
	ct, ok := cardTypes[name]
	if !ok {
		return nil, fmt.Errorf("no card renderer registered for type: %s", name)
	}
	return ct.RenderFunc, nil
}
