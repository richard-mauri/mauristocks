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
assert.ok(source.includes("suiteFeedback(hasUnsavedChanges()?'Unsaved changes · Click Save Changes.':'')"),'preference edits should show unsaved status');
assert.ok(source.includes("wlSave.disabled=settingsSaveBusy||!(committableSymbols||pendingSettings.size>0||accountSyncNeeded)"),'preference edits or failed account writes should enable Save Changes');
assert.ok(source.includes("await saveWatchlistToAccount()"),'explicit commit initiates account persistence');
assert.ok(!source.includes('function queueSuiteSave()'),'background autosave queue must be removed');
assert.ok(!source.includes('function queueWatchlistAccountSave()'),'old immediate watchlist background save must be removed');
assert.ok(!source.includes('function flushSuiteSave()'),'automatic save dispatcher must be removed');
assert.ok(source.includes('window.addEventListener(\'beforeunload\''),'unsaved changes should be protected on navigation');
// Extract the real stageSetting function and verify it never mutates persistent
// storage or starts a PUT. This is deliberately distinct from testing the UI.
const stageStart=source.indexOf('function stageSetting(key,value){');
const stageEnd=source.indexOf('function hasUnsavedChanges()',stageStart);
assert.ok(stageStart>=0&&stageEnd>stageStart);
const storage=new Map([['mauri_stock_refresh_interval','300']]);
const pendingSettings=new Map();
let feedback='',saveButtonUpdates=0;
const stageContext={localStorage:{getItem:key=>storage.has(key)?storage.get(key):null},
 pendingSettings,updateWatchlistButtons:()=>{saveButtonUpdates++},
 suiteFeedback:text=>{feedback=text},hasUnsavedChanges:()=>pendingSettings.size>0};
const stage=vm.runInNewContext(source.slice(stageStart,stageEnd)+'\n({stageSetting})',stageContext).stageSetting;
stage('mauri_stock_refresh_interval','30');
assert.equal(storage.get('mauri_stock_refresh_interval'),'300','frequency edit must not write browser storage');
assert.equal(pendingSettings.get('mauri_stock_refresh_interval'),'30');
assert.match(feedback,/Unsaved changes/);
stage('mauri_stock_refresh_interval','300');
assert.equal(pendingSettings.size,0,'reverting a change should disable Save Changes');
assert.equal(saveButtonUpdates,2);
assert.ok(!source.includes("suiteFeedback('');\n  void reconcileAccountOnSignIn(identity);"),'session check must not erase sync status');

