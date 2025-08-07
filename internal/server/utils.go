package server

func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
