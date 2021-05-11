package view

var (
	listTpl = `{{$controllerPath:="/{{CONTROLLER}}"}}
{{$idName:="{{TABLE}}_id"}}
<div class="container-fluid p-t-15">
    <div class="row">
        <div class="col-lg-12">
            <div class="card">
                <div class="card-toolbar d-flex flex-column flex-md-row">
                    <div class="toolbar-btn-action">
                        <a class="btn btn-primary m-r-5" href="{{$controllerPath}}/create"><i class="mdi mdi-plus"></i>
                            新增</a>
                        <a class="btn btn-danger m-r-5 ajax-post confirm" href="{{$controllerPath}}/delete"
                           target-form="ids"><i class="mdi mdi-check"></i> 删除</a>
                    </div>
                    {{if .enable_search}}
                        <form class="search-bar ml-md-auto" method="get" action="" role="form">
                             <div class="input-group ml-md-auto">
								<div class="btn-group btn-group-xs">
									<input type="hidden" name="search_field">
									<button type="button" class="btn btn-default dropdown-toggle" data-toggle="dropdown"
											aria-haspopup="true" aria-expanded="false"></button>
									<ul class="dropdown-menu" value="">
										<li><a class="dropdown-item" href="javascript:void(0);" value="name">姓名</a></li>
										<li><a class="dropdown-item" href="javascript:void(0);" value="age">年龄</a></li>
									</ul>
								</div>
                                <input type="text" class="form-control" name="keyword" placeholder="请输入">
								{{if .G.keyword}}
                                	<a class="btn btn-warning m-r-5" href="{{$controllerPath}}/list"><i class="mdi mdi-check"></i> 重置</a>
                            	{{end}}
                            </div>
                        </form>
                    {{end}}
                </div>
                <div class="card-body">
                    <div class="table-responsive">
                        <table class="table table-bordered">
                            <thead>
                            <tr>
                                <th>
                                    <div class="custom-control custom-checkbox">
                                        <input type="checkbox" class="custom-control-input" id="check-all">
                                        <label class="custom-control-label" for="check-all"></label>
                                    </div>
                                </th>
                                <th>用户ID</th>
                                <th>姓名</th>
                                <th>年龄</th>
                                <th>修改时间</th>
                                <th>创建时间</th>
                                <th>操作</th>
                            </tr>
                            </thead>
                            <tbody>
                            {{ range $row := .rows}}
                                <tr>
                                    <td>
                                        <div class="custom-control custom-checkbox">
                                            <input type="checkbox" class="custom-control-input ids" name="ids"
                                                   value="{{val $row $idName}}"
                                                   id="ids-{{val $row $idName}}">
                                            <label class="custom-control-label" for="ids-{{val $row $idName}}"></label>
                                        </div>
                                    </td>
                                    <td><a href="{{$controllerPath}}/detail?{{$idName}}={{val $row $idName}}">{{val $row $idName}}</a></td>
                                    <td>{{$row.name}}</td>
                                    <td>{{$row.age}}</td>
                                    <td>{{if eq $row.update_time "0"}}-{{else}}{{$row.update_time | date "2006-01-02 15:04"}}{{end}}</td>
                                    <td>{{$row.create_time | date "2006-01-02 15:04"}}</td>
                                    <td>
                                        <div class="btn-group">
                                            <a class="btn btn-xs btn-default"
                                               href="{{$controllerPath}}/edit?{{$idName}}={{val $row $idName}}" title=""
                                               data-toggle="tooltip"
                                               data-original-title="编辑"><i class="mdi mdi-pencil"></i></a>
                                            <a class="btn btn-xs btn-default ajax-get confirm"
                                               href="{{$controllerPath}}/delete?ids={{val $row $idName}}" title=""
                                               data-toggle="tooltip"
                                               data-original-title="删除"><i class="mdi mdi-window-close"></i></a>
                                        </div>
                                    </td>
                                </tr>
                            {{end}}
                            </tbody>
                        </table>
                    </div>
                    {{template "paginator/default" .}}
                </div>
            </div>
        </div>
    </div>
</div>`
	formTpl = `{{$controllerPath:="/{{CONTROLLER}}"}}
<div class="card">
    <div class="card-body">
        <form class="form-horizontal ajax-form" action="{{$controllerPath}}/{{if .data}}edit{{else}}create{{end}}" method="post" enctype="application/x-www-form-urlencoded">
            {{if .data }}
            {{$name:="{{TABLE}}_id"}}
            <input type="hidden" name="{{$name}}" id="{{$name}}" value="{{val .data $name}}">
            {{end}}
            <div class="form-group">
                {{$name:="name"}}
                <label for="{{$name}}">姓名</label>
                <input type="text" class="form-control" name="{{$name}}" id="{{$name}}" placeholder="" value="{{val .data $name}}">
            </div>
            <div class="form-group">
                {{$name:="age"}}
                <label for="{{$name}}">年龄</label>
                <input type="text" class="form-control" name="{{$name}}" id="{{$name}}" placeholder="" value="{{val .data $name}}">
            </div>
            <div class="col-auto">
                <button type="submit" class="btn btn-primary mb-2">
					<span class="spinner-grow spinner-grow-sm" role="status" style="display: none;" aria-hidden="true"></span>
					{{if .data }}保存{{else}}添加{{end}}
				</button>
                <a href="{{$controllerPath}}/list" class="btn btn-default mb-2">返回</a>
            </div>
        </form>
    </div>
</div>`
	detailTpl = `<div class="container-fluid p-t-15">
    <div class="row">
        <div class="col-lg-12">
            <div class="card">
                <div class="card-body">
                    <h3>详细信息</h3>
                    <hr>
                    <table class="table table-condensed">
                        <tbody>
                        <tr>
                            <td style="width: 150px;">姓名</td>
                            <td>{{val .data "name"}}</td>
                        </tr>
                        <tr>
                            <td>年龄</td>
                            <td>{{val .data "age"}}</td>
                        </tr>
                        <tr>
                            <td>修改时间</td>
                            <td>{{if eq .data.update_time "0"}}-{{else}}{{.data.update_time | date "2006-01-02 15:04"}}{{end}}</td>
                        </tr>
                        <tr>
                            <td>创建时间</td>
                            <td>{{.data.create_time | date "2006-01-02 15:04"}}</td>
                        </tr>
						<tr>
                            <td></td>
                            <td><a class="btn btn-outline-success" href="javascript:history.go(-1);">返回</a></td>
                        </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</div>`
)
