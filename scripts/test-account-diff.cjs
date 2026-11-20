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

// The dialog must keep actions outside the scrollable comparison region.
assert.match(source,/\.suite-dialog\[open\]\{display:flex\}/);
assert.match(source,/#suite-dialog-diff\{flex:0 1 auto;min-height:0;overflow-y:auto;/);
assert.match(source,/\.suite-dialog \.suite-dialog-actions\{padding-top:12px;/);
const helpersStart=source.indexOf('function suiteListEntries(raw){');
const helpersEnd=source.indexOf('function validSuiteSettings(data){',helpersStart);
assert.ok(helpersStart>0&&helpersEnd>helpersStart);
const wlKey='test_watchlists',activeKey='test_active';
const keys=[wlKey,activeKey,'mauri_stock_symbols','mauri_stock_refresh_interval','mauristocks_columns_v1'];
const diffContext={colSpecs,marketNames:{},flagStorageKey:'mauri_flags',
 wlKey,activeKey,suiteKeys:keys,
 suiteSettingLabels:{[activeKey]:'Selected watchlist','mauri_stock_symbols':'Last saved symbols','mauri_stock_refresh_interval':'Auto refresh','mauristocks_columns_v1':'Visible columns'},
 normalizeSymbols:v=>[...new Set(String(v).toUpperCase().split(',').map(x=>x.trim()).filter(Boolean))].join(',')
};
const {suiteSymbolDelta,suiteSettingsDiff}=vm.runInNewContext(
 source.slice(helpersStart,helpersEnd)+'\n({suiteSymbolDelta,suiteSettingsDiff})',diffContext
);
const delta=suiteSymbolDelta('AAPL,GOOG,DRAM','AAPL,GOOG,TSLA');
assert.equal(JSON.stringify(delta.changes),JSON.stringify([
 {label:'Added in this browser (1)',symbols:['DRAM']},
 {label:'Removed from this browser (1)',symbols:['TSLA']}
]));
const browserSettings={
 [activeKey]:'Default',[wlKey]:JSON.stringify({Default:'MU,AMAT',Candidates:'AAPL,GOOG'}),
 mauri_stock_symbols:'MU,AMAT'
};
const accountSettings={
 [activeKey]:'Candidates',[wlKey]:JSON.stringify({Default:'MU,AMAT',Candidates:'AAPL,GOOG'}),
 mauri_stock_symbols:'AAPL,GOOG'
};
const rows=suiteSettingsDiff(browserSettings,accountSettings);
assert.equal(rows.length,1,'do not compare unrelated selected watchlists as added/removed symbols');
assert.equal(rows[0].label,'Selected watchlist');
const diffLists=suiteSettingsDiff({
 [activeKey]:'Candidates',[wlKey]:JSON.stringify({Candidates:'AAPL,GOOG,DRAM'}),mauri_stock_symbols:'AAPL,GOOG,DRAM'
},{
 [activeKey]:'Candidates',[wlKey]:JSON.stringify({Candidates:'AAPL,GOOG'}),mauri_stock_symbols:'AAPL,GOOG'
});
assert.equal(diffLists.length,1,'do not repeat last-saved mirror of named watchlist difference');
assert.equal(diffLists[0].label,'Watchlist: Candidates');
assert.equal(diffLists[0].changes[0].symbols.join(','),'DRAM');

console.log('Account difference, compact dialog and status regression checks passed.');
