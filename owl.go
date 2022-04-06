package owl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"golang.org/x/net/html"
)

type Root struct {
	Node      *html.Node
	NodeValue string
	Error     *Error
}

func HTMLParse(r io.Reader) *Root {
	return htmlparsing(r)
}

func HTMLParseFromString(s string) *Root {
	return htmlparsing(strings.NewReader(s))
}

func htmlparsing(r io.Reader) *Root {
	root, err := html.Parse(r)
	if err != nil {
		return &Root{Node: nil, NodeValue: "",
			// Error: newError(ErrUnableToParse, err.Error()),
			Error: newError(ErrUnableToParse, err),
		}
	}
	for root.Type != html.ElementNode {
		switch root.Type {
		case html.DocumentNode:
			root = root.FirstChild
		case html.DoctypeNode:
			root = root.NextSibling
		case html.CommentNode:
			root = root.NextSibling
		}
	}
	return &Root{Node: root, NodeValue: root.Data, Error: nil}
}

// Find finds the first occurrence of the given tag name,
// with or without attribute key and value specified,
// and returns a struct with a Node to it

func (r *Root) Find(args ...string) *Root {
	temp, ok := findOnce(r.Node, args, false, false)
	if !ok {
		return &Root{Node: nil, NodeValue: "", Error: &Error{
			Type: ErrElementNotFound,
			msg:  errors.New("given element and attriabutes not found"),
		},
		}
	}
	return &Root{Node: temp, NodeValue: temp.Data, Error: nil}
}

// FindStrict finds the first occurrence of the given tag name
// only if all the values of the provided attribute are an exact match
func (r *Root) FindStrict(args ...string) *Root {
	temp, ok := findOnce(r.Node, args, false, true)
	if !ok {
		return &Root{Node: nil, NodeValue: "", Error: &Error{
			Type: ErrElementNotFound,
			msg:  errors.New("given element and attriabutes not found"),
		},
		}
	}

	return &Root{Node: temp, NodeValue: temp.Data, Error: nil}
}

func (r *Root) Title() *Root {
	var slic []string = []string{"title"}
	re, exits := findOnce(r.Node, slic, false, true)
	if !exits {
		return &Root{Node: nil, NodeValue: "", Error: &Error{
			Type: ErrElementNotFound,
			msg:  errors.New("given element and attriabutes not found"),
		},
		}
	}
	return &Root{Node: re, NodeValue: re.Data, Error: nil}
}

// FindNextSibling finds the next sibling of the Node in the DOM
// returning a struct with a Node to it
func (r *Root) FindNextSibling() *Root {
	nextSibling := r.Node.NextSibling
	if nextSibling == nil {
		return &Root{Node: nil, NodeValue: "", Error: newError(ErrNoNextSibling, errors.New("no next sibling found"))}
	}
	return &Root{Node: nextSibling, NodeValue: nextSibling.Data, Error: nil}
}

func (r *Root) FindPrevSibling() *Root {
	prevSibling := r.Node.PrevSibling
	if prevSibling == nil {
		return &Root{Node: nil, NodeValue: "", Error: newError(ErrNoNextSibling, errors.New("no previous sibling found"))}

	}
	return &Root{Node: prevSibling, NodeValue: prevSibling.Data, Error: nil}
}

// FindNextElementSibling finds the next element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindNextElementSibling() *Root {
	nextSibling := r.Node.NextSibling
	if nextSibling == nil {
		return &Root{Node: nil, NodeValue: "", Error: newError(ErrNoNextSibling, errors.New("no next element sibling found"))}
	}
	if nextSibling.Type == html.ElementNode {
		return &Root{Node: nextSibling, NodeValue: nextSibling.Data, Error: nil}
	}
	p := &Root{Node: nextSibling, NodeValue: nextSibling.Data}
	return p.FindNextElementSibling()
}

