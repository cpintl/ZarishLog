package handler

import (
	"github.com/cpintl/zarishlog-api/internal/model"
	"github.com/cpintl/zarishlog-api/internal/pagination"
	"github.com/cpintl/zarishlog-api/internal/response"
	"github.com/cpintl/zarishlog-api/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func ListLocations(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		whID := c.Param("warehouse_id")
		params := pagination.FromQuery(c)

		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM locations WHERE warehouse_id = $1`, whID); err != nil {
			response.InternalError(c, "failed to count locations")
			return
		}

		var locs []model.Location
		err := db.Select(&locs, `SELECT * FROM locations WHERE warehouse_id = $1 ORDER BY code LIMIT $2 OFFSET $3`, whID, params.Limit(), params.Offset())
		if err != nil {
			response.InternalError(c, "failed to fetch locations")
			return
		}

		response.Paginated(c, locs, total, params.Page, params.PageSize)
	}
}

func GetLocation(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var loc model.Location
		err := db.Get(&loc, `SELECT * FROM locations WHERE id = $1`, id)
		if err != nil {
			response.NotFound(c, "location not found")
			return
		}
		response.OK(c, loc)
	}
}

func CreateLocation(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loc model.Location
		if errs := validator.BindAndValidate(c, &loc); errs != nil {
			response.Validation(c, errs)
			return
		}

		err := db.QueryRow(`
			INSERT INTO locations (warehouse_id, parent_id, code, name, type, is_cold_chain, is_hazardous, is_secure, max_capacity, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`,
			loc.WarehouseID, loc.ParentID, loc.Code, loc.Name, loc.Type,
			loc.IsColdChain, loc.IsHazardous, loc.IsSecure, loc.MaxCapacity, loc.IsActive,
		).Scan(&loc.ID, &loc.CreatedAt, &loc.UpdatedAt)

		if err != nil {
			response.InternalError(c, "failed to create location")
			return
		}
		response.Created(c, loc)
	}
}

func UpdateLocation(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var loc model.Location
		if errs := validator.BindAndValidate(c, &loc); errs != nil {
			response.Validation(c, errs)
			return
		}

		_, err := db.Exec(`
			UPDATE locations SET parent_id=$1, code=$2, name=$3, type=$4,
				is_cold_chain=$5, is_hazardous=$6, is_secure=$7, max_capacity=$8, is_active=$9
			WHERE id=$10`,
			loc.ParentID, loc.Code, loc.Name, loc.Type,
			loc.IsColdChain, loc.IsHazardous, loc.IsSecure, loc.MaxCapacity, loc.IsActive, id,
		)
		if err != nil {
			response.InternalError(c, "failed to update location")
			return
		}
		response.OK(c, gin.H{"message": "location updated"})
	}
}

func DeleteLocation(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec(`DELETE FROM locations WHERE id = $1`, id)
		if err != nil {
			response.InternalError(c, "failed to delete location")
			return
		}
		response.OK(c, gin.H{"message": "location deleted"})
	}
}

func GetLocationConstraints(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var con model.LocationConstraint
		err := db.Get(&con, `SELECT * FROM location_constraints WHERE location_id = $1`, id)
		if err != nil {
			response.NotFound(c, "location constraints not found")
			return
		}
		response.OK(c, con)
	}
}

func UpsertLocationConstraints(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var con model.LocationConstraint
		con.LocationID = id
		if errs := validator.BindAndValidate(c, &con); errs != nil {
			response.Validation(c, errs)
			return
		}

		_, err := db.Exec(`
			INSERT INTO location_constraints (location_id, min_temperature, max_temperature, min_humidity, max_humidity,
				is_hazardous_allowed, is_food_grade, is_pharma_grade, max_weight_capacity)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (location_id) DO UPDATE SET
				min_temperature=EXCLUDED.min_temperature,
				max_temperature=EXCLUDED.max_temperature,
				min_humidity=EXCLUDED.min_humidity,
				max_humidity=EXCLUDED.max_humidity,
				is_hazardous_allowed=EXCLUDED.is_hazardous_allowed,
				is_food_grade=EXCLUDED.is_food_grade,
				is_pharma_grade=EXCLUDED.is_pharma_grade,
				max_weight_capacity=EXCLUDED.max_weight_capacity,
				updated_at=now()`,
			con.LocationID, con.MinTemperature, con.MaxTemperature,
			con.MinHumidity, con.MaxHumidity,
			con.IsHazardousAllowed, con.IsFoodGrade, con.IsPharmaGrade,
			con.MaxWeightCapacity,
		)
		if err != nil {
			response.InternalError(c, "failed to upsert location constraints")
			return
		}
		response.OK(c, con)
	}
}

func ListLocationTree(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		whID := c.Param("warehouse_id")

		var locs []model.Location
		err := db.Select(&locs, `SELECT * FROM locations WHERE warehouse_id = $1 ORDER BY type, code`, whID)
		if err != nil {
			response.InternalError(c, "failed to fetch location tree")
			return
		}

		tree := buildLocationTree(locs)
		response.OK(c, tree)
	}
}

type LocationNode struct {
	model.Location
	Children []LocationNode `json:"children,omitempty"`
}

func buildLocationTree(locs []model.Location) []LocationNode {
	byParent := make(map[string][]model.Location)
	for _, l := range locs {
		key := ""
		if l.ParentID != nil {
			key = *l.ParentID
		}
		byParent[key] = append(byParent[key], l)
	}

	var build func(parentID string) []LocationNode
	build = func(parentID string) []LocationNode {
		children := byParent[parentID]
		if len(children) == 0 {
			return nil
		}
		nodes := make([]LocationNode, 0, len(children))
		for _, c := range children {
			nodes = append(nodes, LocationNode{
				Location: c,
				Children: build(c.ID),
			})
		}
		return nodes
	}

	return build("")
}
