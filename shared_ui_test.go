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

func TestConfirmationDialogDoesNotAutofocusOK(t *testing.T) {
	if strings.Contains(pageTemplate, "else suiteDialogOK.focus()") {
		t.Fatal("confirm dialog must not highlight OK before keyboard navigation")
	}
	for _, required := range []string{
		`class="suite-dialog" tabindex="-1"`,
		"else suiteDialog.focus()",
		"suiteDialog.addEventListener('keydown',enter)",
		"suiteDialog.removeEventListener('keydown',enter)",
	} {
		if !strings.Contains(pageTemplate, required) {
			t.Errorf("missing dialog focus/keyboard contract %q", required)
		}
	}
}

func TestObsoletePaletteAndWatchlistCSSRemoved(t *testing.T) {
	for _, obsolete := range []string{
		".watchlist-controls select{",
		"color:#1763b1",
		".up{color:#168243}",
		".down{color:#d54848}",
		".symbol-flag[aria-pressed=\"true\"]{color:#946900}",
		".symbol-flag:hover,.symbol-flag:focus-visible{background:#ffe38b",
		".range-marker{background:#2563a6",
	} {
		if strings.Contains(pageTemplate, obsolete) {
			t.Errorf("obsolete pre-palette CSS remains: %q", obsolete)
		}
	}
	for _, required := range []string{
		".sort-button{border:0;",
		".symbol-flag{background:transparent;",
		".range-marker{background:var(--accent);",
		".up{color:var(--gain)}.down{color:var(--loss)}",
	} {
		if !strings.Contains(pageTemplate, required) {
			t.Errorf("functional semantic rule missing: %q", required)
		}
	}
}
