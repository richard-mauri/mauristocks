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

// Only one visible save status appears, beside the Save Changes button.
assert.match(source,/<span id="watchlist-status" role="status" aria-live="polite"><\/span><\/div><p id="message"/);
assert.ok(!source.includes('id="suite-sync-status"'),'second account status must not exist');
assert.ok(!source.includes("getElementById('suite-sync-status')"),'account messages must use the shared status');
assert.ok(source.includes("function suiteFeedback(text){wlStatus.textContent=text;wlStatus.hidden=!text}"),'account and local save messages share one status line');
assert.ok(source.includes("suiteFeedback('Settings saved to browser and account.')"),'successful PUT must show combined confirmation');
assert.ok(source.includes("suiteFeedback('Settings saved to browser only. Sign in to save to Homepage.')"),'browser-only confirmation remains explicit');
assert.ok(source.includes("throw Error('Browser saved; account save failed (HTTP '+response.status+').')"),'failure distinguishes browser commit from failed account save');

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
// Execute the actual Save Changes body in an isolated environment. This
// checks local commit, a single explicit remote attempt, failed-remote retry,
// browser-only mode, protection of Default and localStorage rollback.
const saveStart=source.indexOf('async function commitSettings(){');
const saveEnd=source.indexOf('async function saveWatchlistToAccount(){',saveStart);
assert.ok(saveStart>0&&saveEnd>saveStart,'explicit commit source must exist');
const saveSource=source.slice(saveStart,saveEnd)+'\ncommitSettings';
function makeSaveHarness({signedIn=true,remoteOK=true,failKey=null}={}){
 const db=new Map([['mauri_stock_refresh_interval','300'],['mauristocks_columns_v1','{"hidden":[]}']]);
 let calls=0,lastFeedback='',buttonUpdates=0,allowRemote=remoteOK;
 const state={
  pendingSettings:new Map(),accountSyncNeeded:false,settingsSaveBusy:false,
  suiteSignedIn:signedIn,marketMode:'watchlist',activeWatchlist:'Candidates',
  dynamicName:'Most Active — Top 20 (Dynamic)',input:{value:'AAPL'},
  watchlists:{Candidates:'AAPL'},wlStatus:{textContent:''},
  watchlistChanged:()=>false,normalizeSymbols:v=>v,
  persistLists:()=>{},stageSetting:()=>{},
  updateWatchlistButtons:()=>{buttonUpdates++},
  hasUnsavedChanges:()=>state.pendingSettings.size>0,
  suiteFeedback:message=>{lastFeedback=message},
  localStorage:{
   getItem:key=>db.has(key)?db.get(key):null,
   setItem:(key,value)=>{if(key===failKey)throw Error('quota');db.set(key,value)},
   removeItem:key=>db.delete(key)
  },
  saveWatchlistToAccount:async()=>{calls++;return allowRemote}
 };
 const commit=vm.runInNewContext(saveSource,state);
 return {db,state,commit,calls:()=>calls,feedback:()=>lastFeedback,updates:()=>buttonUpdates,
  allowRemote:value=>{allowRemote=value}};
}
async function testExplicitSave(){
 const h=makeSaveHarness();
 h.state.pendingSettings.set('mauri_stock_refresh_interval','30');
 h.state.pendingSettings.set('mauristocks_columns_v1','{"hidden":["name"]}');
 assert.equal(h.db.get('mauri_stock_refresh_interval'),'300');
 assert.equal(h.calls(),0,'editing preferences must not trigger account storage');
 await h.commit();
 assert.equal(h.db.get('mauri_stock_refresh_interval'),'30');
 assert.equal(h.db.get('mauristocks_columns_v1'),'{"hidden":["name"]}');
 assert.equal(h.calls(),1,'one explicit save should start exactly one account attempt');
 assert.equal(h.state.pendingSettings.size,0);
 assert.equal(h.state.accountSyncNeeded,false);

 const failure=makeSaveHarness({remoteOK:false});
 failure.state.pendingSettings.set('mauri_stock_refresh_interval','60');
 await failure.commit();
 assert.equal(failure.db.get('mauri_stock_refresh_interval'),'60','remote failure must preserve browser save');
 assert.equal(failure.state.accountSyncNeeded,true,'remote failure must enable retry');
 assert.equal(failure.calls(),1);
 failure.allowRemote(true);
 await failure.commit();
 assert.equal(failure.calls(),2,'retry needs a second explicit button action');
 assert.equal(failure.state.accountSyncNeeded,false);

 const local=makeSaveHarness({signedIn:false});
 local.state.pendingSettings.set('mauri_stock_refresh_interval','5');
 await local.commit();
 assert.equal(local.db.get('mauri_stock_refresh_interval'),'5');
 assert.equal(local.calls(),0,'browser-only save must not send account requests');
 assert.match(local.feedback(),/Settings saved to browser only/);

 const protectedList=makeSaveHarness();
 protectedList.state.watchlistChanged=()=>true;
 protectedList.state.activeWatchlist='Default';
 await protectedList.commit();
 assert.equal(protectedList.calls(),0);
 assert.match(protectedList.feedback(),/protected/);

 const quota=makeSaveHarness({failKey:'mauristocks_columns_v1'});
 quota.state.pendingSettings.set('mauri_stock_refresh_interval','30');
 quota.state.pendingSettings.set('mauristocks_columns_v1','{"hidden":["exchange"]}');
 await quota.commit();
 assert.equal(quota.db.get('mauri_stock_refresh_interval'),'300','partial local write must roll back');
 assert.equal(quota.state.pendingSettings.size,2);
 assert.equal(quota.calls(),0,'browser storage failure must never trigger account sync');
}
testExplicitSave().then(()=>console.log('Explicit save, failed remote retry and browser rollback checks passed.'))
 .catch(error=>{console.error(error);process.exitCode=1});

