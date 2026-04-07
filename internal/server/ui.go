package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Roundup</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;gap:1rem;flex-wrap:wrap}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.hdr h1 span{color:var(--rust)}
.hdr-r{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}
.stats-line{font-size:.6rem;color:var(--cm)}
.stats-line .num{color:var(--cream);font-weight:700}
.stats-line .overdue{color:var(--red)}
.proj-bar{display:flex;gap:.3rem;padding:.5rem 1.5rem;border-bottom:1px solid var(--bg3);flex-wrap:wrap;align-items:center}
.proj-btn{font-size:.6rem;padding:.25rem .55rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);cursor:pointer;font-family:var(--mono);display:inline-flex;align-items:center;gap:.3rem}
.proj-btn:hover{border-color:var(--leather)}
.proj-btn.active{border-color:var(--rust);color:var(--rust)}
.proj-btn .dot{width:8px;height:8px;border-radius:50%;display:inline-block}
.proj-edit{font-size:.55rem;color:var(--cm);cursor:pointer;margin-left:.2rem}
.proj-edit:hover{color:var(--cream)}
.main{padding:1rem 1.5rem 1.5rem;max-width:1200px;margin:0 auto}
.board{display:grid;grid-template-columns:repeat(3,1fr);gap:.8rem;min-height:60vh}
.col{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem;min-height:200px}
.col-hdr{font-size:.6rem;text-transform:uppercase;letter-spacing:1.5px;margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.col-todo .col-hdr{color:var(--cd)}
.col-prog .col-hdr{color:var(--orange)}
.col-done .col-hdr{color:var(--green)}
.col-count{font-size:.55rem;background:var(--bg3);padding:.1rem .35rem;color:var(--cm)}
.task{background:var(--bg);border:1px solid var(--bg3);padding:.7rem .8rem;margin-bottom:.5rem;transition:border-color .15s;font-size:.72rem}
.task:hover{border-color:var(--leather)}
.task.urgent{border-left:3px solid var(--red)}
.task.high{border-left:3px solid var(--orange)}
.task.overdue{background:#2a1818}
.task-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem;margin-bottom:.3rem}
.task-title{color:var(--cream);font-weight:500;flex:1;line-height:1.4}
.task-desc{font-size:.65rem;color:var(--cm);margin-top:.2rem;font-style:italic;line-height:1.4}
.task-meta{font-size:.55rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap;align-items:center;margin-top:.4rem}
.task-meta .due{color:var(--cd)}
.task-meta .due.overdue{color:var(--red)}
.task-tag{font-size:.5rem;background:var(--bg3);color:var(--cd);padding:.05rem .3rem}
.task-proj{font-size:.5rem;color:var(--cm);display:inline-flex;align-items:center;gap:.2rem}
.task-proj .dot{width:6px;height:6px;border-radius:50%;display:inline-block}
.task-extra{font-size:.55rem;color:var(--cd);margin-top:.4rem;padding-top:.35rem;border-top:1px dashed var(--bg3)}
.task-extra-row{display:flex;gap:.4rem;margin-bottom:.1rem}
.task-extra-label{color:var(--cm);text-transform:uppercase;letter-spacing:.5px;min-width:80px}
.task-extra-val{color:var(--cream)}
.pri{font-size:.48rem;padding:.1rem .3rem;text-transform:uppercase;letter-spacing:1px;font-weight:700}
.pri-urgent{background:#c9444433;color:var(--red);border:1px solid #c9444466}
.pri-high{background:#d4843a22;color:var(--orange);border:1px solid #d4843a44}
.pri-medium{background:var(--bg3);color:var(--cd);border:1px solid var(--bg3)}
.pri-low{background:var(--bg3);color:var(--cm);border:1px solid var(--bg3)}
.task-actions{display:flex;gap:.3rem;margin-top:.5rem;flex-wrap:wrap}
.btn{font-size:.58rem;padding:.2rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);font-family:var(--mono);transition:all .15s}
.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-p:hover{opacity:.85;color:#fff}
.btn-sm{font-size:.52rem;padding:.15rem .35rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}
.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:480px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}
.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.fr input[type=color]{height:2rem;padding:.2rem}
.fr-section{margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--bg3)}
.fr-section-label{font-size:.55rem;color:var(--rust);text-transform:uppercase;letter-spacing:1px;margin-bottom:.5rem}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.acts .btn-del{margin-right:auto;color:var(--red);border-color:#3a1a1a}
.acts .btn-del:hover{border-color:var(--red);color:var(--red)}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-size:.7rem}
@media(max-width:780px){.board{grid-template-columns:1fr}}
</style>
</head>
<body>

<div class="hdr">
<h1 id="dash-title"><span>&#9670;</span> ROUNDUP</h1>
<div class="hdr-r">
<span class="stats-line" id="stats"></span>
<button class="btn btn-p" onclick="openTaskForm()">+ Task</button>
<button class="btn" onclick="openProjForm()">+ Project</button>
</div>
</div>

<div class="proj-bar" id="projBar"></div>

<div class="main">
<div class="board" id="board"></div>
</div>

<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()">
<div class="modal" id="mdl"></div>
</div>

<script>
var A='/api';
var RESOURCE_TASKS='tasks';
var RESOURCE_PROJECTS='projects';

// Field defs for tasks. Custom fields injected from /api/config get
// isCustom=true and persist to the extras table.
var fields=[
{name:'title',label:'Title',type:'text',required:true,placeholder:'What needs doing?'},
{name:'description',label:'Description',type:'textarea',placeholder:'Details (optional)'},
{name:'project_id',label:'Project',type:'project_select'},
{name:'priority',label:'Priority',type:'select',options:['low','medium','high','urgent']},
{name:'status',label:'Status',type:'select',options:['todo','in_progress','done']},
{name:'assignee',label:'Assignee',type:'text',placeholder:'Who owns this'},
{name:'due_date',label:'Due Date',type:'date'},
{name:'tags',label:'Tags',type:'tags',placeholder:'comma separated'}
];

var projects=[],tasks=[],taskExtras={},curProj='',editTaskId=null,editProjId=null;

// ─── Helpers ──────────────────────────────────────────────────────

function fmtDate(s){
if(!s)return'';
try{
var d=new Date(s);
if(isNaN(d.getTime()))return s;
return d.toLocaleDateString('en-US',{month:'short',day:'numeric'});
}catch(e){return s}
}

function isOverdue(dueDate,status){
if(!dueDate||status==='done')return false;
var today=new Date();
today.setHours(0,0,0,0);
var d=new Date(dueDate);
if(isNaN(d.getTime()))return false;
return d<today;
}

function projectByID(id){
for(var i=0;i<projects.length;i++)if(projects[i].id===id)return projects[i];
return null;
}

function fieldByName(n){
for(var i=0;i<fields.length;i++)if(fields[i].name===n)return fields[i];
return null;
}

// ─── Loading and rendering ────────────────────────────────────────

async function load(){
try{
var resps=await Promise.all([
fetch(A+'/projects').then(function(r){return r.json()}),
fetch(A+'/stats').then(function(r){return r.json()})
]);
projects=resps[0].projects||[];
renderStats(resps[1]);
}catch(e){
console.error('load failed',e);
projects=[];
}
renderProjBar();
loadTasks();
}

async function loadTasks(){
try{
var url=A+'/tasks';
if(curProj)url+='?project='+encodeURIComponent(curProj);
var r=await fetch(url).then(function(r){return r.json()});
tasks=r.tasks||[];
try{
var ex=await fetch(A+'/extras/'+RESOURCE_TASKS).then(function(r){return r.json()});
taskExtras=ex||{};
tasks.forEach(function(t){
var e=taskExtras[t.id];
if(!e)return;
Object.keys(e).forEach(function(k){if(t[k]===undefined)t[k]=e[k]});
});
}catch(e){taskExtras={}}
}catch(e){
console.error('load tasks failed',e);
tasks=[];
}
renderBoard();
}

function renderStats(st){
if(!st)st={};
var todo=st.todo||0;
var prog=st.in_progress||0;
var done=st.done||0;
var overdue=st.overdue||0;
var html='<span class="num">'+todo+'</span> todo · <span class="num">'+prog+'</span> in progress · <span class="num">'+done+'</span> done';
if(overdue>0)html+=' · <span class="overdue">'+overdue+' overdue</span>';
document.getElementById('stats').innerHTML=html;
}

function renderProjBar(){
var h='<button class="proj-btn'+(curProj===''?' active':'')+'" onclick="filterProj(\'\')">All</button>';
projects.forEach(function(p){
var active=curProj===p.id;
h+='<button class="proj-btn'+(active?' active':'')+'" onclick="filterProj(\''+p.id+'\')">';
h+='<span class="dot" style="background:'+esc(p.color||'#c45d2c')+'"></span>';
h+=esc(p.name)+' <span style="color:var(--cm)">'+p.task_count+'</span>';
if(active)h+='<span class="proj-edit" onclick="event.stopPropagation();openProjEdit(\''+p.id+'\')" title="Edit project">✎</span>';
h+='</button>';
});
document.getElementById('projBar').innerHTML=h;
}

function filterProj(id){
curProj=id;
renderProjBar();
loadTasks();
}

function renderBoard(){
var todo=tasks.filter(function(t){return t.status==='todo'});
var prog=tasks.filter(function(t){return t.status==='in_progress'});
var done=tasks.filter(function(t){return t.status==='done'});
document.getElementById('board').innerHTML=
colHTML('col-todo','Todo',todo)+
colHTML('col-prog','In Progress',prog)+
colHTML('col-done','Done',done);
}

function colHTML(cls,label,list){
var h='<div class="col '+cls+'"><div class="col-hdr"><span>'+label+'</span><span class="col-count">'+list.length+'</span></div>';
if(list.length===0){
h+='<div class="empty">No tasks</div>';
}else{
list.forEach(function(t){h+=taskCard(t)});
}
h+='</div>';
return h;
}

function taskCard(t){
var overdue=isOverdue(t.due_date,t.status);
var cls='task '+(t.priority||'medium');
if(overdue)cls+=' overdue';

var h='<div class="'+cls+'"><div class="task-top">';
h+='<div class="task-title">'+esc(t.title)+'</div>';
h+='<span class="pri pri-'+esc(t.priority||'medium')+'">'+esc(t.priority||'medium')+'</span>';
h+='</div>';

if(t.description)h+='<div class="task-desc">'+esc(t.description)+'</div>';

h+='<div class="task-meta">';
if(t.project_id){
var proj=projectByID(t.project_id);
if(proj){
h+='<span class="task-proj"><span class="dot" style="background:'+esc(proj.color||'#c45d2c')+'"></span>'+esc(proj.name)+'</span>';
}
}
if(t.assignee)h+='<span>@'+esc(t.assignee)+'</span>';
if(t.due_date){
var dueCls='due'+(overdue?' overdue':'');
h+='<span class="'+dueCls+'">'+esc(fmtDate(t.due_date))+'</span>';
}
if(t.tags&&t.tags.length){
t.tags.forEach(function(tag){h+='<span class="task-tag">#'+esc(tag)+'</span>'});
}
h+='</div>';

// Custom field display
var customRows='';
fields.forEach(function(f){
if(!f.isCustom)return;
var v=t[f.name];
if(v===undefined||v===null||v==='')return;
customRows+='<div class="task-extra-row">';
customRows+='<span class="task-extra-label">'+esc(f.label)+'</span>';
customRows+='<span class="task-extra-val">'+esc(String(v))+'</span>';
customRows+='</div>';
});
if(customRows)h+='<div class="task-extra">'+customRows+'</div>';

// Action buttons
var nextStatus=t.status==='todo'?'in_progress':t.status==='in_progress'?'done':null;
h+='<div class="task-actions">';
if(nextStatus){
var label=nextStatus==='in_progress'?'Start':'Complete';
h+='<button class="btn btn-sm" onclick="moveTask(\''+t.id+'\',\''+nextStatus+'\')">'+label+'</button>';
}
if(t.status==='done'){
h+='<button class="btn btn-sm" onclick="moveTask(\''+t.id+'\',\'todo\')">Reopen</button>';
}
h+='<button class="btn btn-sm" onclick="openTaskEdit(\''+t.id+'\')">Edit</button>';
h+='<button class="btn btn-sm" onclick="delTask(\''+t.id+'\')" style="color:var(--red)">&#10005;</button>';
h+='</div>';

h+='</div>';
return h;
}

async function moveTask(id,status){
try{
await fetch(A+'/tasks/'+id+'/status',{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:status})});
}catch(e){alert('Failed to move task');return}
load();
}

