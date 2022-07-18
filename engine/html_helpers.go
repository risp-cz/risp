package engine

import (
	"bytes"

	"golang.org/x/net/html"
)

func findHTMLNode(node *html.Node, check func(*html.Node) bool) (*html.Node, bool) {
	if check(node) {
		return node, true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result, ok := findHTMLNode(child, check); ok {
			return result, ok
		}
	}

	return nil, false
}

func removeAllTagsByName(name string, node *html.Node) {
	if node.Type == html.ElementNode && node.Data == name {
		node.Parent.RemoveChild(node)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		removeAllTagsByName(name, child)
	}
}

func removeAllTagAttributes(node *html.Node) {
	if node.Type == html.DoctypeNode || node.Type == html.CommentNode {
		node.Parent.RemoveChild(node)
		return
	}

	node.Attr = []html.Attribute{}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		removeAllTagAttributes(child)
	}
}

func sanitizeHTMLDocument(document *html.Node) (buffer bytes.Buffer, err error) {
	removeAllTagsByName("head", document)
	removeAllTagsByName("script", document)
	removeAllTagsByName("style", document)

	removeAllTagAttributes(document)

	err = html.Render(&buffer, document)
	return
}
