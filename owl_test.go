package owl

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testHTML = `
<html>
  <head>
    <title>Sample "Hello, World" Application</title>
  </head>
  <body bgcolor=white>

    <table border="0" cellpadding="10">
      <tr>
        <td>
          <img src="images/springsource.png">
        </td>
        <td>
          <h1>Sample "Hello, World" Application</h1>
        </td>
      </tr>
    </table>
    <div id="0">
      <div id="1">Just two divs peacing out</div>
    </div>
    check
    <div id="2">One more</div>
    <p>This is the home page for the HelloWorld Web application. </p>
    <p>To prove that they work, you can execute either of the following links:
    <ul>
      <li>To a <a href="hello.jsp">JSP page</a> right?</li>
      <li>To a <a href="hello">servlet</a></li>
    </ul>
    </p>
    <div id="3">
      <div id="4">Last one</div>
    </div>
    <div id="5">
        <h1><span></span></h1>
    </div>
  </body>
</html>
`

const HtmlRoot2HTML = `
<html>
	<head>
		<title>Sample Application</title>
	</head>
	<body>
		<div class="first second">Multiple classes</div>
		<div class="first">Single class</div>
		<div class="second first third">Multiple classes inorder</div>
		<div>
			<div class="first">Inner single class</div>
			<div class="first second">Inner multiple classes</div>
			<div class="second first">Inner multiple classes inorder</div>
			<div class="third first">Inner multiple classes inorder</div>
		</div>
	</body>
</html>
`

var (
	HtmlRoot  *Root = HTMLParseFromString(testHTML)
	HtmlRoot2 *Root = HTMLParseFromString(HtmlRoot2HTML)
)

func TestHTMLParsers(t *testing.T) {
	HtmlRoot1 := HTMLParseFromString(testHTML)
	require.Nil(t, HtmlRoot1.Error, "Variable of HTMLRoot must have A Nil Error")

	HtmlRoot1 = HTMLParse(strings.NewReader(HtmlRoot2HTML))
	require.Nil(t, HtmlRoot1.Error, "Variable of HTMLRoot must have A Nil Error")
}

func TestFind(t *testing.T) {
	require.Nil(t, HtmlRoot.Error)
	// Find() and Attrs()
	actual := HtmlRoot.Find("img")
	require.Nil(t, actual.Error)
	attrs := actual.Attrs()["src"]
	require.Equal(t, "images/springsource.png", attrs)

	// Find(...) and Text()
	actual = HtmlRoot.Find("a", "href", "hello")
	text := actual.Text()
	require.Equal(t, "servlet", text)

	// Nested Find()
	actual = HtmlRoot.Find("div").Find("div")
	text = actual.Text()
	require.Equal(t, "Just two divs peacing out", text)
	// Find("") for any
	actual = HtmlRoot2.Find("body")
	text = actual.Find("").Text()
	require.Equal(t, "Multiple classes", text)
	// Find("") with attributes
	actual = HtmlRoot.Find("", "id", "4")
	text = actual.Text()
	require.Equal(t, "Last one", text)
}

func TestFindError(t *testing.T) {
	actual := HtmlRoot.Find("footer")
	require.NotNil(t, actual.Error)

	actual = HtmlRoot.Find("img", "4", "id")
	require.NotNil(t, actual.Error)
}
func TestFindAll(t *testing.T) {
	// FindAll() and Attrs()
	allDivs := HtmlRoot.FindAll("div")
	//ForEach
	allDivs.ForEach(func(i int, r *Root) {
		id := r.Attrs()["id"]
		actual, _ := strconv.Atoi(id)
		require.Equal(t, i, actual)
	})
}

func TestFindAllError(t *testing.T) {
	allDivs := HtmlRoot.FindAll("div1")
	require.NotNil(t, allDivs.Error)
}

func TestFindAllBySingleClass(t *testing.T) {
	actual := HtmlRoot2.FindAll("div", "class", "first")
	require.Equal(t, 7, actual.Len)
	actual = HtmlRoot2.FindAll("div", "class", "third")
	require.Equal(t, 2, actual.Len)
}

func TestFindAllByAttribute(t *testing.T) {
	actual := HtmlRoot.FindAll("", "id", "2")
	require.Equal(t, 1, actual.Len)
}

func TestFindBySingleClass(t *testing.T) {
	actual := HtmlRoot2.Find("div", "class", "first")
	require.Equal(t, "Multiple classes", actual.Text())
	actual = HtmlRoot2.Find("div", "class", "third")
	require.Equal(t, "Multiple classes inorder", actual.Text())
}

func TestFindAllStrict(t *testing.T) {
	actual := HtmlRoot2.FindAllStrict("div", "class", "first second")
	require.Equal(t, 2, actual.Len)
	actual = HtmlRoot2.FindAllStrict("div", "class", "first third second")
	require.Equal(t, 0, actual.Len)
	actual = HtmlRoot2.FindAllStrict("div", "class", "second first third")
	require.Equal(t, 1, actual.Len)
}

func TestFindStrict(t *testing.T) {
	actual := HtmlRoot2.FindStrict("div", "class", "first")
	require.Equal(t, "Single class", actual.Text())
	actual = HtmlRoot2.FindStrict("div", "class", "third")
	require.NotNil(t, actual.Error)
}

func TestText(t *testing.T) {
	// <li>To a <a href="hello.jsp">JSP page</a> right?</li>
	li := HtmlRoot.Find("ul").Find("li")
	require.Equal(t, "To a ", li.Text())
}

func TestFullText(t *testing.T) {
	// <li>To a <a href="hello.jsp">JSP page</a> right?</li>
	li := HtmlRoot.Find("ul").Find("li")
	require.Equal(t, "To a JSP page right?", li.FullText())
}

func TestFullTextEmpty(t *testing.T) {
	// <div id="5"><h1><span></span></h1></div>
	h1 := HtmlRoot.Find("div", "id", "5").Find("h1")
	require.Empty(t, h1.FullText())
}

func TestNewErrorReturnsInspectableError(t *testing.T) {
	err := newError(ErrElementNotFound, errors.New("element not found"))
	require.NotNil(t, err)
	require.Equal(t, ErrElementNotFound, err.Type)
	require.Equal(t, "element not found", err.Err().Error())
}

// func TestFindReturnsInspectableError(t *testing.T) {
// 	r := HtmlRoot.Find("bogus", "thing")
// 	require.IsType(t, Error{}, r.Error)
// 	require.Equal(t, "element `bogus` with attributes `thing` not found", r.Error.Error())
// 	require.Equal(t, ErrElementNotFound, r.Error.(Error).Type)
// }
