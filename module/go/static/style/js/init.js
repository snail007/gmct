;jQuery(function () {
    for (let i = 0; i < window.init.length; i++) {
        window.init[i]();
    }
    // 表单或者列表下拉框
    $('.dropdown-menu').each(function () {
        var ul = $(this);
        var btn = ul.prev();
        var input = btn.prev();
        var items = $(this).find('.dropdown-item')
        var value = ul.attr("value");
        var setValue = function (item) {
            var value = item.attr("value");
            var text = item.text();
            btn.text(text)
            input.val(value)
        }
        var found = false
        items.each(function () {
            if ($(this).attr("value") == value) {
                setValue($(this))
                found = true
                return false;
            }
        });
        if (!found) {
            setValue(items.eq(0))
        }
        items.click(function () {
            setValue($(this))
        });
    });
});