package util

func MergeMaps(target, source map[string]interface{}) {
	for key, sourceValue := range source {
		if targetValue, exists := target[key]; exists {
			if targetMap, isMap := targetValue.(map[string]interface{}); isMap {
				if sourceMap, isMap := sourceValue.(map[string]interface{}); isMap {
					MergeMaps(targetMap, sourceMap)
				} else {
					target[key] = sourceValue
				}
			} else {
				target[key] = sourceValue
			}
		} else {
			target[key] = sourceValue
		}
	}
}
