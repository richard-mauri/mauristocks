package main

import (
 "bytes"
 "strings"
 "testing"

 suiteui "github.com/richard-mauri/homepage/shared/ui"
 "github.com/richard-mauri/homepage/shared/ui/banner"
)

func TestSharedBannerUsesViewportLayoutOutsideStockContent(t *testing.T) {
 for _,want:=range []string{
  `<header class="suite-ui suite-banner"`,
  `class="suite-type-title">Mauri&#39;s Stock Quotes`,
  `id="suite-auth-status"`,
  `suite-banner__portrait-caption">The Mauris`,
  `src="/tanner.jpg"`,
  `class="suite-ambient suite-ambient-one"`,
  `class="suite-ambient suite-ambient-two"`,
  `</header>`,
  `<main class="stocks-content">`,
  `class="suite-banner__home suite-type-control"`,
 } {if !strings.Contains(page,want) {t.Errorf("missing shared masthead %q",want)}}
 if !strings.Contains(page,`</header>`+"\n"+`<main class="stocks-content">`) {t.Error("banner should mount outside the stock content container")}
 if strings.Contains(page,`class="page-header"`) || strings.Contains(page,`https://raw.githubusercontent.com/`) {t.Error("legacy header or remote Tanner fetch retained")}
 if strings.Count(page,`id="suite-auth-status"`)!=1 {t.Error("duplicate status fields")}
 if !strings.Contains(page,`getElementById('suite-auth-status')`) {t.Error("existing verified identity flow must target shared banner")}
}

func TestSharedBannerCSSAndTannerRemainLocal(t *testing.T) {
 css:=string(banner.CSS(suiteui.DesignCSS))
 for _,want:=range []string{
  "width:min(1180px,calc(100vw - 96px))",
  "width:min(720px,calc(100vw - 48px))",
  "width:calc(100vw - 36px)",
  "blur(70px)",
 } {if !strings.Contains(page,want) || !strings.Contains(css,want) {t.Errorf("missing exact shared design %q",want)}}
 if strings.Contains(page,"radial-gradient(ellipse 660px") {t.Error("obsolete locally approximated glow retained")}
 if len(suiteTannerJPEG)<1000 || !bytes.HasPrefix(suiteTannerJPEG,[]byte{0xff,0xd8,0xff}) {t.Error("locally embedded Tanner JPEG missing")}
 if strings.Contains(page,"__SUITE_BANNER_") {t.Error("unexpanded banner placeholder")}
}
