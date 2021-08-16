package emailalerts

func removeBlank(m map[string]string) {
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
}

func removeFalse(m map[string]bool) {
	for k, v := range m {
		if !v {
			delete(m, k)
		}
	}
}
