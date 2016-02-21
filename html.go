package glc

import (
	"bytes"
	"strings"
	"golang.org/x/net/html"
)

const (
	elipse = "..."
	anchorElement = "a"
)

func findTextNodes(node *html.Node) []*html.Node {
	var textNodes []*html.Node

	if node.Type == html.TextNode {
		return append(textNodes, node)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		textNode := findTextNodes(child)
		if textNode != nil {
			textNodes = append(textNodes, textNode...)
		}

	}

	return textNodes
}

func nodeToString(nodes []*html.Node) string {
	var snippet bytes.Buffer
	for _, node := range nodes {
		if e := html.Render(&snippet, node); e != nil {
			panic(e)
		}
	}

	return snippet.String()
}

// Take an excerpt from html.Node respecting element boundaries. radius chars will be taken from each side of node, only sibling nodes are truncated.
func excerptHTML(node *html.Node, radius int) string {
	var excerptNodes []*html.Node
	leftLen, rightLen := radius, radius

	for truncated, nodeLeft := false, node.PrevSibling; nodeLeft != nil && !truncated; nodeLeft = nodeLeft.PrevSibling {
		excerptNodes = append([]*html.Node{nodeLeft}, excerptNodes...)

		textNodes := findTextNodes(nodeLeft)
		for i := len(textNodes) - 1; i >= 0; i-- {
			if truncated {
				textNodes[i].Data = ""
				continue
			}

			if len(textNodes[i].Data) <= leftLen {
				leftLen -= len(textNodes[i].Data)
			} else {
				textNodes[i].Data = elipse + textNodes[i].Data[len(textNodes[i].Data) - leftLen: len(textNodes[i].Data)]
				truncated = true
			}
		}
	}

	excerptNodes = append(excerptNodes, node)

	for truncated, nodeRight := false, node.NextSibling; nodeRight != nil && !truncated; nodeRight = nodeRight.NextSibling {
		excerptNodes = append(excerptNodes, nodeRight)

		for _, textNode := range findTextNodes(nodeRight) {
			if truncated {
				textNode.Data = ""
				continue
			}

			if len(textNode.Data) <= rightLen {
				rightLen -= len(textNode.Data)
			} else {
				textNode.Data = textNode.Data[0:rightLen] + elipse
				truncated = true
			}
		}
	}

	return nodeToString(excerptNodes)
}

// Parses an HTML string and walks the resulting tree, calling handler() for each ElementNode
func findLinks(handler func(*html.Node), rawhtml string) error {
	doc, err := html.Parse(strings.NewReader(rawhtml))
	if err != nil {
		return err
	}

	var walker func(*html.Node)
	walker = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == anchorElement {
			handler(node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walker(child)
		}
	}

	walker(doc)

	return nil
}
