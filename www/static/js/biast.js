function trim(str) {
    return str.replace(/^\s+/,'').replace(/\s+$/,'');
}

function quote(id, author) {
    var str = window.getSelection().toString();
    if (str.length == 0) {
        str = trim(document.getElementById("comment-content-" + id).innerHTML);
    }
    document.getElementById("comment-form").content.value += '<blockquote cite="#comment-' + id + '">'
        + '<a href="' + document.getElementById("comment-" + id).baseURI + '#comment-' + id + '">'
        + author + '</a>: '
        + str + '</blockquote>\n';
    location.href='#comment-post';
}

function addTag(tag) {
    if (document.getElementById("tag-input").value.length != 0) {
        document.getElementById("tag-input").value += ", ";
    }
    document.getElementById("tag-input").value += tag;
}
