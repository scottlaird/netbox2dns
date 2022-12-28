package netbox

func String(s *string) string {
	if s != nil {
		return *s
	} else {
		return ""
	}
}

func Bool(b *bool) bool {
	if b != nil {
		return *b
	} else {
		return false
	}
}

func Int64(i *int64) int64 {
	if i != nil {
		return *i
	} else {
		return 0
	}
}
