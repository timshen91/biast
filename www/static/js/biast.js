function trim(str) {
	return str.replace(/^\s+/,'').replace(/\s+$/,'');
}

function quote(id, author) {
	document.getElementById("comment_form").content.value += '<blockquote cite="#comment-' + id + '">'
		+ '<a href="' + document.getElementById("comment-" + id).baseURI + '#comment-' + id + '">'
		+ author + '</a>: '
		+ trim(document.getElementById("comment_content-" + id).innerHTML) + '</blockquote>\n';
}

function addTag(tag) {
	if (document.getElementById("tag_input").value.length != 0) {
		document.getElementById("tag_input").value += ", ";
	}
	document.getElementById("tag_input").value += tag;
}
