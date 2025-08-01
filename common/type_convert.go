package common

import (
	"encoding/json"
	"github.com/spf13/cast"
	"strings"
	"time"
)

func NewTime(v time.Time) *time.Time {
	return &v
}

func NewString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func ToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func ToInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func NewInt(v int) *int {
	return &v
}

func NewInt32(v int32) *int32 {
	return &v
}

func NewInt64(v int64) *int64 {
	return &v
}

func NewFloat64(v float64) *float64 {
	return &v
}

func ToJson(v interface{}) string {
	j, _ := json.Marshal(v)
	return string(j)
}

func ToBytes(v interface{}) []byte {
	j, _ := json.Marshal(v)
	return j
}

func FormatTime(v time.Time) string {
	return v.Format("2006-01-02 15:04:05")
}

func FormatTime2(v *time.Time) string {
	if v == nil {
		return ""
	}
	return v.Format("2006-01-02 15:04:05")
}

func ToDateTime(date string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local)
	return t
}

func Split(str string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}

func ToList(m map[string]int) []string {
	v := make([]string, 0)
	for key, _ := range m {
		v = append(v, key)
	}
	return v
}

func ParseArray(v string) []int32 {
	arr := strings.Split(v, ",")
	ret := make([]int32, 0)
	for _, id := range arr {
		ret = append(ret, cast.ToInt32(id))
	}
	return ret
}
