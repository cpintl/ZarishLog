package repository

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindByOrgID(orgID string) ([]model.Product, error) {
	var products []model.Product
	query := `SELECT * FROM products WHERE org_id = $1 AND status = 'active' ORDER BY name`
	err := r.db.Select(&products, query, orgID)
	return products, err
}

func (r *ProductRepository) FindByID(id, orgID string) (*model.Product, error) {
	var product model.Product
	query := `SELECT * FROM products WHERE id = $1 AND org_id = $2`
	err := r.db.Get(&product, query, id, orgID)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *model.Product) error {
	query := `
		INSERT INTO products (org_id, category_id, uom_id, sku, name, description, item_type,
			gtin, alternative_code, brand, manufacturer, is_batch_tracked, is_serial_tracked,
			is_expiry_tracked, is_hazardous, is_cold_chain, min_stock, max_stock, reorder_point,
			lead_time_days, unit_cost, status, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query,
		product.OrgID, product.CategoryID, product.UomID, product.SKU, product.Name,
		product.Description, product.ItemType, product.GTIN, product.AlternativeCode,
		product.Brand, product.Manufacturer, product.IsBatchTracked, product.IsSerialTracked,
		product.IsExpiryTracked, product.IsHazardous, product.IsColdChain,
		product.MinStock, product.MaxStock, product.ReorderPoint, product.LeadTimeDays,
		product.UnitCost, product.Status, product.CreatedBy, product.UpdatedBy,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}
