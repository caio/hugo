// Copyright 2016 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"fmt"
	"path/filepath"
	"testing"

	"reflect"

	"github.com/stretchr/testify/require"

	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/tpl"
)

const sitemapTemplate = `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  {{ range .Data.Pages }}
  <url>
    <loc>{{ .Permalink }}</loc>{{ if not .Lastmod.IsZero }}
    <lastmod>{{ safeHTML ( .Lastmod.Format "2006-01-02T15:04:05-07:00" ) }}</lastmod>{{ end }}{{ with .Sitemap.ChangeFreq }}
    <changefreq>{{ . }}</changefreq>{{ end }}{{ if ge .Sitemap.Priority 0.0 }}
    <priority>{{ .Sitemap.Priority }}</priority>{{ end }}
  </url>
  {{ end }}
</urlset>`

func TestSitemapOutput(t *testing.T) {
	t.Parallel()
	for _, internal := range []bool{false, true} {
		doTestSitemapOutput(t, internal)
	}
}

func TestSitemapExtraEntries(t *testing.T) {
	t.Parallel()
	cfg, fs := newTestCfg()

	extraEntries := []string{"bar", "baz", "cv/hue.pdf"}
	baseURL := "http://auth/bub"
	cfg.Set("baseURL", baseURL)
	cfg.Set("sitemap", map[string]interface{}{"extra_entries": extraEntries})

	depsCfg := deps.DepsCfg{Fs: fs, Cfg: cfg}

	sources := [][2]string{
		{filepath.FromSlash("foo.html"), "---\nmarkup: markdown\n---\n# title\nsome *content*"},
	}
	writeSourcesToSource(t, "content", fs, sources...)

	s := buildSingleSite(t, depsCfg, BuildCfg{})
	th := testHelper{s.Cfg, s.Fs, t}

	sitemap := readDestination(th.T, th.Fs, "public/sitemap.xml")
	for _, extra := range extraEntries {
		require.Contains(t, sitemap, fmt.Sprintf("<loc>%s/%s</loc>", baseURL, extra))
	}
}

func TestNoEmptyTaxonomy(t *testing.T) {
	t.Parallel()
	cfg, fs := newTestCfg()

	depsCfg := deps.DepsCfg{Fs: fs, Cfg: cfg}

	sources := [][2]string{
		{filepath.FromSlash("foo.html"), "---\nmarkup: markdown\n---\n# title\nsome *content*"},
	}
	writeSourcesToSource(t, "content", fs, sources...)

	s := buildSingleSite(t, depsCfg, BuildCfg{})
	th := testHelper{s.Cfg, s.Fs, t}

	sitemap := readDestination(th.T, th.Fs, "public/sitemap.xml")
	for _, page := range []string{"categories", "tags"} {
		if destinationExists(fs, fmt.Sprintf("public/%s/index.html", page)) {
			t.Errorf("Building with %s-less sources created a /%s/ page", page, page)
		}
		require.NotContains(t, sitemap, page)
	}
}

func doTestSitemapOutput(t *testing.T, internal bool) {

	cfg, fs := newTestCfg()
	cfg.Set("baseURL", "http://auth/bub/")

	depsCfg := deps.DepsCfg{Fs: fs, Cfg: cfg}

	depsCfg.WithTemplate = func(templ tpl.TemplateHandler) error {
		if !internal {
			templ.AddTemplate("sitemap.xml", sitemapTemplate)
		}

		// We want to check that the 404 page is not included in the sitemap
		// output. This template should have no effect either way, but include
		// it for the clarity.
		templ.AddTemplate("404.html", "Not found")
		return nil
	}

	writeSourcesToSource(t, "content", fs, weightedSources...)
	s := buildSingleSite(t, depsCfg, BuildCfg{})
	th := testHelper{s.Cfg, s.Fs, t}
	outputSitemap := "public/sitemap.xml"

	th.assertFileContent(outputSitemap,
		// Regular page
		" <loc>http://auth/bub/sect/doc1/</loc>",
		// Home page
		"<loc>http://auth/bub/</loc>",
		// Section
		"<loc>http://auth/bub/sect/</loc>",
		// Tax terms
		"<loc>http://auth/bub/categories/</loc>",
		// Tax list
		"<loc>http://auth/bub/categories/hugo/</loc>",
	)

	content := readDestination(th.T, th.Fs, outputSitemap)
	require.NotContains(t, content, "404")

}

func TestParseSitemap(t *testing.T) {
	t.Parallel()
	expected := Sitemap{Priority: 3.0, Filename: "doo.xml", ChangeFreq: "3"}
	input := map[string]interface{}{
		"changefreq": "3",
		"priority":   3.0,
		"filename":   "doo.xml",
		"unknown":    "ignore",
	}
	result := parseSitemap(input)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Got \n%v expected \n%v", result, expected)
	}

}
