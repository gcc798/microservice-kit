package types

import (
	"encoding/json"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (r CommonResp) MarshalJSON() ([]byte, error) {
	type commonResp struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data,omitempty"`
	}
	return json.Marshal(commonResp{
		Code: r.Code,
		Msg:  r.Msg,
		Data: normalizeResponseValue(r.Data),
	})
}

func normalizeResponseValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	if msg, ok := value.(proto.Message); ok {
		return normalizeProtoMessage(msg)
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return value
	}
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return normalizeResponseValue(rv.Elem().Interface())
	case reflect.Slice, reflect.Array:
		out := make([]interface{}, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, normalizeResponseValue(rv.Index(i).Interface()))
		}
		return out
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return value
		}
		out := make(map[string]interface{}, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out[iter.Key().String()] = normalizeResponseValue(iter.Value().Interface())
		}
		normalizeNativeAliases(out)
		return out
	default:
		return value
	}
}

func normalizeProtoMessage(msg proto.Message) interface{} {
	raw, err := protojson.MarshalOptions{
		UseProtoNames:   false,
		EmitUnpopulated: true,
	}.Marshal(msg)
	if err != nil {
		return msg
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return msg
	}
	return normalizeResponseValue(out)
}

func normalizeNativeAliases(item map[string]interface{}) {
	if value, ok := item["dataJson"]; ok {
		item["data"] = parseJSONString(value)
		delete(item, "dataJson")
	}
	if value, ok := item["configJson"]; ok {
		item["config"] = parseJSONString(value)
		delete(item, "configJson")
	}
	if value, ok := item["metadataJson"]; ok {
		item["metadata"] = parseJSONString(value)
		delete(item, "metadataJson")
	}
	if value, ok := item["page"]; ok {
		flattenPageInfo(item, value)
		delete(item, "page")
	}
	if value, ok := item["createdAt"]; ok {
		if _, exists := item["createdTime"]; !exists {
			item["createdTime"] = value
		}
	}
	if value, ok := item["updatedAt"]; ok {
		if _, exists := item["updatedTime"]; !exists {
			item["updatedTime"] = value
		}
	}
	if value, ok := item["createTime"]; ok {
		if _, exists := item["createdTime"]; !exists {
			item["createdTime"] = value
		}
	}
	if value, ok := item["updateTime"]; ok {
		if _, exists := item["updatedTime"]; !exists {
			item["updatedTime"] = value
		}
	}
}

func flattenPageInfo(item map[string]interface{}, value interface{}) {
	page, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	for _, key := range []string{"total", "size", "current", "pages"} {
		if _, exists := item[key]; exists {
			continue
		}
		if pageValue, ok := page[key]; ok {
			item[key] = pageValue
		}
	}
}

func parseJSONString(value interface{}) interface{} {
	text, ok := value.(string)
	if !ok {
		return value
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return value
	}
	return out
}
