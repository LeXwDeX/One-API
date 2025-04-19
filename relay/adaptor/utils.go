package adaptor

import (
	"reflect"
)

// StructToMap 将结构体转为 map[string]interface{}，仅处理导出字段和 json 标签
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // 非导出字段跳过
			continue
		}
		jsonTag := field.Tag.Get("json")
		name := field.Name
		if jsonTag != "" && jsonTag != "-" {
			name = jsonTag
			// 处理 omitempty
			if idx := len(name); idx > 0 {
				if idx := len(name); idx > 0 {
					if commaIdx := indexComma(name); commaIdx > 0 {
						name = name[:commaIdx]
					}
				}
			}
		}
		result[name] = val.Field(i).Interface()
	}
	return result, nil
}

// indexComma 返回字符串中第一个逗号的位置
func indexComma(s string) int {
	for i, c := range s {
		if c == ',' {
			return i
		}
	}
	return -1
}