async function delTask(id){
if(!confirm('Delete this task?'))return;
await fetch(A+'/tasks/'+id,{method:'DELETE'});
load();
}

// ─── Task form ────────────────────────────────────────────────────

function fieldHTML(f,value){
var v=value;
if(v===undefined||v===null)v='';
var req=f.required?' *':'';
var ph='';
if(f.placeholder)ph=' placeholder="'+esc(f.placeholder)+'"';
else if(f.name==='title'&&window._placeholderName)ph=' placeholder="'+esc(window._placeholderName)+'"';

var h='<div class="fr"><label>'+esc(f.label)+req+'</label>';

if(f.type==='select'){
h+='<select id="f-'+f.name+'">';
if(!f.required)h+='<option value="">Select...</option>';
(f.options||[]).forEach(function(o){
var sel=(String(v)===String(o))?' selected':'';
var disp=(typeof o==='string')?(o.charAt(0).toUpperCase()+o.slice(1).replace(/_/g,' ')):String(o);
h+='<option value="'+esc(String(o))+'"'+sel+'>'+esc(disp)+'</option>';
});
h+='</select>';
}else if(f.type==='project_select'){
h+='<select id="f-'+f.name+'">';
h+='<option value="">No Project</option>';
projects.forEach(function(p){
var sel=(v===p.id)?' selected':'';
h+='<option value="'+esc(p.id)+'"'+sel+'>'+esc(p.name)+'</option>';
});
h+='</select>';
}else if(f.type==='textarea'){
h+='<textarea id="f-'+f.name+'" rows="2"'+ph+'>'+esc(String(v))+'</textarea>';
}else if(f.type==='tags'){
// Tags come in as array — convert to comma string for the input
var tagStr=Array.isArray(v)?v.join(', '):String(v||'');
h+='<input type="text" id="f-'+f.name+'" value="'+esc(tagStr)+'"'+ph+'>';
}else if(f.type==='checkbox'){
h+='<input type="checkbox" id="f-'+f.name+'"'+(v?' checked':'')+' style="width:auto">';
}else if(f.type==='number'||f.type==='integer'){
h+='<input type="number" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}else{
var inputType=f.type||'text';
h+='<input type="'+esc(inputType)+'" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}

h+='</div>';
return h;
}

function taskFormHTML(task){
var t=task||{};
var isEdit=!!task;
var h='<h2>'+(isEdit?'EDIT TASK':'NEW TASK')+'</h2>';

// Title on its own row
h+=fieldHTML(fieldByName('title'),t.title);
// Description
h+=fieldHTML(fieldByName('description'),t.description);
// Project + priority
h+='<div class="row2">'+fieldHTML(fieldByName('project_id'),t.project_id)+fieldHTML(fieldByName('priority'),t.priority)+'</div>';
// Status + assignee
h+='<div class="row2">'+fieldHTML(fieldByName('status'),t.status||'todo')+fieldHTML(fieldByName('assignee'),t.assignee)+'</div>';
// Due date + tags
h+='<div class="row2">'+fieldHTML(fieldByName('due_date'),t.due_date)+fieldHTML(fieldByName('tags'),t.tags)+'</div>';

// Custom fields from personalization
var customFields=fields.filter(function(f){return f.isCustom});
if(customFields.length){
var sectionLabel=window._customSectionLabel||'Additional Details';
h+='<div class="fr-section"><div class="fr-section-label">'+esc(sectionLabel)+'</div>';
customFields.forEach(function(f){h+=fieldHTML(f,t[f.name])});
h+='</div>';
}

h+='<div class="acts">';
if(isEdit){
h+='<button class="btn btn-del" onclick="delTaskFromForm()">Delete</button>';
}
h+='<button class="btn" onclick="closeModal()">Cancel</button>';
h+='<button class="btn btn-p" onclick="submitTask()">'+(isEdit?'Save':'Create')+'</button>';
h+='</div>';
return h;
}

function openTaskForm(){
editTaskId=null;
editProjId=null;
document.getElementById('mdl').innerHTML=taskFormHTML();
document.getElementById('mbg').classList.add('open');
var titleEl=document.getElementById('f-title');
if(titleEl)titleEl.focus();
}

function openTaskEdit(id){
var t=null;
for(var i=0;i<tasks.length;i++)if(tasks[i].id===id){t=tasks[i];break}
if(!t)return;
editTaskId=id;
editProjId=null;
document.getElementById('mdl').innerHTML=taskFormHTML(t);
document.getElementById('mbg').classList.add('open');
}

async function submitTask(){
var titleEl=document.getElementById('f-title');
if(!titleEl||!titleEl.value.trim()){alert('Title is required');return}

var body={};
var extras={};
fields.forEach(function(f){
var el=document.getElementById('f-'+f.name);
if(!el)return;
var val;
if(f.type==='checkbox')val=el.checked;
else if(f.type==='number'||f.type==='integer')val=parseFloat(el.value)||0;
else if(f.type==='tags'){
val=el.value.split(',').map(function(s){return s.trim()}).filter(function(s){return s.length});
}else val=el.value.trim();
if(f.isCustom)extras[f.name]=val;
else body[f.name]=val;
});

var savedId=editTaskId;
try{
if(editTaskId){
var r1=await fetch(A+'/tasks/'+editTaskId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r1.ok){var e1=await r1.json().catch(function(){return{}});alert(e1.error||'Save failed');return}
}else{
var r2=await fetch(A+'/tasks',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r2.ok){var e2=await r2.json().catch(function(){return{}});alert(e2.error||'Save failed');return}
var created=await r2.json();
savedId=created.id;
}
if(savedId&&Object.keys(extras).length){
await fetch(A+'/extras/'+RESOURCE_TASKS+'/'+savedId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(extras)}).catch(function(){});
}
}catch(e){
alert('Network error: '+e.message);
return;
}

closeModal();
load();
}

