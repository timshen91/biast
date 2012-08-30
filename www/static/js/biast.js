function dir(obj) {
	var names = obj + "\n";
	for (var name in obj) {
		names += name + ": " + obj[name] + "\n";
	}
	return names
}

function quote(id, author) {
	var quoteStr = '<blockquote cite="#comment-' + id + '">'
		+ '<a href="' + document.getElementById("comment-" + id).baseURI + '#comment-' + id + '">'
		+ author + '</a>:'
		+ document.getElementById("comment_content-" + id).innerHTML + '</blockquote>\n';
	document.getElementById("comment_form").content.value += quoteStr;
}
