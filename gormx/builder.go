package gormx

import (
	"github.com/jinzhu/gorm"
	"github.com/qeelyn/go-common/protobuf/paginate"
	"strconv"
)

type Builder struct {
	db         *gorm.DB
	countDb    *gorm.DB
	field      string
	order      string
	pagination *paginate.Pagination
	Pageinfo   *paginate.PageInfo
	Total      int32
	needTotal  bool
	isOffSet   bool
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
		if f := k[0]; f > 96 && f < 123 {
			p = append(p, v)
		}
	}
	t.db = t.db.Where(where, p)
	return t
}

func (t *Builder) PaginateCursor(p *paginate.Pagination) *Builder {
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

func (t *Builder) PaginateOffSet(p *paginate.Pagination, needTotal bool) *Builder {
	t.pagination = p
	t.needTotal = needTotal
	t.isOffSet = true
	return t
}

func (t *Builder) parsePaginateOffSet() {
	if t.pagination == nil {
		return
	}
	var limit, page int
	if t.pagination.First != 0 && t.pagination.After != "" {
		if page, _ = strconv.Atoi(t.pagination.After); page == 0 {
			t.pagination.After = "1"
			page = 1
		}
		limit = int(t.pagination.First)
	}
	t.db = t.db.Offset((int32(page) - 1) * t.pagination.First).Limit(limit)
}

func (t *Builder) Order(order string) *Builder {
	t.order = order
	return t
}

// 返回即将执行的的Db
func (t *Builder) Prepare() *gorm.DB {
	if t.needTotal {
		t.countDb = t.db
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

func (t *Builder) Find(out interface{}) *gorm.DB {
	if t.needTotal && t.countDb != nil {
		if db := t.countDb.Model(out).Count(&t.Total);db.Error!= nil {
			return db
		}
	}
	return t.db.Find(out)
}

func (t *Builder) GetPageInfo(count int) (*paginate.PageInfo, int32) {
	if t.pagination == nil {
		return nil, t.Total
	}
	if t.isOffSet {
		t.Pageinfo = &paginate.PageInfo{
			HasPreviousPage: t.pagination.After != "1",
			HasNextPage:     int32(count) == t.pagination.First,
		}
	}
	return t.Pageinfo, t.Total
}

func (t *Builder) GetDb() *gorm.DB {
	if t != nil {
		return t.db
	}
	return nil
}

func (t *Builder) SetDb(db *gorm.DB) {
	t.db = db
}
