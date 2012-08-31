function trim(str) {
	return str.replace(/^\s+/,'').replace(/\s+$/,'');
}

function quote(id, author) {
	var quoteStr = '<blockquote cite="#comment-' + id + '">'
		+ '<a href="' + document.getElementById("comment-" + id).baseURI + '#comment-' + id + '">'
		+ author + '</a>: '
		+ trim(document.getElementById("comment_content-" + id).innerHTML) + '</blockquote>\n';
	document.getElementById("comment_form").content.value += quoteStr;
}