// FindPrevElementSibling finds the previous element sibling of the pointer in the DOM
// returning a struct with a pointer to it
func (r Root) FindPrevElementSibling() *Root {
	prevSibling := r.Node.PrevSibling
	if prevSibling == nil {
		return &Root{Node: nil, NodeValue: "", Error: newError(ErrNoNextSibling, errors.New("no previous element sibling found"))}
	}
	if prevSibling.Type == html.ElementNode {
		return &Root{Node: prevSibling, NodeValue: prevSibling.Data, Error: nil}
	}
	p := Root{Node: prevSibling, NodeValue: prevSibling.Data}
	return p.FindPrevElementSibling()
}

// FullText returns the string inside even a nested element
func (r Root) FullText() string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		if n.Type == html.ElementNode {
			f(n.FirstChild)
		}
		if n.NextSibling != nil {
			f(n.NextSibling)
		}
	}

	f(r.Node.FirstChild)

	return buf.String()
}

// HTML returns the HTML code for the specific element
func (r Root) Render() []byte {
	var buf bytes.Buffer
	if err := html.Render(&buf, r.Node); err != nil {
		return nil
	}
	return buf.Bytes()
}

type Roots struct {
	Roots [](*Root)
	Len   int
	Error *Error
}

func (r *Root) FindAll(args ...string) Roots {
	temp := findAllofem(r.Node, args, false)
	length := len(temp)
	if length == 0 {
		return Roots{Roots: nil, Error: newError(ErrElementsNotFound, errors.New("no elements or attriabutes found"))}
	}
	Nodes := make([](*Root), 0, length)
	for i := 0; i < length; i++ {
		Nodes = append(Nodes, &Root{Node: temp[i], NodeValue: temp[i].Data})
	}
	return Roots{Roots: Nodes, Len: length, Error: nil}
}

func (rs Roots) First() *Root {
	return rs.Roots[0]
}
func (rs Roots) Last() *Root {
	return rs.Roots[rs.Len-1]
}

// FindAllStrict finds all occurrences of the given tag name
// only if all the values of the provided attribute are an exact match
func (r Root) FindAllStrict(args ...string) Roots {
	temp := findAllofem(r.Node, args, true)
	length := len(temp)
	if length == 0 {
		return Roots{Roots: nil, Len: 0, Error: newError(ErrElementNotFound, fmt.Errorf("element `%s` with attributes `%s` not found", args[0], strings.Join(args[1:], " ")))}
	}
	Nodes := make([](*Root), 0, length)
	for i := 0; i < length; i++ {
		Nodes = append(Nodes, &Root{Node: temp[i], NodeValue: temp[i].Data})
	}
	return Roots{Roots: Nodes, Len: length, Error: nil}
}

func (rs Roots) ForEach(f func(int, *Root)) *Root {
	var (
		i int
		r *Root
	)
	for i, r = range rs.Roots {
		f(i, r)
	}
	return r
}

// Text returns the string inside a non-nested element
func (r *Root) Text() string {
	var f func(*html.Node) string
	k := r.Node.FirstChild

	f = func(n *html.Node) string {
		if n != nil && n.Type != html.TextNode {
			if n = n.NextSibling; n == nil {
				return ""
			}
			f(n)
		}
		if k != nil {
			r, _ := regexp.Compile(`^\s+$`)
			if ok := r.MatchString(k.Data); ok {
				if n = n.NextSibling; n == nil {
					return ""
				}
				f(n)
			}
			return k.Data
		}
		return ""
	}
	return f(k)
}

// Attrs returns a map containing all attributes
func (r *Root) Attrs() map[string]string {
	if (r.Node.Type != html.ElementNode) && (len(r.Node.Attr) == 0) {
		return nil
	}
	return getKeyValue(r.Node.Attr)
}

func (r Root) Children() Roots {
	childNode := r.Node.FirstChild
	var (
		childrenNode Roots
		rootNode     [](*Root)
	)
	for childNode != nil {
		rootNode = append(rootNode, &Root{Node: childNode, NodeValue: childNode.Data})
		childrenNode.Roots = rootNode
		childrenNode.Len = len(rootNode)

		childNode = childNode.NextSibling
	}
	return childrenNode
}

