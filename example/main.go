package main

import (
	"fmt"
	"github.com/astaxie/beedb"
	_ "github.com/ziutek/mymysql/godrv"
	"time"
	"database/sql"
)

/*
CREATE TABLE `userinfo` (
	`uid` INT(10) NULL AUTO_INCREMENT,
	`username` VARCHAR(64) NULL,
	`departname` VARCHAR(64) NULL,
	`created` DATE NULL,
	PRIMARY KEY (`id`)
)
CREATE TABLE `userdeatail` (
	`uid` INT(10) NULL,
	`intro` TEXT NULL,
	`profile` TEXT NULL,
	PRIMARY KEY (`uid`)
)
*/

type Userinfo struct {
	Uid		int	`PK`
	Username	string
	Departname	string
	Created		time.Time
}

func main() {
	db, err := sql.Open("mymysql", "test/xiemengjun/123456")
	if err != nil {
		panic(err)
	}
	orm := beedb.New(db)

	//Original SQL Join Table
	a, _ := orm.SetTable("userinfo").Join("LEFT", "userdeatail", "userinfo.uid=userdeatail.uid").Where("userinfo.uid=?", 1).Select("userinfo.uid,userinfo.username,userdeatail.profile").FindMap()
	fmt.Println(a)

	//Original SQL Group By 
	b, _ := orm.SetTable("userinfo").GroupBy("username").Having("username='astaxie'").FindMap()
	fmt.Println(b)

	//Original SQL Backinfo resultsSlice []map[string][]byte 
	//default PrimaryKey id
	c, _ := orm.SetTable("userinfo").SetPK("uid").Where(2).Select("uid,username").FindMap()
	fmt.Println(c)

	//original SQL update 
	t := make(map[string]interface{})
	var j interface{}
	j = "astaxie"
	t["username"] = j
	//update one
	orm.SetTable("userinfo").SetPK("uid").Where(2).Update(t)
	//update batch
	orm.SetTable("userinfo").Where("uid>?", 3).Update(t)

	// add one
	add := make(map[string]interface{})
	j = "astaxie"
	add["username"] = j
	j = "cloud develop"
	add["departname"] = j
	j = "2012-12-02"
	add["created"] = j

	orm.SetTable("userinfo").Insert(add)

	//original SQL delete
	orm.SetTable("userinfo").Where("uid>?", 3).DelectRow()

	//get all data
	var alluser []Userinfo
	orm.Limit(10).Where("uid>?", 1).FindAll(&alluser)
	fmt.Println(alluser)

	//get one info
	var one Userinfo
	orm.Where("uid=?", 27).Find(&one)
	fmt.Println(one)

	//save data
	var saveone Userinfo
	saveone.Username = "Test Add User"
	saveone.Departname = "Test Add Departname"
	saveone.Created = time.Now()
	orm.Save(&saveone)
	fmt.Println(saveone)

	// //update data
	saveone.Username = "Update Username"
	saveone.Departname = "Update Departname"
	saveone.Created = time.Now()
	orm.Save(&saveone)
	fmt.Println(saveone)

	// // //delete one data
	orm.Delete(&saveone)

	// //delete all data
	orm.DeleteAll(&alluser)
}
