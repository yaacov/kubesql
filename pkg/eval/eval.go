/*
Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
and other contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>
*/

package eval

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Factory build an evaluation method for one item that returns a value using a key.
func Factory(item unstructured.Unstructured) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		return extractValue(item, key)
	}
}

// Factory2 build an evaluation method for two items that returns a value using a key.
func Factory2(item1, item2 unstructured.Unstructured, prefix1, prefix2 string) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		// Use item1 if has prefix1
		if strings.HasPrefix(key, prefix1+".") {
			return extractValue(item1, strings.TrimPrefix(key, prefix1+"."))
		}

		// Use item2 if has prefix2
		if strings.HasPrefix(key, prefix2+".") {
			return extractValue(item2, strings.TrimPrefix(key, prefix2+"."))
		}

		// Defatult to use item2
		return extractValue(item2, key)
	}
}

// extractValue extract a value from an item using a key.
func extractValue(item unstructured.Unstructured, key string) (interface{}, bool) {
	// Check for reserved words.
	switch key {
	case "name":
		return item.GetName(), true
	case "namespace":
		return item.GetNamespace(), true
	case "created":
		return item.GetCreationTimestamp().Time, true
	case "deleted":
		return item.GetDeletionTimestamp().Time, true
	}

	// Check for labels and annotations.
	if len(key) > 7 && key[:7] == "labels." {
		value, ok := item.GetLabels()[key[7:]]

		// Empty label represent the label is present
		if ok && len(value) == 0 {
			return true, true
		}

		// Missing value
		if !ok {
			return nil, true
		}

		v := stringValue(value)
		return v, true
	}

	if len(key) > 12 && key[:12] == "annotations." {
		value, ok := item.GetAnnotations()[key[12:]]

		// Empty annotations represent the annotations is present
		if ok && len(value) == 0 {
			return true, true
		}

		// Missing value
		if !ok {
			return nil, true
		}

		v := stringValue(value)
		return v, true
	}

	// Check for numbers, booleans, dates and strings.
	object, ok := getNestedObject(item.Object, key)
	if !ok {
		// Missing value is interpated as null value.
		return nil, true
	}

	switch object.(type) {
	case float64:
		return object.(float64), true
	case int64:
		return float64(object.(int64)), true
	case string:
		v := stringValue(object.(string))

		return v, true
	}

	// Missing value is interpated as null value.
	return nil, true
}

// Retrun a nested object using a key
func getNestedObject(object interface{}, key string) (interface{}, bool) {
	keys := strings.Split(key, ".")

	var objectList []interface{}
	var objectMap map[string]interface{}
	ok := true

	for _, key := range keys {
		if i, err := strconv.ParseUint(key, 10, 64); err == nil && i > 0 {
			if objectList, ok = object.([]interface{}); !ok {
				break
			}

			if ok = i <= uint64(len(objectList)); !ok {
				break
			}

			object = objectList[i-1]
		} else {
			if objectMap, ok = object.(map[string]interface{}); !ok {
				break
			}

			if object, ok = objectMap[key]; !ok {
				break
			}
		}
	}

	return object, ok
}

// Eval a string to a value, parse booleans, dates, SI values and numbers.
func stringValue(str string) interface{} {
	// Check for SI numbers
	multiplier := 0.0
	s := str

	// Remove SI `i` if exist
	// Note: we support "K", "M", "G" and "Ki", "Mi", "Gi" postfix
	if len(s) > 1 && s[len(s)-1:] == "i" {
		s = s[:len(s)-1]
	}

	// Check for SI postfix
	if len(s) > 1 {
		postfix := s[len(s)-1:]
		switch postfix {
		case "K":
			multiplier = 1024.0
		case "M":
			multiplier = math.Pow(1024, 2)
		case "G":
			multiplier = math.Pow(1024, 3)
		case "T":
			multiplier = math.Pow(1024, 4)
		case "P":
			multiplier = math.Pow(1024, 5)
		}

		if multiplier >= 1.0 {
			s = s[:len(s)-1]
		}

		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			newValue := float64(i) * multiplier

			return newValue
		}
	}

	// Check for false / true
	if str == "true" || str == "True" {
		return true
	}
	if str == "false" || str == "False" {
		return false
	}

	// Check for RFC3339 dates
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t
	}

	// Check for short hand dates
	if t, err := time.Parse("2014-06-23", str); err == nil {
		return t
	}

	// Default to string
	return str
}
