package handler

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
