package controller

var (
	tpl= `package %s

import (
	"github.com/snail007/gmc"
	gcast "github.com/snail007/gmc/util/cast"
	gmap "github.com/snail007/gmc/util/map"
	"github.com/gookit/validate"
	"time"
)

type {{HOLDER}} struct {
	Base
}

func (this *{{HOLDER}}) Index() {

}

func (this *{{HOLDER}}) List() {
	enableSearch := true
	where := gmap.M{}
	search := this.Ctx.GET("search_field")
	keyword := this.Ctx.GET("keyword")
	if enableSearch && search != "" && keyword != "" {
		data, err := validate.FromRequest(this.Request)
		if err != nil {
			this.Stop(err)
		}
		v := data.Create()
		v.StringRule("search_field", "enum:name,age")
		if !v.Validate() {
			this.Stop(v.Errors.One())
		}
		where = gmap.M{search + " like": "%%" + keyword + "%%"}
	}
	page := gcast.ToInt(this.Ctx.GET("page"))
	pageSize := gcast.ToInt(this.Ctx.GET("count"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	table := gmc.DB.Table("{{TABLE}}")
	rows, total, err := table.Page(where, start, pageSize, gmap.M{"{{TABLE}}_id": "desc"})
	if err != nil {
		this.Stop(err)
	}
	this.View.Set("rows", rows)
	this.View.Set("enable_search", enableSearch)
	this.View.Set("paginator", this.Ctx.NewPager(pageSize, int64(total)))
	this.View.Layout("list").Render("{{TABLE}}/list")
}

func (this *{{HOLDER}}) Detail() {
	id := this.Ctx.GET("{{TABLE}}_id")
	if id == "" {
		id = this.Ctx.POST("{{TABLE}}_id")
	}
	table := gmc.DB.Table("{{TABLE}}")
	row, err := table.GetByID(id)
	if err != nil {
		this._JSONFail(err.Error())
	}
	if len(row) == 0 {
		this._JSONFail("data not found")
	}
	this.View.Set("data",row)
	this.View.Layout("list").Render("{{TABLE}}/detail")
}

func (this *{{HOLDER}}) Create() {
	if this.Ctx.IsPOST() {
		// do create
		data, err := validate.FromRequest(this.Request)
		if err != nil {
			this.Stop(err)
		}
		v := data.Create()
		v.FilterRule("age", "int")
		v.StringRule("age", "required|max:99")
		v.StringRule("name", "required|minLen:7")
		if !v.Validate() { // validate ok
			this._JSONFail(v.Errors.One())
		}
		table := gmc.DB.Table("{{TABLE}}")
		dataInsert := gmap.M{}
		dataInsert["name"], _ = data.Get("name")
		dataInsert["age"], _ = data.Get("age")
		dataInsert["create_time"] = time.Now().Unix()
		_, err = table.Insert(dataInsert)
		if err != nil { // validate ok
			this._JSONFail(err.Error())
		}
		this._JSONSuccess("","","/{{TABLE}}/list")
	} else {
		// show create page
		this.View.Layout("list").Render("{{TABLE}}/form")
	}
}

func (this *{{HOLDER}}) Edit() {
	data, err := validate.FromRequest(this.Request)
	if err != nil {
		this.Stop(err)
	}
	table := gmc.DB.Table("{{TABLE}}")
	id := this.Ctx.GET("{{TABLE}}_id")
	if id == "" {
		id = this.Ctx.POST("{{TABLE}}_id")
	}
	row, err := table.GetByID(id)
	if err != nil {
		this._JSONFail(err.Error())
	}
	if len(row) == 0 {
		this._JSONFail("data not found")
	}
	if this.Ctx.IsPOST() {
		// do create
		v := data.Create()
		v.FilterRule("{{TABLE}}_id", "int")
		v.FilterRule("age", "int")

		v.StringRule("{{TABLE}}_id", "required|min:1")
		v.StringRule("age", "required|max:99")
		v.StringRule("name", "required|minLen:7")
		if !v.Validate() { // validate ok
			this._JSONFail(v.Errors.One())
		}

		dataInsert := gmap.M{}
		dataInsert["name"], _ = data.Get("name")
		dataInsert["age"], _ = data.Get("age")
		dataInsert["update_time"] = time.Now().Unix()
		_, err = table.UpdateBy(gmap.M{"{{TABLE}}_id": id}, dataInsert)
		if err != nil { // validate ok
			this._JSONFail(err.Error())
		}
		this._JSONSuccess("", "", "/{{TABLE}}/list")
	} else {
		// show create page
		this.View.Set("data", row)
		this.View.Layout("list").Render("{{TABLE}}/form")
	}
}

func (this *{{HOLDER}}) Delete() {
	var ids []string
	this.Request.ParseForm()
	id := this.Request.Form["ids"]
	if len(id) > 0 {
		ids = append(ids, id...)
	}
	table := gmc.DB.Table("{{TABLE}}")
	_, err := table.DeleteByIDs(ids)
	this.StopE(err, func() {
		this._JSONFail(err.Error())
	})
	this._JSONSuccess("", nil, "/{{TABLE}}/list")
}
`
)
