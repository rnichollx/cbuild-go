package cli

func PBool(b bool) *bool {
	return &b
}

func PString(s string) *string {
	return &s
}

func PInt(i int64) *int64 {
	return &i
}
