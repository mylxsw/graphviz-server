function getQueryString(name) {
    var url = document.URL;
    var arr = url.split('?');
    if (arr.length < 2)    {
        return '';
    }
    url = arr[1];

    if (url.lastIndexOf('#') == (url.length - 1))
        url = url.substring(0, url.length - 1);

    var arrQueryStringPair = url.split('&');
    if (arrQueryStringPair.length == 0)
        return '';


    for (var i = 0; i < arrQueryStringPair.length; ++i)    {
        var startIndex = arrQueryStringPair[i].indexOf('=') + 1;
        var sName = arrQueryStringPair[i].substr(0, startIndex - 1);
        var result = arrQueryStringPair[i].substr(startIndex, arrQueryStringPair[i].length - startIndex);
        if (sName.toLowerCase() == name.toLowerCase()) {
            return result;
        }
    }

    return '';
}

function interactMode(selector, imageURL) {
    $(selector).graphviz({
        url: imageURL,
        ready: function() {
            let gv = this
            gv.nodes().click(function () {
                let $set = $()
                $set.push(this)
                $set = $set.add(gv.linkedFrom(this, true))
                $set = $set.add(gv.linkedTo(this, true))
                gv.highlight($set, true)
                gv.bringToFront($set)
            })
            $(document).keydown(function (evt) {
                if (evt.keyCode === 27) {
                    gv.highlight()
                }
            });
        }
    });
}

/**
 * fillStyle: hachure,solid,zigzag,cross-hatch,dots,dashed,zigzag-line
 *
 * @param container
 */
function sketchMode(container, roughness) {
    roughness = roughness === '' ? 0:roughness;
    const option = {
        fillStyle: 'hachure',
        roughness: roughness,
        bowing: 1,
        chartType: 'highcharts',
    };
    const handler = Sketchifier(container, option);
    handler.handify();
}