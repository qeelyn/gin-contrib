package gormx_test

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/qeelyn/gin-contrib/gormx"
	"fmt"
)

var(
	Db *gorm.DB
)

func init()  {
	Db,_ = gorm.Open("mysql","root:@tcp(localhost:3306)/test")
}

func TestBuilder_Where(t *testing.T) {
	bl := gormx.NewBuild(Db)
	wstr := "id = ? and date between ? and ?"
	wps := map[string]string{
		"id": "1",
		"sd": "2017-01-01",
		"ed": "2017-02-02",
	}
	bl.Where(wstr,wps)
	fmt.Println(bl.Prepare().SubQuery())
}