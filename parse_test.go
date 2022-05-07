package taevas

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHTML(t *testing.T) {
	//spec := "<!DOCTYPE html><html lang=\"en\"><head><title>Swapping Songs</title></head><body><h1>Swapping Songs</h1><p>Tonight I swapped some of the songs I wrote with some friends, who gave me some of the songs they wrote. I love sharing my music.</p></body></html>"

	tt := []struct {
		name string
		in   string
	}{
		//{name: "spec", in: spec},
		{name: "paragraph open/close", in: "<p>testing testing</p>"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// doc, err := html.Parse(strings.NewReader(tc.in))
			// require.NoError(t, err)

			// var f func(*html.Node)
			// f = func(n *html.Node) {
			// 	t.Logf("type: %v", n.Type)
			// 	t.Logf("atom: %s", n.DataAtom)
			// 	t.Logf("data: %v", n.Data)
			// 	t.Logf("ns: %v", n.Namespace)
			// 	t.Logf("attr: %v", n.Attr)

			// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
			// 		f(c)
			// 	}
			// }
			// f(doc)

			z := html.NewTokenizer(strings.NewReader(tc.in))
			for {
				tag := z.Next()
				if tag == html.ErrorToken {
					return
				}
				t.Logf("token type: %s", tag)
				switch tag {
				case html.StartTagToken, html.EndTagToken:
					name, _ := z.TagName()
					t.Logf("token name: %s", name)
					for {
						k, v, more := z.TagAttr()
						t.Logf("%s = %s", k, v)
						if !more {
							break
						}
					}
				}

			}
		})
	}
}