async function delTaskFromForm(){
if(!editTaskId)return;
if(!confirm('Delete this task?'))return;
await fetch(A+'/tasks/'+editTaskId,{method:'DELETE'});
closeModal();
load();
}

// ─── Project form ─────────────────────────────────────────────────

function projFormHTML(proj){
var p=proj||{name:'',color:'#c45d2c'};
var isEdit=!!proj;
var h='<h2>'+(isEdit?'EDIT PROJECT':'NEW PROJECT')+'</h2>';
h+='<div class="fr"><label>Name *</label><input type="text" id="f-pn" value="'+esc(p.name)+'" placeholder="Project name"></div>';
h+='<div class="fr"><label>Color</label><input type="color" id="f-pc" value="'+esc(p.color||'#c45d2c')+'"></div>';
h+='<div class="acts">';
if(isEdit){
h+='<button class="btn btn-del" onclick="delProjFromForm()">Delete (and all tasks)</button>';
}
h+='<button class="btn" onclick="closeModal()">Cancel</button>';
h+='<button class="btn btn-p" onclick="submitProj()">'+(isEdit?'Save':'Create')+'</button>';
h+='</div>';
return h;
}

function openProjForm(){
editProjId=null;
editTaskId=null;
document.getElementById('mdl').innerHTML=projFormHTML();
document.getElementById('mbg').classList.add('open');
var nameEl=document.getElementById('f-pn');
if(nameEl)nameEl.focus();
}

