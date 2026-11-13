// Optional activity and verified Homepage identity. Handoff stays in this tab.
(() => {
  const tabKey='mauristocks-activity-tab-v1';
  const handoffKey='mauristocks.homepage.handoff.v1';
  let id='';
  try { id=sessionStorage.getItem(tabKey)||''; } catch (_) {}
  if(!/^[a-f0-9]{32}$/.test(id)) {
    id=Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,'0')).join('');
    try {sessionStorage.setItem(tabKey,id)} catch (_) {}
  }
  let bearer='';
  try {bearer=sessionStorage.getItem(handoffKey)||''} catch (_) {}
  if(!/^[a-f0-9]{64}$/.test(bearer))bearer='';
  const status=document.getElementById('suite-account-identity');
  const dialog=document.createElement('dialog');
  dialog.className='suite-signin-dialog';
  dialog.setAttribute('aria-labelledby','suite-signin-title');
  const heading=document.createElement('h2');heading.id='suite-signin-title';heading.textContent='Suite sign-in required';
  const explanation=document.createElement('p');
  explanation.textContent='Go to Homepage to sign in with your personal token or obtain a temporary authenticated guest session valid for one hour. Guest sessions use browser-only settings.';
  const link=document.createElement('a');link.href='https://home.themauris.org/';link.target='_blank';link.rel='noopener noreferrer';link.textContent='Go to Homepage';
  dialog.append(heading,explanation,link);
  dialog.addEventListener('cancel',event=>event.preventDefault());
  document.body.append(dialog);
  function verified(ok,explain=false){
    if(ok && dialog.open)dialog.close();
    if(!ok && explain && !dialog.open)dialog.showModal();
  }
  let checking=false,initialChecked=false,interacted=false,sentAt=0;
  async function checkIdentity(){
    if(checking)return;
    checking=true;
    try {
      bearer=sessionStorage.getItem(handoffKey)||'';
      if(!/^[a-f0-9]{64}$/.test(bearer))bearer='';
      if(!bearer){
        // Homepage's existing fixed launch bridge handles direct-visit recovery.
        if(location.hostname==='stocks.themauris.org' && !location.hash.includes('suite_mode=local')){
          verified(false,false);return;
        }
        status.textContent='Sign-in required';verified(false,true);return;
      }
      const response=await fetch('/api/activity/identity',{
        cache:'no-store',credentials:'same-origin',headers:{Authorization:'Bearer '+bearer}
      });
      if(response.status===503){status.textContent='Homepage verification unavailable';return}
      const result=response.ok?await response.json():{state:'guest'};
      if(result.state==='signed-in' && result.user){
        status.textContent='Signed in as '+result.user;
        verified(true);return;
      }
      const stale=bearer;
      try {if(sessionStorage.getItem(handoffKey)===stale)sessionStorage.removeItem(handoffKey)}catch(_){}
      bearer='';status.textContent='Session expired or revoked — Sign in again';verified(false,true);
    }catch(_){status.textContent='Homepage verification unavailable'}
    finally{checking=false}
  }
  function send(){
    if(!initialChecked || document.hidden)return;
    const now=Date.now();if(now-sentAt<25000)return;
    sentAt=now;
    const headers={'Content-Type':'application/json'};
    if(bearer)headers.Authorization='Bearer '+bearer;
    const usedBearer=bearer;
    const body=JSON.stringify({id,visible:true,interacted});interacted=false;
    fetch('/api/activity/ping',{method:'POST',cache:'no-store',credentials:'same-origin',
      headers,body,keepalive:true}).then(response=>{
      if(response.status!==401 || !usedBearer || bearer!==usedBearer)return;
      try{sessionStorage.removeItem(handoffKey)}catch(_){}
      bearer='';status.textContent='Session expired or revoked — Sign in again';verified(false,true);
    }).catch(()=>{});
  }
  function mark(){interacted=true;send()}
  document.addEventListener('pointerdown',mark,{passive:true});
  document.addEventListener('keydown',mark,{passive:true});
  document.addEventListener('visibilitychange',()=>{if(!document.hidden){void checkIdentity();send()}});
  window.addEventListener('pageshow',send);window.addEventListener('focus',()=>{void checkIdentity();send()});
  void checkIdentity().finally(()=>{initialChecked=true;send()});
  setInterval(()=>{void checkIdentity();send()},60000);
})();
