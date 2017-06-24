package config

func (c Config) CfgString(name string) string {
	c.VarsMutex.RLock()
	defer c.VarsMutex.RUnlock()

	v, exists := c.Vars[name]
	if !exists {
		return ""
	}

	if v, ok := v.Interface().(string); ok {
		return v
	}
	return ""
}

func (c Config) CfgBool(name string) bool {
	c.VarsMutex.RLock()
	defer c.VarsMutex.RUnlock()

	v, exists := c.Vars[name]
	if !exists {
		return false
	}

	if v, ok := v.Interface().(bool); ok {
		return v
	}
	return false
}

func (c Config) CfgFloat(name string) float64 {
	c.VarsMutex.RLock()
	defer c.VarsMutex.RUnlock()

	v, exists := c.Vars[name]
	if !exists {
		return 0
	}

	if v, ok := v.Interface().(float64); ok {
		return v
	}
	return 0
}
