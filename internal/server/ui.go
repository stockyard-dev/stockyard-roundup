package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Roundup</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#4a7ec9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.hdr-r{display:flex;gap:.5rem;align-items:center}
.main{padding:1rem}
.board{display:grid;grid-template-columns:repeat(3,1fr);gap:.8rem;min-height:70vh}
@media(max-width:700px){.board{grid-template-columns:1fr}}
.col{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem;min-height:200px}
.col-hdr{font-size:.65rem;text-transform:uppercase;letter-spacing:1px;margin-bottom:.8rem;display:flex;justify-content:space-between}
.col-todo .col-hdr{color:var(--cd)}.col-prog .col-hdr{color:var(--orange)}.col-done .col-hdr{color:var(--green)}
.task{background:var(--bg);border:1px solid var(--bg3);padding:.6rem .8rem;margin-bottom:.5rem;cursor:pointer;transition:border-color .15s;font-size:.75rem}
.task:hover{border-color:var(--leather)}
.task-title{color:var(--cream);margin-bottom:.2rem}
.task-meta{font-size:.6rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap}
.pri{font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.pri-urgent{background:#c9444433;color:var(--red);border:1px solid #c9444444}
.pri-high{background:#d4843a22;color:var(--orange);border:1px solid #d4843a44}
.pri-medium{background:var(--bg3);color:var(--cd);border:1px solid var(--bg3)}
.pri-low{background:var(--bg3);color:var(--cm);border:1px solid var(--bg3)}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}.btn-p:hover{opacity:.85}
.proj-bar{display:flex;gap:.3rem;padding:.5rem 1rem;border-bottom:1px solid var(--bg3);flex-wrap:wrap}
.proj-btn{font-size:.6rem;padding:.2rem .5rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}
.proj-btn:hover{border-color:var(--leather)}.proj-btn.active{border-color:var(--rust);color:var(--rust)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:400px;max-width:90vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.6rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .6rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.75rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-size:.7rem}
</style></head><body>
<div class="hdr"><h1>ROUNDUP</h1><div class="hdr-r"><span id="stats" style="font-size:.65rem;color:var(--cm)"></span><button class="btn btn-p" onclick="openTask()">+ Task</button><button class="btn" onclick="openProj()">+ Project</button></div></div>
<div class="proj-bar" id="projBar"></div>
<div class="main"><div class="board" id="board"></div></div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let projects=[],tasks=[],curProj='';
async function load(){
  const[p,st]=await Promise.all([fetch(A+'/projects').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
  projects=p.projects||[];
  document.getElementById('stats').textContent=st.todo+' todo · '+st.in_progress+' in progress · '+st.done+' done';
  renderProjs();loadTasks();
}
function renderProjs(){
  let h='<button class="proj-btn'+(curProj===''?' active':'')+'" onclick="filterProj(\'\')">All</button>';
  projects.forEach(p=>{h+='<button class="proj-btn'+(curProj===p.id?' active':'')+'" onclick="filterProj(\''+p.id+'\')"><span style="color:'+p.color+'">●</span> '+esc(p.name)+' ('+p.task_count+')</button>';});
  document.getElementById('projBar').innerHTML=h;
}
function filterProj(id){curProj=id;renderProjs();loadTasks();}
async function loadTasks(){
  let url=A+'/tasks?';if(curProj)url+='project='+curProj+'&';
  const r=await fetch(url).then(r=>r.json());tasks=r.tasks||[];renderBoard();
}
function renderBoard(){
  const todo=tasks.filter(t=>t.status==='todo'),prog=tasks.filter(t=>t.status==='in_progress'),done=tasks.filter(t=>t.status==='done');
  document.getElementById('board').innerHTML=
    '<div class="col col-todo"><div class="col-hdr"><span>TODO ('+todo.length+')</span></div>'+todo.map(taskCard).join('')+(todo.length?'':'<div class="empty">No tasks</div>')+'</div>'+
    '<div class="col col-prog"><div class="col-hdr"><span>IN PROGRESS ('+prog.length+')</span></div>'+prog.map(taskCard).join('')+(prog.length?'':'<div class="empty">No tasks</div>')+'</div>'+
    '<div class="col col-done"><div class="col-hdr"><span>DONE ('+done.length+')</span></div>'+done.map(taskCard).join('')+(done.length?'':'<div class="empty">No tasks</div>')+'</div>';
}
function taskCard(t){
  const next=t.status==='todo'?'in_progress':t.status==='in_progress'?'done':null;
  let h='<div class="task"><div style="display:flex;justify-content:space-between"><div class="task-title">'+esc(t.title)+'</div><span class="pri pri-'+t.priority+'">'+t.priority+'</span></div>';
  h+='<div class="task-meta">';
  if(t.assignee)h+='<span>@'+esc(t.assignee)+'</span>';
  if(t.due_date)h+='<span>due '+t.due_date+'</span>';
  if(t.tags&&t.tags.length)h+='<span>'+t.tags.join(', ')+'</span>';
  h+='</div>';
  h+='<div style="display:flex;gap:.3rem;margin-top:.4rem">';
  if(next)h+='<button class="btn" onclick="event.stopPropagation();mv(\''+t.id+'\',\''+next+'\')">'+(next==='in_progress'?'Start':'Complete')+'</button>';
  if(t.status==='done')h+='<button class="btn" onclick="event.stopPropagation();mv(\''+t.id+'\',\'todo\')">Reopen</button>';
  h+='<button class="btn" onclick="event.stopPropagation();del(\''+t.id+'\')" style="color:var(--red)">✕</button>';
  h+='</div></div>';return h;
}
async function mv(id,status){await fetch(A+'/tasks/'+id+'/status',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({status})});load();}
async function del(id){if(confirm('Delete?')){await fetch(A+'/tasks/'+id,{method:'DELETE'});load();}}
function openTask(){
  let opts=projects.map(p=>'<option value="'+p.id+'">'+esc(p.name)+'</option>').join('');
  document.getElementById('mdl').innerHTML='<h2>New Task</h2><div class="fr"><label>Title</label><input id="f-t"></div><div class="fr"><label>Description</label><textarea id="f-d" rows="2"></textarea></div><div class="fr"><label>Project</label><select id="f-p"><option value="">None</option>'+opts+'</select></div><div class="fr"><label>Priority</label><select id="f-pr"><option value="low">Low</option><option value="medium" selected>Medium</option><option value="high">High</option><option value="urgent">Urgent</option></select></div><div class="fr"><label>Assignee</label><input id="f-a"></div><div class="fr"><label>Due Date</label><input id="f-dd" type="date"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="subTask()">Create</button></div>';
  document.getElementById('mbg').classList.add('open');
}
async function subTask(){await fetch(A+'/tasks',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('f-t').value,description:document.getElementById('f-d').value,project_id:document.getElementById('f-p').value,priority:document.getElementById('f-pr').value,assignee:document.getElementById('f-a').value,due_date:document.getElementById('f-dd').value})});cm();load();}
function openProj(){
  document.getElementById('mdl').innerHTML='<h2>New Project</h2><div class="fr"><label>Name</label><input id="f-pn"></div><div class="fr"><label>Color</label><input id="f-pc" type="color" value="#c45d2c"></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="subProj()">Create</button></div>';
  document.getElementById('mbg').classList.add('open');
}
async function subProj(){await fetch(A+'/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('f-pn').value,color:document.getElementById('f-pc').value})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
