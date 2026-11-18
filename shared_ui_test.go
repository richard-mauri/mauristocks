package main

import (
	"strings"
	"testing"

	suiteui "github.com/richard-mauri/homepage/shared/ui"
)

func TestHomepageCanonicalButtons(t *testing.T) {
	if !strings.Contains(page, string(suiteui.ButtonCSS)) {
		t.Fatal("dashboard does not embed the pinned Homepage canonical button CSS")
	}
	if strings.Contains(page, "__SUITE_BUTTON_CSS__") {
		t.Fatal("unresolved shared CSS placeholder")
	}
	for _, id := range []string{"refresh", "watchlist-save", "watchlist-new", "watchlist-restore", "watchlist-delete", "suite-dialog-cancel", "suite-dialog-ok"} {
		if !strings.Contains(page, `<button class="ui-button" id="`+id+`"`) {
			t.Errorf("standard button %q does not use Homepage class", id)
		}
	}
	if !strings.Contains(page, "resetCols.className='ui-button'") {
		t.Error("dynamic Reset columns button does not use Homepage class")
	}
	for _, special := range []string{"b.className='sort-button'", "flag.className='symbol-flag'"} {
		if !strings.Contains(page, special) {
			t.Errorf("specialized control %q changed", special)
		}
	}
	for _, obsolete := range []string{
		"button{padding:.7rem 1.1rem", "button{background:#2868bd", "button{background:light-dark(",
		".watchlist-controls button{", ".suite-dialog button{", "button:focus-visible,a:focus-visible",
	} {
		if strings.Contains(pageTemplate, obsolete) {
			t.Errorf("dead local button CSS retained: %q", obsolete)
		}
	}
}
