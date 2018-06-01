package gorm

import (
	"github.com/jinzhu/gorm"
	"github.com/qeelyn/go-common/protobuf/paginate"
	"strconv"
)

type Builder struct {
	db    *gorm.DB
	field string
	order string
	paginate *paginate.Paginate
	Pageinfo paginate.PageInfo
	Total int32
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

// only support string params
func (t *Builder) Where(where string,params map[string]string) *Builder {
	var p []string
	for _,v := range params {
		p = append(p,v)
	}
	t.db = t.db.Where(where,p)
	return t
}

func (t *Builder) Paginate(p *paginate.Paginate) *Builder {
	if p.First == 0 && p.After == "" {
		//向后分页 TODO
	}
	if p.Last == 0 && p.Before == "" {
		//向前分页 TODO
	}
	return t
}

func (t *Builder) PaginateOffSet(p *paginate.Paginate) *Builder {
	t.paginate = p
	return t
}

func (t *Builder) parsePaginateOffSet(){
	if t.paginate == nil {
		return
	}
	var limit,page int
	if t.paginate.First != 0 && t.paginate.After != "" {
		//向后分页
		page, _ = strconv.Atoi(t.paginate.After)
		limit = int(t.paginate.First)
	}
	if t.paginate.Last != 0 && t.paginate.Before != "" {
		//向前分页
		page, _ = strconv.Atoi(t.paginate.Before)
		limit = int(t.paginate.Last)
	}
	if page == 0 {
		page = 1
	}
	t.db = t.db.Offset((int32(page) - 1) * t.paginate.First).Limit(limit)
}

func (t *Builder) Order(order string) *Builder {
	t.order = order
	return t
}

// 返回即将执行的的Db
func (t *Builder) Prepare(count bool) *gorm.DB {
	if count {
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