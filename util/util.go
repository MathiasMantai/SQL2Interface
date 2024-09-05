package util


func ValueInSlice(value interface{}, slice []interface{}) bool {
	for _, item := range slice {
        if item == value {
            return true
        }
    }
    return false
}

func StringToInterfaceSlice(arr []string) []interface{} {
	result := make([]interface{}, len(arr))
    for i, v := range arr {
        result[i] = v
    }
    return result
}