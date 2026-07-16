package handler

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullIfEmptyPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