// One global comparison header replaces repeated cell labels on desktop.
// Mobile retains labels in each stacked cell for unambiguous reading.
assert.match(source,/id="suite-dialog-column-headings" class="suite-diff-column-headings" hidden/);
assert.match(source,/<span>This browser<\/span><span>Saved account<\/span>/);
assert.match(source,/suiteDialogColumnHeadings.hidden=!diffRows;/);
assert.match(source,/\.suite-diff-column-headings\{display:grid;grid-template-columns:minmax\(0,1fr\) minmax\(0,1fr\);/);
assert.match(source,/\.suite-diff-value strong\{display:none;/);
assert.match(source,/@media\(max-width:550px\)\{[^\n]*\.suite-diff-column-headings\{display:none\}[^\n]*\.suite-diff-value strong\{display:block\}/);
assert.ok(source.indexOf('id="suite-dialog-column-headings"') < source.indexOf('id="suite-dialog-diff"'), 'shared headings are outside scroll area');
assert.ok(source.indexOf('id="suite-dialog-diff"') < source.indexOf('class="suite-dialog-actions"'), 'action buttons stay outside scroll area');
const sides=(source.match(/\['This browser',row\.browser\],\['Saved account',row\.account\]/g)||[]);
assert.equal(sides.length,1,'one consistent browser/account order for all diff cards');

// The dialog must keep actions outside the scrollable comparison region.
assert.match(source,/\.suite-dialog\[open\]\{display:flex\}/);
assert.match(source,/#suite-dialog-diff\{flex:0 1 auto;min-height:0;overflow-y:auto;/);
assert.match(source,/\.suite-dialog \.suite-dialog-actions\{padding-top:12px;/);
const helpersStart=source.indexOf('function suiteListEntries(raw){');
const helpersEnd=source.indexOf('function validSuiteSettings(data){',helpersStart);
assert.ok(helpersStart>0&&helpersEnd>helpersStart);
const wlKey='test_watchlists',activeKey='test_active';
const keys=[wlKey,activeKey,'mauri_stock_symbols','mauri_stock_refresh_interval','mauristocks_columns_v1'];
const diffContext={colSpecs,marketNames:{},flagStorageKey:'mauri_flags',dynamicName:'Most Active — Top 20 (Dynamic)',
 wlKey,activeKey,suiteKeys:keys,
 suiteSettingLabels:{[activeKey]:'Selected watchlist','mauri_stock_symbols':'Last saved symbols','mauri_stock_refresh_interval':'Auto refresh','mauristocks_columns_v1':'Visible columns'},
 normalizeSymbols:v=>[...new Set(String(v).toUpperCase().split(',').map(x=>x.trim()).filter(Boolean))].join(',')
};
const {suiteSymbolDelta,suiteSettingsDiff}=vm.runInNewContext(
 source.slice(helpersStart,helpersEnd)+'\n({suiteSymbolDelta,suiteSettingsDiff})',diffContext
);
const delta=suiteSymbolDelta('AAPL,GOOG,DRAM','AAPL,GOOG,TSLA');
assert.equal(delta.browser,'Added here (1): DRAM\nMissing here (1): TSLA');
assert.equal(delta.account,'Added here (1): TSLA\nMissing here (1): DRAM');
assert.equal(suiteSymbolDelta('AAPL,GOOG','AAPL,GOOG'),null);
assert.equal(suiteSymbolDelta('GOOG,AAPL','AAPL,GOOG').browser,'Order differs: GOOG, AAPL');
const list=(watchlists,active,mirror)=>({
 [activeKey]:active,[wlKey]:JSON.stringify(watchlists),
 mauri_stock_symbols:mirror
});
// Selected Default versus selected Candidates is NOT a watchlist content change.
const browserSettings=list({Candidates:'AAPL,GOOG'},'Default','MU,AMAT');
const accountSettings=list({Candidates:'AAPL,GOOG'},'Candidates','AAPL,GOOG');
const selected=suiteSettingsDiff(browserSettings,accountSettings);
assert.equal(selected.length,1,'legacy selected-symbol mirrors must not create a false symbol delta');
assert.equal(selected[0].label,'Selected watchlist');
assert.equal(selected[0].browser,'Default');
assert.equal(selected[0].account,'Candidates');
// Default is built-in and cannot be changed, even if stale saved JSON includes it.
const defaultOnly=suiteSettingsDiff(
 list({Default:'MU,AMAT',Candidates:'AAPL,GOOG'},'Candidates','AAPL,GOOG'),
 list({Default:'TSLA',Candidates:'AAPL,GOOG'},'Candidates','AAPL,GOOG')
);
assert.equal(defaultOnly.length,0,'Default never becomes a saved editable watchlist difference');
// Each named list is compared with itself only, independent of selected list.
const named=suiteSettingsDiff(
 list({Candidates:'AAPL,GOOG,DRAM',Growth:'NVDA,TSLA'},'Default','MU,AMAT'),
 list({Candidates:'AAPL,GOOG,TSLA',Growth:'NVDA,TSLA'},'Candidates','AAPL,GOOG,TSLA')
);
assert.equal(named.length,2);
const candidateChange=named.find(row=>row.label==='Watchlist: Candidates');
assert.ok(named.some(row=>row.label==='Selected watchlist'));
assert.ok(candidateChange,'the changed named watchlist must be compared');
assert.equal(candidateChange.browser,'Added here (1): DRAM\nMissing here (1): TSLA');
assert.equal(candidateChange.account,'Added here (1): TSLA\nMissing here (1): DRAM');
// An absent list can be newly created or deleted: show "Not present", never "Deleted".
const missing=suiteSettingsDiff(
 list({Candidates:'AAPL,GOOG',LocalOnly:'MU'},'Default','MU,AMAT'),
 list({Candidates:'AAPL,GOOG',RemoteOnly:'NVDA,TSLA'},'Default','MU,AMAT')
);
assert.equal(missing.length,2);
assert.equal(missing[0].label,'Watchlist: LocalOnly');
assert.equal(missing[0].browser,'Present (1 symbol)');
assert.equal(missing[0].account,null);
assert.equal(missing[1].label,'Watchlist: RemoteOnly');
assert.equal(missing[1].browser,null);
assert.equal(missing[1].account,'Present (2 symbols)');
// The legacy mirror must never appear, even if its stored copy is stale.
const mirrorOnly=suiteSettingsDiff(
 list({Candidates:'AAPL'},'Candidates','AAPL'),
 list({Candidates:'AAPL'},'Candidates','AAPL,GOOG')
);
assert.equal(mirrorOnly.length,0);
// Object key order is not meaningful and cannot trigger a spurious review.
const order=suiteSettingsDiff(
 list({Candidates:'AAPL',Growth:'NVDA'},'Candidates','AAPL'),
 list({Growth:'NVDA',Candidates:'AAPL'},'Candidates','AAPL')
);
assert.equal(order.length,0);
// Preserve readable, consistent side-by-side columns for visible column changes.
const columnRows=suiteSettingsDiff(
 {...browserSettings,'mauristocks_columns_v1':JSON.stringify({hidden:[]})},
 {...browserSettings,'mauristocks_columns_v1':JSON.stringify({hidden:['exchange','name']})}
);
assert.equal(columnRows.length,1);
assert.equal(columnRows[0].label,'Visible columns');
assert.equal(columnRows[0].browser,'Exchange: shown\nName: shown');
assert.equal(columnRows[0].account,'Exchange: hidden\nName: hidden');
assert.ok(source.includes("if(!differences.length){suiteRevision=existing.revision;"),'semantically equal settings should not prompt');
assert.ok(!source.includes("label:'Last saved symbols'"),'legacy mirror must not be shown as a card');
console.log('Named-watchlist identity, two-column diff, compact dialog and sync feedback checks passed.');
