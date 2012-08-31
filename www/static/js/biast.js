function trim(str) {
	return str.replace(/^\s+/,'').replace(/\s+$/,'');
}

function quote(id, author) {
	var str = window.getSelection().toString();
	if (str.length == 0) {
		str = trim(document.getElementById("comment_content-" + id).innerHTML);
	}
	document.getElementById("comment_form").content.value += '<blockquote cite="#comment-' + id + '">'
		+ '<a href="' + document.getElementById("comment-" + id).baseURI + '#comment-' + id + '">'
		+ author + '</a>: '
		+ str + '</blockquote>\n';
}

function addTag(tag) {
	if (document.getElementById("tag_input").value.length != 0) {
		document.getElementById("tag_input").value += ", ";
	}
	document.getElementById("tag_input").value += tag;
}
