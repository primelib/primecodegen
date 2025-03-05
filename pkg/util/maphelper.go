package util

func GetMapString(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func GetMapBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

func GetMapSliceString(config map[string]interface{}, key string, defaultSlice []string) []string {
	if val, ok := config[key].([]string); ok {
		return val
	}
	return defaultSlice
}

func GetMapMap(config map[string]interface{}, key string) map[string]interface{} {
	if val, ok := config[key].(map[string]interface{}); ok {
		return val
	}
	return make(map[string]interface{})
}
