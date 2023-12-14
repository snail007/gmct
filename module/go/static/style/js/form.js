var callback = function (msg, typ, resp, a, b, c) {
    if (window[msg]) {
        window[msg](resp, a, b, c);
    } else {
        let d = new Dialog().alert(msg, {
            type: typ,
        });
        $(d.element.dialog).find('button').hide();
        setTimeout(function () {
            d.remove();
        }, 1000)
    }
}
var successNotify = function (msg, dur) {
    dur = dur || 800;
    let d = new Dialog().alert(msg, {
        type: "success",
    });
    $(d.element.dialog).find('button').hide();
    setTimeout(function () {
        d.remove();
    }, dur);
}
$(document).ready(function () {
    $('.ajax-form').submit(function () {
        var form = $(this)
        $(this).ajaxSubmit({
            dataType: "json",
            beforeSerialize: function ($form, options) {
                var beforeSerialize = $form.attr("beforeSerialize");
                if (window[beforeSerialize]) {
                    return window[beforeSerialize]($form, options);
                }
                return true;
            },
            beforeSubmit: function (arr, $form, options) {
                var beforeSubmit = $form.attr("beforeSubmit");
                if (window[beforeSubmit]) {
                    return window[beforeSubmit](arr, $form, options);
                }
                return true;
            },
            beforeSend: function () {
                form.find(".spinner-grow").parent().attr("disabled", "disabled")
                form.find(".spinner-grow").show()
            },
            complete: function (a, b) {
                form.find(".spinner-grow").parent().removeAttr("disabled")
                form.find(".spinner-grow").hide()
            },
            error: function (a, b, c) {
                var onError = $("form").attr("onError");
                if (window[onError]) {
                    return window[onError](a, b, c);
                } else {
                    alert("请求错误，请重试。响应码：" + a.status)
                }
            },
            success: function (resp, statusText, jqXHR, jqForm) {
                var success = jqForm.attr("success");
                var fail = jqForm.attr("fail");
                var done = jqForm.attr("done");
                var isSuccess = resp.code == 200
                var timeout = 2000;
                if (isSuccess) {
                    if (success) {
                        callback(success, "success", resp, statusText, jqXHR, jqForm);
                    } else {
                        successNotify('操作成功！');
                    }
                } else {
                    timeout = 3000;
                    if (fail) {
                        callback(fail, "warning", resp, statusText, jqXHR, jqForm);
                    } else {
                        new Dialog().alert(resp.msg ? resp.msg : '操作失败，请重试！', {type: 'warning'});
                    }
                }
                if (done) {
                    callback(done, "remind", resp, statusText, jqXHR, jqForm);
                }
                if (resp.url) {
                    setTimeout(function () {
                        location = resp.url
                    }, timeout)
                } else if (resp.code != 200) {
                    alert(resp.msg || "操作失败，请重试。")
                }
            }
        })
        return false;
    })
});