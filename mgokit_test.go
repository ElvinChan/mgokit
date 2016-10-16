package mgokit

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserInfo struct {
	EmpNo      string    `bson:"emp_no"`
	Name       string    `bson:"name"`
	Age        int       `bson:"age"`
	Department string    `bson:"department"`
	Registered bool      `bson:"registered"`
	CreateAt   time.Time `mgo:"created"`
	UpdateAt   time.Time `mgo:"updated"`
}

var c *mgo.Collection

func init() {
	session, err := mgo.Dial("139.196.228.246:27017")
	if err != nil {
		panic(err)
	}
	// defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c = session.DB("test").C("mgokit")
}

// TestInsert test C
func TestInsert(t *testing.T) {
	userInfo := UserInfo{}
	userInfo.EmpNo = "00000000"
	userInfo.Name = "John"
	userInfo.Age = 21
	userInfo.Department = "DEV"
	userInfo.Registered = true

	err := Insert(c, &userInfo)
	if err != nil {
		t.Error("Insert failed: " + err.Error())
	}

	// Check data
	userResult := UserInfo{}
	err = c.Find(bson.M{"name": "John"}).One(&userResult)
	if err != nil {
		t.Error("Insert failed: " + err.Error())
	}

	fmt.Println("TestInsert result: ", userResult)

	defer clearData()
}

// TestInsert test R
func TestFindOne(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	userInfo := UserInfo{}
	has, err := FindOne(c, func(c *mgo.Collection) *mgo.Query {
		return c.Find(bson.M{"department": "DEV"}).Skip(1).Sort("age")
	}, &userInfo)

	if err != nil {
		t.Error("FindOne failed: " + err.Error())
		return
	}
	if !has {
		t.Error("Not found")
		return
	}

	fmt.Println("TestFindOne result: ", userInfo)

	defer clearData()
}

func TestFindAll(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	userInfo := []UserInfo{}
	err := FindAll(c, func(c *mgo.Collection) *mgo.Query {
		return c.Find(bson.M{"department": "DEV"}).Skip(1).Sort("age")
	}, &userInfo)

	if err != nil {
		t.Error("FindAll failed: " + err.Error())
		return
	}

	fmt.Println("TestFindAll result: ", userInfo)

	defer clearData()
}

func TestUpdate(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	userInfo := UserInfo{}
	userInfo.Registered = false
	err := Update(c, bson.M{"emp_no": "00000002"}, &userInfo, "registered")
	if err != nil {
		t.Error("Update failed: " + err.Error())
		return
	}

	// Check data
	userResult := UserInfo{}
	err = c.Find(bson.M{"emp_no": "00000002"}).One(&userResult)
	if err != nil {
		t.Error("Update failed: " + err.Error())
	}

	fmt.Println("TestUpdate result: ", userResult)

	defer clearData()
}

func TestUpsert(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	userInfo := UserInfo{}
	userInfo.Registered = false
	err := Update(c, bson.M{"emp_no": "00000002"}, &userInfo, "registered")
	if err != nil {
		t.Error("Update failed: " + err.Error())
		return
	}

	// Check data
	userResult := UserInfo{}
	err = c.Find(bson.M{"emp_no": "00000002"}).One(&userResult)
	if err != nil {
		t.Error("Update failed: " + err.Error())
	}

	fmt.Println("TestUpsert result: ", userResult)

	defer clearData()
}

func TestDelete(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	err := Delete(c, bson.M{"emp_no": "00000001"})
	if err != nil {
		t.Error("Delete failed: " + err.Error())
		return
	}

	// Check data
	userResult := []UserInfo{}
	err = c.Find(nil).All(&userResult)
	if err != nil {
		t.Error("Delete failed: " + err.Error())
	}

	fmt.Println("TestDelete result: ", userResult)

	defer clearData()
}

func TestDeleteAll(t *testing.T) {
	// Prepare data
	if err := prepareData(); err != nil {
		t.Error("PrepareData failed: " + err.Error())
		return
	}

	_, err := DeleteAll(c, bson.M{"department": "DEV"})
	if err != nil {
		t.Error("Delete failed: " + err.Error())
		return
	}

	// Check data
	userResult := []UserInfo{}
	err = c.Find(nil).All(&userResult)
	if err != nil {
		t.Error("DeleteAll failed: " + err.Error())
	}

	fmt.Println("TestDeleteAll result: ", userResult)

	defer clearData()
}

func prepareData() error {
	userInfoA := UserInfo{}
	userInfoA.EmpNo = "00000001"
	userInfoA.Name = "John"
	userInfoA.Age = 21
	userInfoA.Department = "DEV"
	userInfoA.Registered = true
	err := prepareTime(&userInfoA)
	if err != nil {
		return err
	}

	userInfoB := UserInfo{}
	userInfoB.EmpNo = "00000002"
	userInfoB.Name = "Mary"
	userInfoB.Age = 23
	userInfoB.Department = "DEV"
	userInfoB.Registered = true
	err = prepareTime(&userInfoB)
	if err != nil {
		return err
	}

	userInfoC := UserInfo{}
	userInfoC.EmpNo = "00000003"
	userInfoC.Name = "Mark"
	userInfoC.Age = 22
	userInfoC.Department = "DEV"
	userInfoC.Registered = false
	err = prepareTime(&userInfoC)
	if err != nil {
		return err
	}

	userInfoD := UserInfo{}
	userInfoD.EmpNo = "00000004"
	userInfoD.Name = "Nancy"
	userInfoD.Age = 24
	userInfoD.Department = "HR"
	userInfoD.Registered = true
	err = prepareTime(&userInfoD)
	if err != nil {
		return err
	}

	return nil
}

func prepareTime(u *UserInfo) error {
	u.CreateAt = time.Now()
	u.UpdateAt = time.Now()
	return c.Insert(u)
}

func clearData() error {
	return c.DropCollection()
}
