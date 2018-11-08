/* Collection of helper for slice of string */
package strslice

func Contain(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}

func Equal(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, _ := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true

}
