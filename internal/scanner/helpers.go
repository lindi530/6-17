package scanner

func maxTTL(left, right uint32) uint32 {
	if right > left {
		return right
	}
	return left
}

func appendUnique(values []string, next ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	for _, value := range next {
		if _, ok := seen[value]; value != "" && !ok {
			values = append(values, value)
			seen[value] = struct{}{}
		}
	}
	return values
}
