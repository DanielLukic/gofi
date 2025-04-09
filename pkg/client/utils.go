package client

// containsStr checks if a string is in a slice
// Args:
//
//	slice: Slice to search
//	str: String to find
//
// Returns:
//
//	bool: True if string is found
func containsStr(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
