package _excel

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelTag is the tag used to map struct fields to Excel columns
const ExcelTag = "excel"

// ExcelConverter provides methods to convert between Excel rows and Go structs
type ExcelConverter struct {
	file *excelize.File
}

// NewExcelConverter creates a new Excel converter with the provided Excel file
func NewExcelConverter(f *excelize.File) *ExcelConverter {
	return &ExcelConverter{
		file: f,
	}
}

// RowToStruct converts an Excel row to a struct
// sheet: Excel sheet name
// rowIndex: row index (1-based, as in Excel)
// headers: map of column names to column indices (0-based)
// dest: pointer to the destination struct
func (c *ExcelConverter) RowToStruct(sheet string, rowIndex int, headers map[string]int, dest interface{}) error {
	// Check if dest is a pointer to a struct
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	// Get struct value
	destElem := destValue.Elem()
	destType := destElem.Type()

	// Read row from Excel
	row, err := c.file.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("cannot read sheet %s: %w", sheet, err)
	}

	// Check if rowIndex is valid
	if rowIndex <= 0 || rowIndex > len(row) {
		return fmt.Errorf("rowIndex %d is invalid, sheet has %d rows", rowIndex, len(row))
	}

	// Get row data
	rowData := row[rowIndex-1]

	// Iterate through struct fields
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		fieldValue := destElem.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get excel tag
		excelTag := field.Tag.Get(ExcelTag)
		if excelTag == "" || excelTag == "-" {
			continue
		}

		// Find column index from headers
		colIndex, ok := headers[excelTag]
		if !ok {
			continue
		}

		// Check if column index is valid
		if colIndex < 0 || colIndex >= len(rowData) {
			continue
		}

		// Get value from Excel
		cellValue := rowData[colIndex]

		// Convert value to field's data type
		if err := c.setCellValue(fieldValue, cellValue); err != nil {
			return fmt.Errorf("cannot set value for field %s: %w", field.Name, err)
		}
	}

	return nil
}

// HeadersToMap creates a map of column names to column indices from the header row
func (c *ExcelConverter) HeadersToMap(sheet string, headerRowIndex int) (map[string]int, error) {
	// Read header row from Excel
	rows, err := c.file.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("cannot read sheet %s: %w", sheet, err)
	}

	// Check if headerRowIndex is valid
	if headerRowIndex <= 0 || headerRowIndex > len(rows) {
		return nil, fmt.Errorf("headerRowIndex %d is invalid, sheet has %d rows", headerRowIndex, len(rows))
	}

	// Get header row data
	headerRow := rows[headerRowIndex-1]

	// Create map of column names to column indices
	headers := make(map[string]int)
	for i, header := range headerRow {
		headers[header] = i
	}

	return headers, nil
}

// setCellValue sets the value of a struct field from an Excel cell value string
func (c *ExcelConverter) setCellValue(fieldValue reflect.Value, cellValue string) error {
	// If string is empty, don't set value
	if cellValue == "" {
		return nil
	}

	// Process based on field's data type
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(cellValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(cellValue, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intValue)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(cellValue, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintValue)

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(cellValue, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatValue)

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(cellValue)
		if err != nil {
			// Handle special cases
			cellValueLower := strings.ToLower(cellValue)
			if cellValueLower == "yes" || cellValueLower == "y" || cellValueLower == "true" || cellValueLower == "t" || cellValueLower == "1" {
				fieldValue.SetBool(true)
			} else {
				fieldValue.SetBool(false)
			}
		} else {
			fieldValue.SetBool(boolValue)
		}

	case reflect.Struct:
		// Handle common struct types
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			// Try common time formats
			formats := []string{
				"2006-01-02",
				"02/01/2006",
				"02-01-2006",
				"2006/01/02",
				"2006-01-02 15:04:05",
				"02/01/2006 15:04:05",
				time.RFC3339,
			}

			var timeValue time.Time
			var err error
			for _, format := range formats {
				timeValue, err = time.Parse(format, cellValue)
				if err == nil {
					fieldValue.Set(reflect.ValueOf(timeValue))
					return nil
				}
			}
			return fmt.Errorf("cannot convert '%s' to time.Time", cellValue)
		}
		return fmt.Errorf("unsupported struct type %s", fieldValue.Type().Name())

	case reflect.Ptr:
		// Handle pointers
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return c.setCellValue(fieldValue.Elem(), cellValue)

	case reflect.Slice:
		// Handle slices (assuming comma-separated values)
		values := strings.Split(cellValue, ",")
		sliceType := fieldValue.Type().Elem()
		slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		for i, val := range values {
			val = strings.TrimSpace(val)
			elemValue := reflect.New(sliceType).Elem()
			if err := c.setCellValue(elemValue, val); err != nil {
				return err
			}
			slice.Index(i).Set(elemValue)
		}
		fieldValue.Set(slice)

	default:
		return fmt.Errorf("unsupported data type %s", fieldValue.Kind())
	}

	return nil
}
