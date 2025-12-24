package handler

func extractName(names map[string]string, preferredLang string) string {
	if name, ok := names[preferredLang]; ok && name != "" {
		return name
	}
	for _, name := range names {
		if name != "" {
			return name
		}
	}
	return ""
}