function openProjEdit(id){
var p=projectByID(id);
if(!p)return;
editProjId=id;
editTaskId=null;
document.getElementById('mdl').innerHTML=projFormHTML(p);
document.getElementById('mbg').classList.add('open');
}

async function submitProj(){
var nameEl=document.getElementById('f-pn');
if(!nameEl||!nameEl.value.trim()){alert('Project name is required');return}
var body={name:nameEl.value.trim(),color:document.getElementById('f-pc').value};
try{
if(editProjId){
var r1=await fetch(A+'/projects/'+editProjId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r1.ok){alert('Save failed');return}
}else{
var r2=await fetch(A+'/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r2.ok){alert('Save failed');return}
}
}catch(e){alert('Network error: '+e.message);return}
closeModal();
load();
}

async function delProjFromForm(){
if(!editProjId)return;
if(!confirm('Delete this project AND all its tasks? This cannot be undone.'))return;
await fetch(A+'/projects/'+editProjId,{method:'DELETE'});
if(curProj===editProjId)curProj='';
closeModal();
load();
}

function closeModal(){
document.getElementById('mbg').classList.remove('open');
editTaskId=null;
editProjId=null;
}

function esc(s){
if(s===undefined||s===null)return'';
var d=document.createElement('div');
d.textContent=String(s);
return d.innerHTML;
}

document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});

// ─── Personalization ──────────────────────────────────────────────

(function loadPersonalization(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||typeof cfg!=='object')return;

if(cfg.dashboard_title){
var h1=document.getElementById('dash-title');
if(h1)h1.innerHTML='<span>&#9670;</span> '+esc(cfg.dashboard_title);
document.title=cfg.dashboard_title;
}

if(cfg.empty_state_message)window._emptyMsg=cfg.empty_state_message;
if(cfg.placeholder_name)window._placeholderName=cfg.placeholder_name;
if(cfg.primary_label)window._customSectionLabel=cfg.primary_label+' Details';

if(Array.isArray(cfg.custom_fields)){
cfg.custom_fields.forEach(function(cf){
if(!cf||!cf.name||!cf.label)return;
if(fieldByName(cf.name))return;
fields.push({
name:cf.name,
label:cf.label,
type:cf.type||'text',
options:cf.options||[],
isCustom:true
});
});
}
}).catch(function(){
}).finally(function(){
load();
});
})();
</script>
</body>
</html>`
