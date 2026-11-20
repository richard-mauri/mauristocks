// Regression cases for Issue #33 browser/account comparison.
// Run with: node scripts/test-account-diff.cjs
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const source=fs.readFileSync('main.go','utf8');
const start=source.indexOf('function suiteHiddenColumns(raw){');
const end=source.indexOf('function suiteSettingsDiff(browser,account){',start);
assert.ok(start>0&&end>start,'account-diff helper functions must be present');
const colSpecs=[['symbol','Symbol'],['exchange','Exchange'],['name','Name'],['display_price','Price'],['extended_price','Extended price'],['change','Change $'],['change_percent','Change %'],['day_position','Day Range'],['week_52_position','52-Week Range']];
const context={colSpecs,marketNames:{},flagStorageKey:'mauri_flags'};
const {suiteHiddenColumns,suiteColumnDelta,suiteReadableSetting}=vm.runInNewContext(
 source.slice(start,end)+'\n({suiteHiddenColumns,suiteColumnDelta,suiteReadableSetting})',context
);
const saved=JSON.stringify({hidden:['exchange']});
const browser=JSON.stringify({hidden:['exchange','name']});
const differences=suiteColumnDelta(browser,saved);
assert.equal(differences.browser,'Name: hidden');
assert.equal(differences.account,'Name: shown');
assert.equal(suiteReadableSetting('mauristocks_columns_v1',browser),'Hidden: Exchange, Name');
assert.equal(suiteReadableSetting('mauristocks_columns_v1',JSON.stringify({hidden:[]})),'All columns shown');
assert.equal(suiteColumnDelta(JSON.stringify({hidden:['name','exchange']}),browser),null,'hidden-array order does not affect visibility');
assert.equal(suiteColumnDelta('bad json',saved),null,'legacy malformed settings do not crash diff');
assert.equal(suiteHiddenColumns(JSON.stringify({hidden:['symbol','name','name','obsolete']})).join(','),'name','symbol must remain visible and obsolete entries ignored');
assert.ok(source.includes("suiteFeedback('Preference saved in this browser. Saving to your account…')"),'preference sync should expose pending feedback');
assert.ok(!source.includes("suiteFeedback('');\n  void reconcileAccountOnSignIn(identity);"),'session check must not erase sync status');
console.log('Account difference and status regression checks passed.');
