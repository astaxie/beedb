package main

import (
	"fmt"
	"github.com/astaxie/beedb"
	_ "github.com/bmizerany/pq"
	"time"
	"database/sql"
)

var orm beedb.Model

type Userinfo struct {
	Uid        int `PK`
	Username   string
	Departname string
	Created    time.Time
}

func main() {
	db, err := sql.Open("postgres", "user=asta dbname=123456 sslmode=verify-full")
	if err != nil {
		panic(err)
	}
	orm = beedb.New(db)
	insert()
	// insertsql()
	// a := selectone()
	// fmt.Println(a)
	// b := selectall()
	// fmt.Println(b)
	// update()
	// updatesql()
	// findmap()
	// groupby()
	// jointable()
	//delete()
	//deleteall()
	//deletesql()
}

func insert() {
	//save data
	var saveone Userinfo
	saveone.Username = "Test Add User"
	saveone.Departname = "Test Add Departname"
	saveone.Created = time.Now()
	orm.Save(&saveone)
	fmt.Println(saveone)
}
