package src

// ValueInSlice checks if a given value exists in a slice of empty interfaces.//+
// //+
// The function iterates over each element in the provided slice and compares it with the given value.//+
// If a match is found, the function returns true. Otherwise, it returns false.//+
// //+
// Parameters://+
// - value: The value to be checked for existence in the slice.//+
// - slice: The slice of empty interfaces to search for the given value.//+
// //+
// Return://+
// - A boolean value indicating whether the given value exists in the slice.//+
func ValueInSlice(value interface{}, slice []interface{}) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// StringToInterfaceSlice converts a slice of strings to a slice of empty interfaces.//+
// //+
// The function takes a slice of strings as input and returns a new slice of empty interfaces.//+
// Each element in the input slice is assigned to the corresponding element in the output slice.//+
// //+
// Parameters://+
// - arr: A slice of strings to be converted.//+
// //+
// Return://+
// - A new slice of empty interfaces, where each element is the corresponding element from the input slice.//+
func StringToInterfaceSlice(arr []string) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[i] = v
	}
	return result
}

// AddTabToSlice adds a tab character ("\t") to the beginning of each string in the given slice.//+
// //+
// The function takes a pointer to a slice of strings as input. It iterates over each string in the slice//+
// and prepends a tab character to it. The modified slice is then returned.//+
// //+
// Parameters://+
// - values: A pointer to a slice of strings. The function modifies this slice in-place.//+
// //+
// Return://+
// The function does not return a value, but it modifies the slice pointed to by the 'values' parameter.//+
func AddTabToSlice(values *[]string) {
	for i, item := range *values {
		(*values)[i] = "\t" + item
	}
}
