package devcontainer

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

type StringArray []string

func (c *StringArray) UnmarshalJSON(data []byte) error {
	var jsonObj interface{}
	err := json.Unmarshal(data, &jsonObj)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal StringArray: %w", err)
	}

	switch obj := jsonObj.(type) {
	case string:
		*c = []string{obj}
		return nil
	case []interface{}:
		strings := make([]string, 0, len(obj))
		for _, v := range obj {
			if value, ok := v.(string); ok {
				strings = append(strings, value)
			} else {
				return fmt.Errorf("Unsupported type for StringArray: %T", v)
			}
		}

		*c = strings
		return nil
	}

	return fmt.Errorf("Unsupported type for StringArray: %T", jsonObj)
}

type LifecycleCommand map[string][]string

func (c *LifecycleCommand) UnmarshalJSON(data []byte) error {
	var jsonObj interface{}
	err := json.Unmarshal(data, &jsonObj)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal LifecycleCommand: %w", err)
	}

	switch obj := jsonObj.(type) {
	case string:
		*c = LifecycleCommand{
			"": []string{DefaultShell, "-c", obj},
		}
		return nil
	case []interface{}:
		commands := make(map[string][]string, len(obj))

		for _, v := range obj {
			if value, ok := v.(string); ok {
				commands[""] = append(commands[""], value)
			} else {
				return fmt.Errorf("Unsupported type for LifecycleCommand: %T", v)
			}
		}

		*c = commands
		return nil
	case map[string]interface{}:
		commands := make(map[string][]string, len(obj))
		for key, value := range obj {
			switch obj := value.(type) {
			case string:
				commands[key] = []string{DefaultShell, "-c", obj}
			case []interface{}:
				strings := make([]string, 0, len(obj))
				for _, v := range obj {
					if value, ok := v.(string); ok {
						strings = append(strings, value)
					} else {
						return fmt.Errorf("Unsupported type for LifecycleCommand: %T", v)
					}
				}

				commands[key] = strings
			default:
				return fmt.Errorf("Unsupported type for LifecycleCommand: %T", value)
			}
		}

		*c = commands
		return nil
	}

	return fmt.Errorf("Unsupported type for LifecycleCommand: %T", jsonObj)

}

func (c *LifecycleCommand) Equals(other *LifecycleCommand) bool {
	return maps.EqualFunc(*c, *other, slices.Equal)
}
