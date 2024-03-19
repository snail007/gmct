package model

var (
	tpl = `package {{PKG}}

import (
	"github.com/snail007/gmc/core"
	"strconv"
	"sync"
)

type {{TABLE_STRUCT_NAME}}Model struct {
	db    gcore.Database
	table string
	primaryKey  string
	once  *sync.Once
}

var (
	{{TABLE_STRUCT_NAME}} = New{{TABLE_STRUCT_NAME}}Model()
)

func New{{TABLE_STRUCT_NAME}}Model() *{{TABLE_STRUCT_NAME}}Model {
	u := &{{TABLE_STRUCT_NAME}}Model{
		table: "{{TABLE_NAME}}",
		primaryKey:  "{{TABLE_PKEY}}",
		once:  &sync.Once{},
	}
	return u
}

func (s *{{TABLE_STRUCT_NAME}}Model) DB() gcore.Database {
	if s.db == nil {
		s.once.Do(func() {
			s.db = gmc.DB.DB()
		})
	}
	return s.db
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetByID(id string) (ret map[string]string, error error) {
	return s.GetByIDWithFields("*", id)
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetByIDWithFields(fields string, id string) (ret map[string]string, error error) {
	db := s.DB()
	rs, err := db.Query(db.AR().Select(fields).From(s.table).Where(map[string]interface{}{
		s.primaryKey: id,
	}).Limit(0, 1))
	if err != nil {
		return nil, err
	}
	ret = rs.Row()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetBy(where map[string]interface{}) (ret map[string]string, error error) {
	return s.GetByWithFields("*", where)
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetByWithFields(fields string, where map[string]interface{}) (ret map[string]string, error error) {
	db := s.DB()
	rs, err := db.Query(db.AR().Select(fields).From(s.table).Where(where).Limit(0, 1))
	if err != nil {
		return nil, err
	}
	ret = rs.Row()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) MGetByIDs(ids []string, orderBy ...interface{}) (ret map[string]string, error error) {
	return s.MGetByIDsWithFields("*", ids, orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) MGetByIDsWithFields(fields string, ids []string, orderBy ...interface{}) (ret map[string]string, error error) {
	db := s.DB()
	ar := db.AR().Select(fields).From(s.table).Where(map[string]interface{}{
		s.primaryKey: ids,
	})
	if col, by := s.OrderBy(orderBy...); col != "" {
		ar.OrderBy(col, by)
	}
	rs, err := db.Query(ar)
	if err != nil {
		return nil, err
	}
	ret = rs.Row()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetAll(orderBy ...interface{}) (ret []map[string]string, error error) {
	return s.GetAllWithFields("*", orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) GetAllWithFields(fields string, orderBy ...interface{}) (ret []map[string]string, error error) {
	return s.MGetByWithFields(fields, nil, orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) MGetBy(where map[string]interface{}, orderBy ...interface{}) (ret []map[string]string, error error) {
	return s.MGetByWithFields("*", where, orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) MGetByWithFields(fields string, where map[string]interface{}, orderBy ...interface{}) (ret []map[string]string, error error) {
	db := s.DB()
	ar := db.AR().Select(fields).From(s.table).Where(where).Limit(0, 1)
	if col, by := s.OrderBy(orderBy...); col != "" {
		ar.OrderBy(col, by)
	}
	rs, err := db.Query(ar)
	if err != nil {
		return nil, err
	}
	ret = rs.Rows()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) DeleteBy(where map[string]interface{}) (cnt int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().Delete(s.table, where))
	if err != nil {
		return 0, err
	}
	cnt = rs.RowsAffected()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) DeleteByIDs(ids []string) (cnt int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().Delete(s.table, map[string]interface{}{
		s.primaryKey: ids,
	}))
	if err != nil {
		return 0, err
	}
	cnt = rs.RowsAffected()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) Insert(data map[string]interface{}) (lastInsertID int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().Insert(s.table, data))
	if err != nil {
		return 0, err
	}
	lastInsertID = rs.LastInsertID()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) InsertBatch(data []map[string]interface{}) (cnt, lastInsertID int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().InsertBatch(s.table, data))
	if err != nil {
		return 0, 0, err
	}
	lastInsertID = rs.LastInsertID()
	cnt = rs.RowsAffected()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) UpdateByIDs(ids []string, data map[string]interface{}) (cnt int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().Update(s.table, data, map[string]interface{}{
		s.primaryKey: ids,
	}))
	if err != nil {
		return 0, err
	}
	cnt = rs.RowsAffected()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) UpdateBy(where, data map[string]interface{}) (cnt int64, err error) {
	db := s.DB()
	rs, err := db.Exec(db.AR().Update(s.table, data, where))
	if err != nil {
		return 0, err
	}
	cnt = rs.RowsAffected()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) Page(where map[string]interface{}, offset, length int, orderBy ...interface{}) (ret []map[string]string, total int, err error) {
	return s.PageWithFields("*", where, offset, length, orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) PageWithFields(fields string, where map[string]interface{}, offset, length int, orderBy ...interface{}) (ret []map[string]string, total int, err error) {
	db := s.DB()
	ar := db.AR().Select("count(*) as total").From(s.table)
	if len(where) > 0 {
		ar.Where(where)
	}
	rs, err := db.Query(ar)
	if err != nil {
		return nil, 0, err
	}
	t, _ := strconv.ParseInt(rs.Value("total"), 10, 64)
	total = int(t)
	ar = db.AR().Select(fields).From(s.table).Where(where).Limit(offset, length)
	if len(where) > 0 {
		ar.Where(where)
	}
	if col, by := s.OrderBy(orderBy...); col != "" {
		ar.OrderBy(col, by)
	}
	rs, err = db.Query(ar)
	if err != nil {
		return nil, 0, err
	}
	ret = rs.Rows()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) List(where map[string]interface{}, offset, length int, orderBy ...interface{}) (ret []map[string]string, err error) {
	return s.ListWithFields("*", where, offset, length, orderBy...)
}

func (s *{{TABLE_STRUCT_NAME}}Model) ListWithFields(fields string, where map[string]interface{}, offset, length int, orderBy ...interface{}) (ret []map[string]string, err error) {
	db := s.DB()
	ar := db.AR().Select(fields).From(s.table).Where(where).Limit(offset, length)
	if len(where) > 0 {
		ar.Where(where)
	}
	if col, by := s.OrderBy(orderBy...); col != "" {
		ar.OrderBy(col, by)
	}
	rs, err := db.Query(ar)
	if err != nil {
		return nil, err
	}
	ret = rs.Rows()
	return
}

func (s *{{TABLE_STRUCT_NAME}}Model) OrderBy(orderBy ...interface{}) (col, by string) {
	if len(orderBy) > 0 {
		switch val := orderBy[0].(type) {
		case map[string]interface{}:
			for k, v := range val {
				col, by = k, v.(string)
				break
			}
		case map[string]string:
			for k, v := range val {
				col, by = k, v
				break
			}
		}
	}
	return
}
`
)
