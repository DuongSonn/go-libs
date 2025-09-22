# Excel Package

Excel package provides utilities for working with Excel files in Go, using the [excelize](https://github.com/xuri/excelize) library.

## Features

-   Convert Excel rows to Go structs
-   Support for various data types: string, int, float, bool, time.Time, slice, pointer
-   Map Excel columns to struct fields using `excel` tag
-   Automatic type conversion
-   Excel converter for row-to-struct conversion
-   Batch reading and writing

## Installation

```bash
go get -u go-libs/pkg/excel
go get -u github.com/xuri/excelize/v2
```

## Usage

### Define struct with excel tags

```go
type Product struct {
    ID          int       `excel:"Product ID"`
    Name        string    `excel:"Product Name"`
    Price       float64   `excel:"Price"`
    Quantity    int       `excel:"Quantity"`
    Description string    `excel:"Description"`
    IsAvailable bool      `excel:"Available"`
    Categories  []string  `excel:"Categories"`
    CreatedAt   time.Time `excel:"Created Date"`
}
```

### Using the ExcelConverter

The ExcelConverter provides methods to convert between Excel rows and Go structs:

```go
// Open an Excel file
f, err := excelize.OpenFile("products.xlsx")
if err != nil {
    log.Fatalf("Cannot open Excel file: %v", err)
}
defer f.Close()

// Create a new converter
converter := excel.NewExcelConverter(f)

// Get headers map from header row (row 1)
headers, err := converter.HeadersToMap("Sheet1", 1)
if err != nil {
    log.Fatalf("Cannot get headers: %v", err)
}

// Read row and convert to struct
var product Product
err = converter.RowToStruct("Sheet1", 2, headers, &product)
if err != nil {
    log.Fatalf("Cannot read row: %v", err)
}

// Use product
fmt.Printf("Product: %s, Price: %.2f\n", product.Name, product.Price)
```

### Reading Multiple Rows

```go
// Read multiple rows
for rowIndex := 2; rowIndex <= 10; rowIndex++ {
    var product Product
    err := converter.RowToStruct("Sheet1", rowIndex, headers, &product)
    if err != nil {
        log.Printf("Error reading row %d: %v", rowIndex, err)
        continue
    }

    // Process product
    fmt.Printf("Product: %s\n", product.Name)
}
```

### Creating a Helper Function for Batch Reading

```go
// ReadProductsFromExcel reads products from an Excel file
func ReadProductsFromExcel(filePath string) ([]Product, error) {
    // Open Excel file
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("cannot open Excel file: %w", err)
    }
    defer f.Close()

    // Create converter
    converter := excel.NewExcelConverter(f)

    // Get headers
    headers, err := converter.HeadersToMap("Sheet1", 1)
    if err != nil {
        return nil, fmt.Errorf("cannot get headers: %w", err)
    }

    // Get row count
    rows, err := f.GetRows("Sheet1")
    if err != nil {
        return nil, fmt.Errorf("cannot read sheet: %w", err)
    }

    // Read products
    products := make([]Product, 0, len(rows)-1)
    for rowIndex := 2; rowIndex <= len(rows); rowIndex++ {
        var product Product
        err := converter.RowToStruct("Sheet1", rowIndex, headers, &product)
        if err != nil {
            return nil, fmt.Errorf("error reading row %d: %w", rowIndex, err)
        }
        products = append(products, product)
    }

    return products, nil
}
```

## API Reference

### ExcelConverter

#### `NewExcelConverter(f *excelize.File) *ExcelConverter`

Creates a new Excel converter with the provided Excel file.

#### `(c *ExcelConverter) RowToStruct(sheet string, rowIndex int, headers map[string]int, dest interface{}) error`

Converts an Excel row to a struct.

-   `sheet`: Excel sheet name
-   `rowIndex`: Row index (1-based, as in Excel)
-   `headers`: Map of column names to column indices (0-based)
-   `dest`: Pointer to the destination struct

#### `(c *ExcelConverter) HeadersToMap(sheet string, headerRowIndex int) (map[string]int, error)`

Creates a map of column names to column indices from the header row.

-   `sheet`: Excel sheet name
-   `headerRowIndex`: Index of the header row (1-based, as in Excel)

## Supported Data Types

-   `string`: String values
-   `int`, `int8`, `int16`, `int32`, `int64`: Integer values
-   `uint`, `uint8`, `uint16`, `uint32`, `uint64`: Unsigned integer values
-   `float32`, `float64`: Floating-point values
-   `bool`: Boolean values (true/false, yes/no, y/n, 1/0)
-   `time.Time`: Time values (supports multiple formats)
-   `[]T`: Slices (comma-separated values)
-   `*T`: Pointers to type T

## Examples

See the [example](./example) directory for complete examples.

## License

This package is part of the go-libs project.
