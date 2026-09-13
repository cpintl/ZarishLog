package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func init() {
	validate = validator.New()
	register := func(tag string, fn validator.Func) {
		if err := validate.RegisterValidation(tag, fn); err != nil {
			panic(fmt.Sprintf("validator: failed to register %q: %v", tag, err))
		}
	}
	register("uuid7", validateUUIDv7)
	register("date", validateDate)
	register("item_type", validateItemType)
	register("movement_type", validateMovementType)
	register("wh_type", validateWarehouseType)
	register("loc_type", validateLocationType)
	register("uom_category", validateUoMCategory)

	validate.RegisterAlias("opt_uuid7", "omitempty,uuid7")
	validate.RegisterAlias("opt_date", "omitempty,date")
}

func validateUUIDv7(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, fl.Field().String())
	return matched
}

func validateDate(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, fl.Field().String())
	return matched
}

func validateItemType(fl validator.FieldLevel) bool {
	valid := map[string]bool{
		"drug": true, "medical_supply": true, "equipment": true,
		"nutrition": true, "wash": true, "office": true, "lab": true,
		"cold_chain": true, "shelter": true, "comms": true, "vehicle": true,
		"ppe": true, "cleaning": true, "asset": true, "consumable": true,
	}
	return valid[fl.Field().String()]
}

func validateMovementType(fl validator.FieldLevel) bool {
	valid := map[string]bool{"receipt": true, "issue": true, "transfer": true, "adjustment": true, "return": true, "disposal": true}
	return valid[fl.Field().String()]
}

func validateWarehouseType(fl validator.FieldLevel) bool {
	valid := map[string]bool{"central": true, "sub_warehouse": true, "transit": true, "quarantine": true}
	return valid[fl.Field().String()]
}

func validateLocationType(fl validator.FieldLevel) bool {
	valid := map[string]bool{"zone": true, "aisle": true, "rack": true, "bin": true, "shelf": true, "area": true}
	return valid[fl.Field().String()]
}

func validateUoMCategory(fl validator.FieldLevel) bool {
	valid := map[string]bool{
		"count": true, "weight": true, "volume": true, "length": true,
		"dosage": true, "packaging": true,
	}
	return valid[fl.Field().String()]
}

func BindAndValidate(c *gin.Context, obj interface{}) []FieldError {
	if err := c.ShouldBindJSON(obj); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(verrs)
		}
		return []FieldError{{Field: "_", Message: err.Error()}}
	}
	if err := validate.Struct(obj); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(verrs)
		}
		return []FieldError{{Field: "_", Message: err.Error()}}
	}
	return nil
}

func ValidateStruct(obj interface{}) []FieldError {
	if err := validate.Struct(obj); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(verrs)
		}
		return []FieldError{{Field: "_", Message: err.Error()}}
	}
	return nil
}

func formatValidationErrors(errs validator.ValidationErrors) []FieldError {
	fields := make([]FieldError, 0, len(errs))
	for _, e := range errs {
		field := strings.ToLower(e.Field())
		msg := messageForTag(e.Tag(), e.Param(), field)
		fields = append(fields, FieldError{Field: field, Message: msg})
	}
	return fields
}

func messageForTag(tag, param, field string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "uuid7":
		return field + " must be a valid UUIDv7"
	case "date":
		return field + " must be a valid date (YYYY-MM-DD)"
	case "item_type":
		return field + " is not a valid item type"
	case "movement_type":
		return field + " is not a valid movement type"
	case "wh_type":
		return field + " is not a valid warehouse type"
	case "loc_type":
		return field + " is not a valid location type (zone, aisle, rack, bin, shelf, area)"
	case "uom_category":
		return field + " is not a valid UoM category"
	case "min":
		return field + " must be at least " + param
	case "max":
		return field + " must be at most " + param
	case "gt":
		return field + " must be greater than " + param
	case "gte":
		return field + " must be at least " + param
	case "lt":
		return field + " must be less than " + param
	case "lte":
		return field + " must be at most " + param
	case "oneof":
		return field + " must be one of " + param
	default:
		return field + " failed " + tag + " validation"
	}
}
