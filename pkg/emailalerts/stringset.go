package emailalerts

func stringsToSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func setToStrings(set map[string]bool) []string {
	ss := make([]string, 0, len(set))
	for s := range set {
		ss = append(ss, s)
	}
	return ss
}

func symDiff(oldss, newss []string) (added, removed []string) {
	oldSet := stringsToSet(oldss)
	addedSet := make(map[string]bool, len(newss))
	for _, s := range newss {
		if !oldSet[s] {
			addedSet[s] = true
		}
	}
	added = setToStrings(addedSet)

	newSet := stringsToSet(newss)
	removedSet := make(map[string]bool, len(oldss))
	for _, s := range oldss {
		if !newSet[s] {
			removedSet[s] = true
		}
	}
	removed = setToStrings(removedSet)
	return
}

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
