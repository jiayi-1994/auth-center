package util

func DeleteSliceElement[T string | int](slices []T, str T) []T {
	for i, slice := range slices {
		if slice == str {
			return append(slices[:i], slices[i+1:]...)
		}
	}
	return slices
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
