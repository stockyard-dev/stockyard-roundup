package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Roundup</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#4a7ec9;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.header{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.header h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.content{padding:1.5rem}
.stats-row{display:grid;grid-template-columns:repeat(auto-fit,minmax(100px,1fr));gap:.6rem;margin-bottom:1rem;max-width:600px}
.stat{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.stat-val{font-family:var(--mono);font-size:1.2rem}.stat-label{font-family:var(--mono);font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;flex-wrap:wrap;align-items:center}
.toolbar select,.toolbar input{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem;padding:.3rem .5rem}
.btn{font-family:var(--mono);font-size:.65rem;padding:.3rem .7rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-primary{background:var(--rust);border-color:var(--rust);color:var(--bg)}.btn-primary:hover{opacity:.85}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.kanban{display:grid;grid-template-columns:1fr 1fr 1fr;gap:.8rem}
@media(max-width:700px){.kanban{grid-template-columns:1fr}}
.col{background:var(--bg2);border:1px solid var(--bg3);min-height:200px}
.col-header{font-family:var(--mono);font-size:.65rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;padding:.6rem .8rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between}
.col-count{background:var(--bg3);padding:0 .4rem;border-radius:2px;font-size:.6rem}
.task{border-bottom:1px solid var(--bg3);padding:.6rem .8rem;cursor:pointer;transition:background .1s}
.task:hover{background:var(--bg)}
.task-title{font-family:var(--mono);font-size:.75rem;margin-bottom:.2rem}
.task-meta{font-family:var(--mono);font-size:.55rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap}
.pri{font-size:.5rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px}
.pri-urgent{background:#c9444433;color:var(--red)}.pri-high{background:#d4843a22;color:var(--orange)}.pri-medium{background:var(--bg3);color:var(--cm)}.pri-low{background:var(--bg3);color:var(--cm)}
.tag{font-size:.5rem;padding:.1rem .3rem;background:var(--bg3);color:var(--cm)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:450px;max-width:90vw;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--mono);font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.form-row{margin-bottom:.6rem}
.form-row label{display:block;font-family:var(--mono);font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.form-row input,.form-row select,.form-row textarea{width:100%;padding:.4rem .6rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.75rem}
.form-row textarea{min-height:50px;resize:vertical}
.actions{display:flex;gap:.5rem;justify-content:flex-end;margin-top:1rem}
.empty-col{padding:1rem;text-align:center;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="header"><h1>ROUNDUP</h1><div style="display:flex;gap:.5rem"><button class="btn btn-primary" onclick="openTaskForm()">+ Task</button><button class="btn" onclick="openProjectForm()">+ Project</button></div></div>
<div class="content">
<div class="stats-row" id="statsRow"></div>
<div class="toolbar" id="toolbar"></div>
<div class="kanban" id="kanban"></div>
</div>
<div class="modal-bg" id="modalBg" onclick="if(event.target===this)closeModal()"><div class="modal" id="modal"></div></div>

<script>
const API='/api';let tasks=[],projects=[],stats={},filterProject='',filterPriority='';

async function load(){
  const[t,p,s]=await Promise.all([fetch(API+'/tasks').then(r=>r.json()),fetch(API+'/projects').then(r=>r.json()),fetch(API+'/stats').then(r=>r.json())]);
  tasks=t.tasks||[];projects=p.projects||[];stats=s;
  renderStats();renderToolbar();renderKanban();
}

function renderStats(){
  document.getElementById('statsRow').innerHTML='<div class="stat"><div class="stat-val">'+stats.todo+'</div><div class="stat-label">To Do</div></div><div class="stat"><div class="stat-val">'+stats.in_progress+'</div><div class="stat-label">In Progress</div></div><div class="stat"><div class="stat-val">'+stats.done+'</div><div class="stat-label">Done</div></div><div class="stat"><div class="stat-val">'+stats.projects+'</div><div class="stat-label">Projects</div></div>';
}

function renderToolbar(){
  let h='<select onchange="filterProject=this.value;applyFilters()"><option value="">All projects</option>';
  (projects||[]).forEach(p=>{h+='<option value="'+p.id+'">'+esc(p.name)+' ('+p.task_count+')</option>';});
  h+='</select><select onchange="filterPriority=this.value;applyFilters()"><option value="">All priorities</option><option value="urgent">Urgent</option><option value="high">High</option><option value="medium">Medium</option><option value="low">Low</option></select>';
  document.getElementById('toolbar').innerHTML=h;
}

async function applyFilters(){
  let url=API+'/tasks?';
  if(filterProject)url+='project_id='+filterProject+'&';
  if(filterPriority)url+='priority='+filterPriority+'&';
  const r=await fetch(url).then(r=>r.json());
  tasks=r.tasks||[];renderKanban();
}

function renderKanban(){
  const todo=(tasks||[]).filter(t=>t.status==='todo');
  const wip=(tasks||[]).filter(t=>t.status==='in_progress');
  const done=(tasks||[]).filter(t=>t.status==='done');
  let h='';
  [['To Do','todo',todo],['In Progress','in_progress',wip],['Done','done',done]].forEach(([label,status,items])=>{
    h+='<div class="col"><div class="col-header">'+label+'<span class="col-count">'+items.length+'</span></div>';
    if(!items.length)h+='<div class="empty-col">No tasks</div>';
    items.forEach(t=>{
      h+='<div class="task" onclick="openEditForm(\''+t.id+'\')"><div class="task-title">'+esc(t.title)+'</div><div class="task-meta">';
      h+='<span class="pri pri-'+t.priority+'">'+t.priority+'</span>';
      if(t.assignee)h+='<span>'+esc(t.assignee)+'</span>';
      if(t.due_date)h+='<span>due '+t.due_date+'</span>';
      (t.tags||[]).forEach(tag=>{if(tag)h+='<span class="tag">'+esc(tag)+'</span>';});
      h+='</div></div>';
    });
    h+='</div>';
  });
  document.getElementById('kanban').innerHTML=h;
}

function openTaskForm(){
  let opts=(projects||[]).map(p=>'<option value="'+p.id+'">'+esc(p.name)+'</option>').join('');
  document.getElementById('modal').innerHTML='<h2>New Task</h2><div class="form-row"><label>Title</label><input id="f-title"></div><div class="form-row"><label>Description</label><textarea id="f-desc"></textarea></div><div class="form-row"><label>Project</label><select id="f-proj"><option value="">None</option>'+opts+'</select></div><div class="form-row"><label>Priority</label><select id="f-pri"><option value="medium">Medium</option><option value="urgent">Urgent</option><option value="high">High</option><option value="low">Low</option></select></div><div class="form-row"><label>Assignee</label><input id="f-assign"></div><div class="form-row"><label>Due date</label><input id="f-due" type="date"></div><div class="actions"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="submitTask()">Create</button></div>';
  document.getElementById('modalBg').classList.add('open');
}

async function submitTask(){
  await fetch(API+'/tasks',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('f-title').value,description:document.getElementById('f-desc').value,project_id:document.getElementById('f-proj').value,priority:document.getElementById('f-pri').value,assignee:document.getElementById('f-assign').value,due_date:document.getElementById('f-due').value})});
  closeModal();load();
}

function openEditForm(id){
  const t=(tasks||[]).find(x=>x.id===id);if(!t)return;
  let opts=(projects||[]).map(p=>'<option value="'+p.id+'"'+(p.id===t.project_id?' selected':'')+'>'+esc(p.name)+'</option>').join('');
  document.getElementById('modal').innerHTML='<h2>Edit Task</h2><div class="form-row"><label>Title</label><input id="f-title" value="'+esc(t.title)+'"></div><div class="form-row"><label>Description</label><textarea id="f-desc">'+esc(t.description||'')+'</textarea></div><div class="form-row"><label>Status</label><select id="f-status"><option value="todo"'+(t.status==='todo'?' selected':'')+'>To Do</option><option value="in_progress"'+(t.status==='in_progress'?' selected':'')+'>In Progress</option><option value="done"'+(t.status==='done'?' selected':'')+'>Done</option></select></div><div class="form-row"><label>Priority</label><select id="f-pri"><option value="urgent"'+(t.priority==='urgent'?' selected':'')+'>Urgent</option><option value="high"'+(t.priority==='high'?' selected':'')+'>High</option><option value="medium"'+(t.priority==='medium'?' selected':'')+'>Medium</option><option value="low"'+(t.priority==='low'?' selected':'')+'>Low</option></select></div><div class="form-row"><label>Assignee</label><input id="f-assign" value="'+esc(t.assignee||'')+'"></div><div class="form-row"><label>Due date</label><input id="f-due" type="date" value="'+esc(t.due_date||'')+'"></div><div class="actions"><button class="btn" onclick="delTask(\''+id+'\')" style="color:var(--red)">Delete</button><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="updateTask(\''+id+'\')">Save</button></div>';
  document.getElementById('modalBg').classList.add('open');
}

async function updateTask(id){
  await fetch(API+'/tasks/'+id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({title:document.getElementById('f-title').value,description:document.getElementById('f-desc').value,status:document.getElementById('f-status').value,priority:document.getElementById('f-pri').value,assignee:document.getElementById('f-assign').value,due_date:document.getElementById('f-due').value,tags:[]})});
  closeModal();load();
}
async function delTask(id){if(confirm('Delete?')){await fetch(API+'/tasks/'+id,{method:'DELETE'});closeModal();load();}}

function openProjectForm(){
  document.getElementById('modal').innerHTML='<h2>New Project</h2><div class="form-row"><label>Name</label><input id="f-name"></div><div class="form-row"><label>Color</label><input id="f-color" type="color" value="#c45d2c"></div><div class="actions"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-primary" onclick="submitProject()">Create</button></div>';
  document.getElementById('modalBg').classList.add('open');
}
async function submitProject(){await fetch(API+'/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('f-name').value,color:document.getElementById('f-color').value})});closeModal();load();}

function closeModal(){document.getElementById('modalBg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