// This is for Scraping HTML documents for a Visited Link
func (r *Root) Visit(str string, client *Client) (*Root, error) {
	var c *Client
	g := glob.MustCompile("https://*, http://*, /*")
	if !g.Match(str) {
		return nil, fmt.Errorf("string %s is not a link", str)
	}
	if client == nil {
		c = NewClient(nil)
	}
	reader, err := c.Get(str)
	return HTMLParse(reader), err
}

// This Download files, this is different from Visit
func (r *Root) Download(url string, client *Client) ([]byte, error) {
	var (
		body []byte
		err  error
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, errors.New("file is corrupted and sothing else happened")
	}
	return body, nil
}

func matchElementName(n *html.Node, name string) bool {
	return name == "" || name == n.Data
}

// attributeAndValueEquals reports when the html.Attribute attr has the same attribute name and value as from
// provided arguments
func attributeAndValueEquals(attr html.Attribute, attribute, value string) bool {
	return attr.Key == attribute && attr.Val == value
}

// attributeContainsValue reports when the html.Attribute attr has the same attribute name as from provided
// attribute argument and compares if it has the same value in its values parameter
func attributeContainsValue(attr html.Attribute, attribute, value string) bool {
	if attr.Key == attribute {
		for _, attrVal := range strings.Fields(attr.Val) {
			if attrVal == value {
				return true
			}
		}
	}
	return false
}

// Using depth first search to find the first occurrence and return
func findOnce(n *html.Node, args []string, uni bool, strict bool) (*html.Node, bool) {
	if uni {
		if n.Type == html.ElementNode && matchElementName(n, args[0]) {
			if len(args) > 1 && len(args) < 4 {
				for i := 0; i < len(n.Attr); i++ {
					attr := n.Attr[i]
					searchAttrName := args[1]
					searchAttrVal := args[2]
					if (strict && attributeAndValueEquals(attr, searchAttrName, searchAttrVal)) ||
						(!strict && attributeContainsValue(attr, searchAttrName, searchAttrVal)) {
						return n, true
					}
				}
			} else if len(args) == 1 {
				return n, true
			}
		}
	}
	uni = true
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p, q := findOnce(c, args, true, strict)
		if q {
			return p, q
		}
	}
	return nil, false
}

// Using depth first search to find all occurrences and return
func findAllofem(n *html.Node, args []string, strict bool) []*html.Node {
	var nodeLinks = make([]*html.Node, 0, 10)
	var f func(*html.Node, []string, bool)
	f = func(n *html.Node, args []string, uni bool) {
		if uni {
			if n.Type == html.ElementNode && matchElementName(n, args[0]) {
				if len(args) > 1 && len(args) < 4 {
					for i := 0; i < len(n.Attr); i++ {
						attr := n.Attr[i]
						searchAttrName := args[1]
						searchAttrVal := args[2]
						if (strict && attributeAndValueEquals(attr, searchAttrName, searchAttrVal)) ||
							(!strict && attributeContainsValue(attr, searchAttrName, searchAttrVal)) {
							nodeLinks = append(nodeLinks, n)
						}
					}
				} else if len(args) == 1 {
					nodeLinks = append(nodeLinks, n)
				}
			}
		}
		uni = true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, args, true)
		}
	}
	f(n, args, false)
	return nodeLinks
}

// Returns a key pair value (like a dictionary) for each attribute
func getKeyValue(attributes []html.Attribute) map[string]string {
	length := len(attributes)
	var keyvalues = make(map[string]string, length)
	for i := 0; i < length; i++ {
		_, exists := keyvalues[attributes[i].Key]
		if !exists {
			keyvalues[attributes[i].Key] = attributes[i].Val
		}
	}
	return keyvalues
}
