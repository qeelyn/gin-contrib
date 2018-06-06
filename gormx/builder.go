package gormx

import (
	"github.com/jinzhu/gorm"
	"github.com/qeelyn/go-common/protobuf/paginate"
	"strconv"
)

type Builder struct {
	db        *gorm.DB
	field     string
	order     string
	paginate  *paginate.Paginate
	Pageinfo  *paginate.PageInfo
	Total     int32
	needTotal bool
	isOffSet  bool
}

func NewBuild(db *gorm.DB) *Builder {
	return &Builder{
		db: db,
	}
}

func (t *Builder) Field(field string) *Builder {
	t.field = field
	return t
}

// only support string params,only parameter which start low case will set to where parameters
func (t *Builder) Where(where string, params map[string]string) *Builder {
	var p []string
	for k, v := range params {
		if f:=k[0]; f > 96 && f < 123 {
			p = append(p, v)
		}
	}
	t.db = t.db.Where(where, p)
	return t
}

func (t *Builder) PaginateCursor(p *paginate.Paginate) *Builder {
	if p == nil {
		return t
	}
	if p.First == 0 && p.After == "" {
		//向后分页 TODO
	}
	if p.Last == 0 && p.Before == "" {
		//向前分页 TODO
	}
	return t
}

func (t *Builder) PaginateOffSet(p *paginate.Paginate, needTotal bool) *Builder {
	t.paginate = p
	t.needTotal = needTotal
	t.isOffSet = true
	return t
}

func (t *Builder) parsePaginateOffSet() {
	if t.paginate == nil {
		return
	}
	var limit, page int
	if t.paginate.First != 0 && t.paginate.After != "" {
		if page, _ = strconv.Atoi(t.paginate.After); page == 0 {
			t.paginate.After = "1"
			page = 1
		}
		limit = int(t.paginate.First)
	}
	t.db = t.db.Offset((int32(page) - 1) * t.paginate.First).Limit(limit)
}

func (t *Builder) Order(order string) *Builder {
	t.order = order
	return t
}

// 返回即将执行的的Db
func (t *Builder) Prepare() *gorm.DB {
	if t.needTotal {
		t.db.Count(&t.Total)
	}
	t.parsePaginateOffSet()
	if t.field != "" {
		t.db = t.db.Select(t.field)
	}
	if t.order != "" {
		t.db = t.db.Order(t.order)
	}
	return t.db
}

func (t *Builder) GetPageInfo(count int)(*paginate.PageInfo,int32) {
	if t.paginate == nil {
		return nil,t.Total
	}
	if t.isOffSet {
		t.Pageinfo = &paginate.PageInfo{
			HasPreviousPage:      t.paginate.After != "1",
			HasNextPage: int32(count) == t.paginate.First,
		}
	}
	return t.Pageinfo,t.Total
}

func (t *Builder)GetDb() *gorm.DB {
	if t != nil {
		return t.db
	}
	return nil
}
