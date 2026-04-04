package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Roundup</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#4a7ec9;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.hdr-stats{font-family:var(--mono);font-size:.7rem;color:var(--cm)}
.toolbar{padding:.6rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.filter{font-family:var(--mono);font-size:.65rem;padding:.2rem .5rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}
.filter:hover{border-color:var(--leather)}.filter.active{border-color:var(--rust);color:var(--rust)}
.wrap{padding:1.5rem;max-width:900px;margin:0 auto}
.kanban{display:grid;grid-template-columns:repeat(3,1fr);gap:.8rem;margin-bottom:1rem}
@media(max-width:700px){.kanban{grid-template-columns:1fr}}
.col{background:var(--bg2);border:1px solid var(--bg3);min-height:200px}
.col-hdr{font-family:var(--mono);font-size:.65rem;text-transform:uppercase;letter-spacing:1px;padding:.6rem .8rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between}
.col-todo .col-hdr{color:var(--cd)}.col-progress .col-hdr{color:var(--blue)}.col-done .col-hdr{color:var(--green)}
.col-body{padding:.4rem}
.task{background:var(--bg);border:1px solid var(--bg3);padding:.6rem .7rem;margin-bottom:.4rem;cursor:pointer;transition:border-color .15s}
.task:hover{border-color:var(--leather)}
.task-title{font-family:var(--mono);font-size:.75rem;margin-bottom:.2rem}
.task-meta{font-family:var(--mono);font-size:.55rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap}
.pri{font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.pri-urgent{color:var(--red);border:1px solid #c9444444}.pri-high{color:var(--orange);border:1px solid #d4843a44}.pri-medium{color:var(--cd);border:1px solid var(--bg3)}.pri-low{color:var(--cm);border:1px solid var(--bg3)}
.tag{font-size:.5rem;padding:.1rem .25rem;background:var(--bg3);color:var(--cm)}
.btn{font-family:var(--mono);font-size:.65rem;padding:.3rem .7rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-primary{background:var(--rust);border-color:var(--rust);color:var(--bg)}.btn-primary:hover{opacity:.85}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:420px;max-width:90vw;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--mono);font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.6rem}.fr label{display:block;font-family:var(--mono);font-size:.6rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .6rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.78rem}
.fr textarea{min-height:50px;resize:vertical}
.actions{display:flex;gap:.5rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-size:.8rem}
.proj-bar{display:flex;gap:.3rem;margin-bottom:1rem;flex-wrap:wrap}
.proj-chip{font-family:var(--mono);font-size:.6rem;padding:.2rem .5rem;border:1px solid var(--bg3);cursor:pointer;color:var(--cm)}
.proj-chip:hover{border-color:var(--leather)}.proj-chip.active{border-color:var(--gold);color:var(--gold)}
</style></head><body>
<div class="hdr"><h1>ROUNDUP</h1><div class="hdr-stats" id="stats"></div></div>
<div class="toolbar">
<button class="btn btn-primary" onclick="openTaskForm()" style="font-size:.6rem">+ New Task</button>
<button class="btn" onclick="openProjForm()" style="font-size:.6rem">+ Project</button>
<span style="width:1px;height:16px;background:var(--bg3)"></span>
<span id="projBar"></span>
</div>
<div class="wrap">
<div class="kanban" id="kanban"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let tasks=[],projects=[],filterProj='';
async function load(){
const[t,p,s]=await Promise.all([fetch(A+'/tasks?status=all').then(r=>r.json()),fetch(A+'/projects').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);
tasks=t.tasks||[];projects=p.projects||[];
document.getElementById('stats').textContent=s.todo+' todo · '+s.in_progress+' active · '+s.done+' done';
renderProjects();render();}
function renderProjects(){let h='<span class="proj-chip'+(filterProj===''?' active':'')+'" onclick="setProj(\'\')">All</span>';
(projects||[]).forEach(p=>{h+='<span class="proj-chip'+(filterProj===p.id?' active':'')+'" onclick="setProj(\''+p.id+'\')" style="border-left:3px solid '+p.color+'">'+esc(p.name)+' ('+p.task_count+')</span>';});
document.getElementById('projBar').innerHTML=h;}
function setProj(id){filterProj=id;applyFilter();}
async function applyFilter(){let url=A+'/tasks?status=all';if(filterProj)url+='&project_id='+filterProj;
const r=await fetch(url).then(r=>r.json());tasks=r.tasks||[];render();}
function render(){const k=document.getElementById('kanban');
const todo=(tasks||[]).filter(t=>t.status==='todo');const prog=(tasks||[]).filter(t=>t.status==='in_progress');const done=(tasks||[]).filter(t=>t.status==='done');
k.innerHTML='<div class="col col-todo"><div class="col-hdr"><span>To Do</span><span>'+todo.length+'</span></div><div class="col-body">'+(todo.length?todo.map(taskCard).join(''):'<div class="empty">No tasks</div>')+'</div></div>'+
'<div class="col col-progress"><div class="col-hdr"><span>In Progress</span><span>'+prog.length+'</span></div><div class="col-body">'+(prog.length?prog.map(taskCard).join(''):'<div class="empty">No tasks</div>')+'</div></div>'+
'<div class="col col-done"><div class="col-hdr"><span>Done</span><span>'+done.length+'</span></div><div class="col-body">'+(done.length?done.slice(0,15).map(taskCard).join(''):'<div class="empty">No tasks</div>')+'</div></div>';}
function taskCard(t){let h='<div class="task" onclick="openEdit(\''+t.id+'\')"><div class="task-title">'+esc(t.title)+'</div><div class="task-meta">';
h+='<span class="pri pri-'+t.priority+'">'+t.priority+'</span>';
if(t.assignee)h+='<span>@'+esc(t.assignee)+'</span>';
if(t.due_date)h+='<span>due '+t.due_date+'</span>';
if(t.tags&&t.tags.length)t.tags.forEach(tg=>{h+='<span class="tag">'+esc(tg)+'</span>';});
h+='</div></div>';return h;}
function openTaskForm(){let opts=(projects||[]).map(p=>'<option value="'+p.id+'">'+esc(p.name)+'</option>').join('');
document.getElementById('mdl').innerHTML='<h2>New Task</h2><div class="fr"><label>Title</label><input id="f-title"></div><div class="fr"><label>Description</label><textarea id="f-desc"></textarea></div><div class="fr"><label>Project</label><select id="f-proj"><option value="">None</option>'+opts+'</select></div><div class="fr"><label>Priority</label><select id="f-pri"><option value="low">Low</option><option value="medium" selected>Medium</option><option value="high">High</option><option value="urgent">Urgent</option></select></div><div class="fr"><label>Assignee</label><input id="f-assign" placeholder="name"></div><div class="fr"><label>Due Date</label><input id="f-due" type="date"></div><div class="fr"><label>Tags (comma separated)</label><input id="f-tags"></div><div class="actions"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-primary" onclick="submitTask()">Create</button></div>';
document.getElementById('mbg').classList.add('open');}
async function submitTask(){const tags=(document.getElementById('f-tags').value||'').split(',').map(s=>s.trim()).filter(Boolean);
await fetch(A+'/tasks',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('f-title').value,description:document.getElementById('f-desc').value,project_id:document.getElementById('f-proj').value,priority:document.getElementById('f-pri').value,assignee:document.getElementById('f-assign').value,due_date:document.getElementById('f-due').value,tags})});cm();load();}
function openEdit(id){const t=(tasks||[]).find(t=>t.id===id);if(!t)return;
document.getElementById('mdl').innerHTML='<h2>Edit Task</h2><div class="fr"><label>Title</label><input id="e-title" value="'+esc(t.title)+'"></div><div class="fr"><label>Status</label><select id="e-status"><option value="todo"'+(t.status==='todo'?' selected':'')+'>To Do</option><option value="in_progress"'+(t.status==='in_progress'?' selected':'')+'>In Progress</option><option value="done"'+(t.status==='done'?' selected':'')+'>Done</option></select></div><div class="fr"><label>Priority</label><select id="e-pri"><option value="low"'+(t.priority==='low'?' selected':'')+'>Low</option><option value="medium"'+(t.priority==='medium'?' selected':'')+'>Medium</option><option value="high"'+(t.priority==='high'?' selected':'')+'>High</option><option value="urgent"'+(t.priority==='urgent'?' selected':'')+'>Urgent</option></select></div><div class="fr"><label>Assignee</label><input id="e-assign" value="'+(t.assignee||'')+'"></div><div class="actions"><button class="btn" style="color:var(--red)" onclick="delTask(\''+id+'\')">Delete</button><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-primary" onclick="saveTask(\''+id+'\')">Save</button></div>';
document.getElementById('mbg').classList.add('open');}
async function saveTask(id){await fetch(A+'/tasks/'+id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('e-title').value,status:document.getElementById('e-status').value,priority:document.getElementById('e-pri').value,assignee:document.getElementById('e-assign').value})});cm();load();}
async function delTask(id){if(confirm('Delete?')){await fetch(A+'/tasks/'+id,{method:'DELETE'});cm();load();}}
function openProjForm(){document.getElementById('mdl').innerHTML='<h2>New Project</h2><div class="fr"><label>Name</label><input id="p-name"></div><div class="fr"><label>Color</label><input id="p-color" type="color" value="#c45d2c"></div><div class="actions"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-primary" onclick="submitProj()">Create</button></div>';document.getElementById('mbg').classList.add('open');}
async function submitProj(){await fetch(A+'/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('p-name').value,color:document.getElementById('p-color').value})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
