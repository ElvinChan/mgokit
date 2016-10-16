package mgokit

import (
	"reflect"
	"strings"
	"time"
	"unicode"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ProcessType specify type for CURD
type ProcessType int

const (
	insertType ProcessType = iota
	updateType
	upsertType
	deleteType
)

// FindOne find single result from DB by query s
func FindOne(c *mgo.Collection, s func(*mgo.Collection) *mgo.Query, m interface{}) (bool, error) {
	err := s(c).One(m)
	if err != nil {
		if err.Error() == "not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// FindAll find all result from DB by query s
func FindAll(c *mgo.Collection, s func(*mgo.Collection) *mgo.Query, m interface{}) error {
	err := s(c).All(m)
	if err != nil {
		if err.Error() == "not found" {
			return nil
		}
		return err
	}
	return nil
}

// Insert object docs to DB
func Insert(c *mgo.Collection, docs interface{}) error {
	inserter := getBson(docs, insertType)
	return c.Insert(inserter)
}

// Update object docs in DB with selector
func Update(c *mgo.Collection, selector bson.M, docs interface{}, cols ...string) error {
	updater := getBson(docs, updateType, cols...)
	return c.Update(selector, bson.M{"$set": updater})
}

// Upsert update object docs in DB with selector if exist, otherwise insert
func Upsert(c *mgo.Collection, selector bson.M, docs interface{}, cols ...string) error {
	updater := getBson(docs, updateType, cols...)
	info, err := c.Upsert(selector, bson.M{"$set": updater})
	if err != nil {
		return err
	}

	if info.Updated == 0 {
		upserter := getBson(docs, upsertType)
		return c.Update(selector, bson.M{"$set": upserter})
	}

	return nil
}

// Delete object in DB with selector
func Delete(c *mgo.Collection, selector bson.M) error {
	return c.Remove(selector)
}

// DeleteAll object in DB with selector
func DeleteAll(c *mgo.Collection, selector bson.M) (*mgo.ChangeInfo, error) {
	return c.RemoveAll(selector)
}

func getBson(m interface{}, process ProcessType, cols ...string) bson.M {
	object := reflect.ValueOf(m)
	ref := object.Elem()
	typeObject := object.Elem().Type()

	result := make(map[string]interface{}, 0)

	for i := 0; i < ref.NumField(); i++ {
		// 1.Key
		key := typeObject.Field(i).Tag.Get("bson")

		if key == "" {
			// No bsontag
			field := typeObject.Field(i).Name
			// Handle field
			key = processField(field)
		} else if key == "-" {
			continue
		}

		// 2.Value
		value := ref.Field(i).Interface()

		// 3.Special mark
		otherTag := typeObject.Field(i).Tag.Get("mgo")
		otherTag = strings.ToUpper(otherTag)

		if process == upsertType && otherTag != "CREATED" {
			continue
		}

		switch otherTag {
		case "CREATED":
			if process == updateType {
				continue
			}
			value = time.Now()
		case "UPDATED":
			value = time.Now()
		}

		// Skip other cols
		if len(cols) > 0 {
			flag := false
			for _, item := range cols {
				if item == key {
					flag = true
					break
				}
			}
			if !flag {
				continue
			}
		}

		result[key] = value
	}

	return result
}

func processField(field string) string {
	r := []rune(field)
	result := ""
	for i, item := range r {
		if unicode.IsLetter(item) && unicode.IsUpper(item) {
			if i > 0 {
				result += "_"
			}
			result += string(unicode.ToLower(item))
		} else {
			result += string(item)
		}
	}
	return result
}
