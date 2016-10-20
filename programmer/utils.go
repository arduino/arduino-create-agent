package programmer

// differ returns the first item that differ between the two input slices
func differ(slice1 []string, slice2 []string) string {
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			return mKey
		}
	}

	return ""
}
