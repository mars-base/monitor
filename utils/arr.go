package utils

import (
	"reflect"
)

func AppendInt(slice []int, item int) []int {
	return append(slice, item)
}

func AppendString(slice []string, item string) []string {
	return append(slice, item)
}

func AppendByte(slice []byte, item byte) []byte {
	return append(slice, item)
}

func Append(slice interface{}, item interface{}) interface{} {
	v := reflect.ValueOf(slice)
	v = reflect.Append(v, reflect.ValueOf(item))
	return v.Interface()
}

func DelStringArrayItem(slice []string, item string) []string {
	l := make([]string, 0)
	for _, v := range slice {
		if v != item {
			l = append(l, v)
		}
	}
	return l
}

func ReplaceStringArrayItem(slice []string, find string, item string) []string {
	l := make([]string, 0)
	for _, v := range slice {
		if v == find {
			l = append(l, item)
		} else {
			l = append(l, v)
		}
	}
	return l
}

func DelIntArrayItem(slice []int, item int) []int {
	l := make([]int, 0)
	for _, v := range slice {
		if v != item {
			l = append(l, v)
		}
	}
	return l
}

func ReplaceIntArrayItem(slice []int, find int, item int) []int {
	l := make([]int, 0)
	for _, v := range slice {
		if v == find {
			l = append(l, item)
		} else {
			l = append(l, v)
		}
	}
	return l
}
