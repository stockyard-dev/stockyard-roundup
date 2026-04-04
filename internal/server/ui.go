package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Roundup</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#4a7ec9;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.hdr-stats{font-family:var(--mono);font-size:.7rem;color:var(--cm)}
.board{display:grid;grid-template-columns:repeat(3,1fr);gap:1rem;padding:1.5rem;min-height:calc(100vh - 60px)}
@media(max-width:700px){.board{grid-template-columns:1fr}}
.col{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem;min-height:200px}
.col-header{font-family:var(--mono);font-size:.7rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.8rem;display:flex;justify-content:space-between;align-items:center}
.col-count{background:var(--bg3);padding:.1rem .4rem;font-size:.6rem;color:var(--cm)}
.task{background:var(--bg);border:1px solid var(--bg3);padding:.7rem;margin-bottom:.5rem;cursor:pointer;transition:border-color .15s}
.task:hover{border-color:var(--leather)}
.task-title{font-family:var(--mono);font-size:.78rem;margin-bottom:.2rem}
.task-meta{font-family:var(--mono);font-size:.58rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap}
.pri{font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.pri-urgent{background:#c9444433;color:var(--red);border:1px solid #c9444444}
.pri-high{background:#d4843a22;color:var(--orange);border:1px solid #d4843a44}
.pri-medium{background:var(--bg3);color:var(--cm);border:1px solid var(--bg3)}
.pri-low{background:#4a7ec922;color:var(--blue);border:1px solid #4a7ec944}
.tag{font-size:.5rem;background:var(--bg3);color:var(--cm);padding:.1rem .3rem}
.toolbar{padding:.5rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;gap:.5rem;justify-content:space-between;flex-wrap:wrap}
.btn{font-family:var(--mono);font-size:.65rem;padding:.3rem .7rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}.btn-p:hover{opacity:.85}
.filter-btn{font-size:.6rem;padding:.2rem .5rem}.filter-btn.active{border-color:var(--rust);color:var(--rust)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:420px;max-width:90vw}
.modal h2{font-family:var(--mono);font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.6rem}.fr label{display:block;font-family:var(--mono);font-size:.6rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .6rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.78rem}
.fr textarea{min-height:60px;resize:vertical}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.status-btns{display:flex;gap:.3rem;margin-top:.5rem}
</style></head><body>
<div class="hdr"><h1>ROUNDUP</h1><div class="hdr-stats" id="st"></div></div>
<div class="toolbar"><div id="projFilter"></div><button class="btn btn-p" onclick="openTask()">+ New Task</button></div>
<div class="board" id="board"></div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let tasks=[],projects=[],projFilter='';
async function ld(){const[t,p,s]=await Promise.all([fetch(A+'/tasks'+(projFilter?'?project='+projFilter:'')).then(r=>r.json()),fetch(A+'/projects').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
tasks=t.tasks||[];projects=p.projects||[];document.getElementById('st').textContent=s.tasks+' tasks, '+s.projects+' projects';rn();}
function rn(){
  // Project filter
  let pf='<button class="btn filter-btn'+(projFilter===''?' active':'')+'" onclick="setProj(\'\')">All</button>';
  (projects||[]).forEach(p=>{pf+='<button class="btn filter-btn'+(projFilter===p.id?' active':'')+'" onclick="setProj(\''+p.id+'\')">'+esc(p.name)+'</button>';});
  pf+='<button class="btn" onclick="openProj()" style="font-size:.55rem">+ Project</button>';
  document.getElementById('projFilter').innerHTML=pf;
  // Board
  const cols={todo:tasks.filter(t=>t.status==='todo'),in_progress:tasks.filter(t=>t.status==='in_progress'),done:tasks.filter(t=>t.status==='done')};
  let h='';
  [{k:'todo',l:'To Do'},{k:'in_progress',l:'In Progress'},{k:'done',l:'Done'}].forEach(c=>{
    const items=cols[c.k]||[];
    h+='<div class="col"><div class="col-header"><span>'+c.l+'</span><span class="col-count">'+items.length+'</span></div>';
    items.forEach(t=>{
      h+='<div class="task" onclick="openDetail(\''+t.id+'\')"><div class="task-title">'+esc(t.title)+'</div><div class="task-meta"><span class="pri pri-'+t.priority+'">'+t.priority+'</span>';
      if(t.assignee)h+='<span>@'+esc(t.assignee)+'</span>';
      if(t.due_date)h+='<span>due '+t.due_date+'</span>';
      (t.tags||[]).forEach(tg=>{h+='<span class="tag">'+esc(tg)+'</span>';});
      h+='</div></div>';
    });
    h+='</div>';
  });
  document.getElementById('board').innerHTML=h;
}
async function setProj(id){projFilter=id;const r=await fetch(A+'/tasks'+(id?'?project='+id:'')).then(r=>r.json());tasks=r.tasks||[];rn();}
function openTask(){
  let opts=(projects||[]).map(p=>'<option value="'+p.id+'">'+esc(p.name)+'</option>').join('');
  document.getElementById('mdl').innerHTML='<h2>New Task</h2><div class="fr"><label>Title</label><input id="tt"></div><div class="fr"><label>Description</label><textarea id="td"></textarea></div><div class="fr"><label>Project</label><select id="tp"><option value="">None</option>'+opts+'</select></div><div class="fr"><label>Priority</label><select id="tr"><option value="low">Low</option><option value="medium" selected>Medium</option><option value="high">High</option><option value="urgent">Urgent</option></select></div><div class="fr"><label>Assignee</label><input id="ta" placeholder="e.g. michael"></div><div class="fr"><label>Due Date</label><input id="tdu" type="date"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="sTask()">Create</button></div>';
  document.getElementById('mbg').classList.add('open');
}
async function sTask(){
  await fetch(A+'/tasks',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('tt').value,description:document.getElementById('td').value,project_id:document.getElementById('tp').value,priority:document.getElementById('tr').value,assignee:document.getElementById('ta').value,due_date:document.getElementById('tdu').value})});cm();ld();
}
function openDetail(id){
  const t=tasks.find(x=>x.id===id);if(!t)return;
  document.getElementById('mdl').innerHTML='<h2>'+esc(t.title)+'</h2>'+(t.description?'<p style="font-size:.82rem;color:var(--cd);margin-bottom:.8rem">'+esc(t.description)+'</p>':'')+'<div class="task-meta" style="margin-bottom:.8rem"><span class="pri pri-'+t.priority+'">'+t.priority+'</span>'+(t.assignee?'<span>@'+esc(t.assignee)+'</span>':'')+(t.due_date?'<span>due '+t.due_date+'</span>':'')+'</div><div class="status-btns"><button class="btn'+(t.status==='todo'?' btn-p':'')+'" onclick="ss(\''+id+'\',\'todo\')">Todo</button><button class="btn'+(t.status==='in_progress'?' btn-p':'')+'" onclick="ss(\''+id+'\',\'in_progress\')">In Progress</button><button class="btn'+(t.status==='done'?' btn-p':'')+'" onclick="ss(\''+id+'\',\'done\')">Done</button></div><div class="acts" style="margin-top:1rem"><button class="btn" style="color:var(--red)" onclick="dt(\''+id+'\')">Delete</button><button class="btn" onclick="cm()">Close</button></div>';
  document.getElementById('mbg').classList.add('open');
}
async function ss(id,status){await fetch(A+'/tasks/'+id+'/status',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({status})});cm();ld();}
async function dt(id){if(confirm('Delete task?')){await fetch(A+'/tasks/'+id,{method:'DELETE'});cm();ld();}}
function openProj(){document.getElementById('mdl').innerHTML='<h2>New Project</h2><div class="fr"><label>Name</label><input id="pn"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="sProj()">Create</button></div>';document.getElementById('mbg').classList.add('open');}
async function sProj(){await fetch(A+'/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('pn').value})});cm();ld();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
ld();
</script></body></html>`
