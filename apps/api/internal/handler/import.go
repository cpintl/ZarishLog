package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/cpintl/ZarishLog/apps/api/internal/pagination"
	"github.com/cpintl/ZarishLog/apps/api/internal/response"
	"github.com/gin-gonic/gin"
)

type ImportRow struct {
	Row      int               `json:"row"`
	SKU      string            `json:"sku"`
	Name     string            `json:"name"`
	Category string            `json:"category"`
	UoM      string            `json:"uom"`
	ItemType string            `json:"item_type"`
	Fields   map[string]string `json:"-"`
}

type ImportResult struct {
	Imported int           `json:"imported"`
	Skipped  int           `json:"skipped"`
	Errors   []ImportError `json:"errors,omitempty"`
}

type ImportError struct {
	Row   int    `json:"row"`
	Field string `json:"field"`
	Error string `json:"error"`
}

func ImportProducts(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		userID := c.GetString("user_id")

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			response.BadRequest(c, "file is required")
			return
		}
		defer file.Close()

		if !strings.HasSuffix(header.Filename, ".csv") {
			response.BadRequest(c, "only CSV files are supported")
			return
		}

		reader := csv.NewReader(file)
		reader.TrimLeadingSpace = true
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1

		headers, _, err := readCSVRecord(reader)
		if err != nil {
			response.BadRequest(c, "failed to read CSV headers")
			return
		}

		headerMap := make(map[string]int)
		for i, h := range headers {
			key := strings.ToLower(strings.TrimSpace(h))
			headerMap[key] = i
			switch key {
			case "category_name":
				if _, exists := headerMap["category"]; !exists {
					headerMap["category"] = i
				}
			case "uom_abbreviation":
				if _, exists := headerMap["uom"]; !exists {
					headerMap["uom"] = i
				}
			}
		}

		required := []string{"sku", "name"}
		for _, r := range required {
			if _, ok := headerMap[r]; !ok {
				response.BadRequest(c, fmt.Sprintf("missing required column: %s", r))
				return
			}
		}

		result := ImportResult{}

		tx, err := db.Beginx()
		if err != nil {
			response.InternalError(c, "failed to begin transaction")
			return
		}
		defer func() { _ = tx.Rollback() }()

		for {
			record, line, err := readCSVRecord(reader)
			if err == io.EOF {
				break
			}
			if err != nil {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "_", Error: fmt.Sprintf("parse error: %v", err)})
				result.Skipped++
				continue
			}

			if len(record) <= headerMap["sku"] || len(record) <= headerMap["name"] {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "sku", Error: "row has fewer columns than the header"})
				result.Skipped++
				continue
			}
			sku := strings.TrimSpace(record[headerMap["sku"]])
			name := strings.TrimSpace(record[headerMap["name"]])

			if sku == "" || name == "" {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "sku", Error: "sku and name are required"})
				result.Skipped++
				continue
			}

			var exists int
			if err := tx.Get(&exists, `SELECT COUNT(*) FROM products WHERE sku = $1 AND org_id = $2`, sku, orgID); err != nil {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "sku", Error: fmt.Sprintf("lookup error: %v", err)})
				result.Skipped++
				continue
			}
			if exists > 0 {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "sku", Error: fmt.Sprintf("duplicate SKU: %s", sku)})
				result.Skipped++
				continue
			}

			categoryName := getCSVField(record, headerMap, "category")
			uomAbbr := getCSVField(record, headerMap, "uom")
			itemType := getCSVField(record, headerMap, "item_type")

			var catID, uomID *string
			if categoryName != "" {
				var id string
				if err := tx.Get(&id, `SELECT id FROM product_categories WHERE name = $1 AND org_id = $2 LIMIT 1`, categoryName, orgID); err == nil {
					catID = &id
				}
			}
			if uomAbbr != "" {
				var id string
				if err := tx.Get(&id, `SELECT id FROM units_of_measure WHERE abbreviation = $1 LIMIT 1`, uomAbbr); err == nil {
					uomID = &id
				}
			}

			if itemType == "" {
				itemType = "consumable"
			}

			desc := getCSVField(record, headerMap, "description")
			brand := getCSVField(record, headerMap, "brand")
			manufacturer := getCSVField(record, headerMap, "manufacturer")

			_, err = tx.Exec(`
				INSERT INTO products (org_id, category_id, uom_id, sku, name, description, item_type, brand, manufacturer, status, created_by, updated_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active', $10, $11)`,
				orgID, catID, uomID, sku, name, desc, itemType, brand, manufacturer, userID, userID,
			)
			if err != nil {
				result.Errors = append(result.Errors, ImportError{Row: line, Field: "_", Error: fmt.Sprintf("insert error: %v", err)})
				result.Skipped++
				continue
			}

			result.Imported++
		}

		if len(result.Errors) > 0 {
			_ = tx.Rollback()
			response.Validation(c, result)
			return
		}

		if err := tx.Commit(); err != nil {
			response.InternalError(c, "failed to commit transaction")
			return
		}

		response.Created(c, result)
	}
}

func SearchProducts(db DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db = requestDB(c, db)
		orgID := c.GetString("org_id")
		q := strings.TrimSpace(c.Query("q"))
		params := pagination.FromQuery(c)

		if q == "" {
			response.BadRequest(c, "search query is required")
			return
		}

		var total int
		countQuery := `SELECT COUNT(*) FROM products WHERE org_id = $1 AND (name ILIKE $2 OR sku ILIKE $2 OR description ILIKE $2 OR brand ILIKE $2 OR manufacturer ILIKE $2)`
		likeQ := "%" + q + "%"
		if err := db.Get(&total, countQuery, orgID, likeQ); err != nil {
			response.InternalError(c, "failed to search products")
			return
		}

		type SearchProduct struct {
			ID       string `json:"id" db:"id"`
			SKU      string `json:"sku" db:"sku"`
			Name     string `json:"name" db:"name"`
			ItemType string `json:"item_type" db:"item_type"`
			Brand    string `json:"brand" db:"brand"`
			Status   string `json:"status" db:"status"`
		}

		var products []SearchProduct
		query := `SELECT id, sku, name, item_type, brand, status FROM products WHERE org_id = $1 AND (name ILIKE $2 OR sku ILIKE $2 OR description ILIKE $2 OR brand ILIKE $2 OR manufacturer ILIKE $2) ORDER BY name LIMIT $3 OFFSET $4`
		err := db.Select(&products, query, orgID, likeQ, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to search products")
			return
		}

		response.Paginated(c, products, total, params.Page, params.PageSize)
	}
}

func getCSVField(record []string, headerMap map[string]int, field string) string {
	if idx, ok := headerMap[field]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

// readCSVRecord returns the next data row from a CSV stream together with its
// physical line number. Rows are treated as comments only when the entire row
// is a single cell whose first non-whitespace character starts with "#", so
// templates may include instruction lines anywhere but a data row (e.g. a SKU
// that begins with "#") is never mistaken for a comment.
func readCSVRecord(reader *csv.Reader) ([]string, int, error) {
	for {
		record, err := reader.Read()
		if err != nil {
			if len(record) > 0 {
				line, _ := reader.FieldPos(0)
				return nil, line, err
			}
			return nil, 0, err
		}
		if len(record) == 1 && strings.HasPrefix(strings.TrimSpace(record[0]), "#") {
			continue
		}
		line, _ := reader.FieldPos(0)
		return record, line, nil
	}
}
