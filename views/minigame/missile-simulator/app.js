'use strict';
const {Simulation,CONFIG,TYPES}=NavalSim,sim=new Simulation();
const canvas=document.getElementById('scene'),ctx=canvas.getContext('2d'),$=id=>document.getElementById(id);
let width=1000,height=480,paused=false,last=0,accumulator=0;
const world={w:6800,h:1450},camera=new NavalCamera();
function cameraUI(){ $('pan').max=camera.maxX;$('pan').value=camera.x;$('zoomLabel').textContent=Math.round(camera.scale/.22*100)+'%';}
$('pan').oninput=e=>{camera.pan(+e.target.value);cameraUI();};
$('zoomIn').onclick=()=>{camera.zoom(camera.scale*1.25);cameraUI();};
$('zoomOut').onclick=()=>{camera.zoom(camera.scale/1.25);cameraUI();};
$('fit').onclick=()=>{camera.fit();cameraUI();};
$('homeView').onclick=()=>{camera.scale=.22;camera.pan(0);cameraUI();};
function fullscreenLabel(){ $('fullscreen').textContent=document.fullscreenElement||$('stage').classList.contains('expanded')?'전체화면 종료':'전체화면';}
$('fullscreen').onclick=async()=>{
 const stage=$('stage');
 if(document.fullscreenElement){await document.exitFullscreen();}
 else if(stage.classList.contains('expanded')){stage.classList.remove('expanded');}
 else {try{await stage.requestFullscreen();}catch{stage.classList.add('expanded');}}
 fullscreenLabel();
};
document.addEventListener('fullscreenchange',fullscreenLabel);
document.addEventListener('keydown',e=>{if(e.key==='Escape'){$('stage').classList.remove('expanded');fullscreenLabel();}});
let lastSlotPress={slot:-1,time:-Infinity};
function pressSlot(slot){
 const now=performance.now();
 const twice=lastSlotPress.slot===slot&&now-lastSlotPress.time<=350;
 sim.select(slot);
 lastSlotPress=twice?{slot:-1,time:-Infinity}:{slot,time:now};
 if(twice)fire();
}
function resize(){const r=canvas.getBoundingClientRect();width=r.width;height=r.height;camera.resize(width,height);cameraUI();const dpr=window.devicePixelRatio||1;canvas.width=Math.round(width*dpr);canvas.height=Math.round(height*dpr);ctx.setTransform(dpr,0,0,dpr,0,0);}
new ResizeObserver(resize).observe(canvas);
TYPES.forEach((t,i)=>{const b=document.createElement('button');b.className='slot';b.style.setProperty('--accent',t.color);b.innerHTML=`<b><kbd>${i+1}</kbd> ${t.name}</b><span></span><div class="bar"></div>`;b.addEventListener('click',()=>pressSlot(i));$('slots').append(b);});
function fire(){if(!paused)sim.fire();}
function pointer(e){const r=canvas.getBoundingClientRect();sim.aim=camera.world(e.clientX-r.left,e.clientY-r.top);}
canvas.addEventListener('pointermove',pointer);canvas.addEventListener('pointerdown',e=>{pointer(e);fire();});$('fire').onclick=fire;
function pause(){paused=!paused;$('pause').textContent=paused?'계속하기':'일시정지';}
$('pause').onclick=pause;$('reset').onclick=()=>{lastSlotPress={slot:-1,time:-Infinity};sim.reset();sim.boost=+$('boost').value;sim.reacquire=+$('turn').value;sim.lateralAcceleration=+$('lateral').value*18;sim.obstaclesEnabled=$('obstacles').checked;sim.aimAssist=$('aimAssist').checked;paused=false;$('pause').textContent='일시정지';};
document.addEventListener('keydown',e=>{if(/INPUT|TEXTAREA|SELECT/.test(e.target.tagName))return;if(e.repeat)return;if(/^[1-4]$/.test(e.key)){e.preventDefault();pressSlot(+e.key-1);}if(e.code==='Space'){e.preventDefault();fire();}if(e.code==='KeyP')pause();});
$('boost').oninput=e=>{sim.boost=+e.target.value;$('boostValue').textContent=sim.boost.toFixed(1)+' s';};$('turn').oninput=e=>{sim.reacquire=+e.target.value;$('turnValue').textContent=sim.reacquire+' °/s';};
$('lateral').oninput=e=>{sim.lateralAcceleration=+e.target.value*18;$('lateralValue').textContent=e.target.value+' unit/s²';};
$('aimAssist').onchange=e=>sim.aimAssist=e.target.checked;
$('obstacles').onchange=e=>sim.obstaclesEnabled=e.target.checked;
document.addEventListener('visibilitychange',()=>{last=0;accumulator=0;});
function line(x1,y1,x2,y2,color,w=1){ctx.beginPath();ctx.moveTo(x1,y1);ctx.lineTo(x2,y2);ctx.strokeStyle=color;ctx.lineWidth=w;ctx.stroke();}
function draw(){
 const aim=sim.correctedAim();
 const sx=camera.scale,sy=camera.scale,X=x=>(x-camera.x)*sx,Y=y=>(y-camera.y)*sy,S=n=>n*camera.scale;
 const bg=ctx.createLinearGradient(0,0,0,height);bg.addColorStop(0,'#0e1929');bg.addColorStop(1,'#142d40');ctx.fillStyle=bg;ctx.fillRect(0,0,width,height);
 for(let x=0;x<world.w;x+=200)line(X(x),0,X(x),height,'#91b4d009');for(let y=0;y<world.h;y+=200)line(0,Y(y),width,Y(y),'#91b4d009');
 ctx.fillStyle='#0b263b';ctx.fillRect(0,Y(1180),width,height-Y(1180));line(0,Y(1180),width,Y(1180),'#3c829a',1);
 for(let i=0;i<28;i++){let x=(i*143+sim.time*14)%world.w;line(X(x),Y(1210+(i%4)*38),X(x+50),Y(1210+(i%4)*38),'#4d8fad25');}
 ctx.font='10px sans-serif';ctx.fillStyle='#567e98';ctx.fillText('SEA LEVEL',12,Y(1180)-8);
 function ship(x,y,w,h,own){ctx.fillStyle=own?'#36576b':'#374857';ctx.beginPath();ctx.moveTo(X(x),Y(y+h*.55));ctx.lineTo(X(x+w),Y(y+h*.55));ctx.lineTo(X(x+w*.86),Y(y+h));ctx.lineTo(X(x+w*.13),Y(y+h));ctx.closePath();ctx.fill();ctx.fillStyle=own?'#658797':'#687783';ctx.fillRect(X(x+w*.35),Y(y),S(w*.3),S(h*.6));ctx.fillRect(X(x+w*.48),Y(y-35),Math.max(2,S(8)),S(40));}
 ship(650,1095,400,85,true);ctx.fillStyle='#8cabbf';ctx.fillText('YOUR SHIP',X(710),Y(1300));
 sim.targets.forEach((t,i)=>{ship(t.x,t.y,t.w,t.h,false);ctx.fillStyle='#bc9677';ctx.fillText('TARGET 0'+(i+1),X(t.x),Y(t.y-65));});
 if(sim.obstaclesEnabled)for(const rock of sim.obstacles){ctx.fillStyle='#344752';ctx.fillRect(X(rock.x),Y(rock.y),S(rock.w),S(rock.h));ctx.strokeStyle='#657f88';ctx.strokeRect(X(rock.x),Y(rock.y),S(rock.w),S(rock.h));ctx.fillStyle='#9bacb7';ctx.fillText('OBSTACLE',X(rock.x)-5,Y(rock.y)-14);}
 TYPES.forEach((t,i)=>{ctx.fillStyle=t.color;for(let n=0;n<t.count;n++)ctx.fillRect(X(755+i*55+n*10),Y(1090),Math.max(2,S(7)),S(15));});
 for(const m of sim.missiles){if(sim.time<m.launchAt)continue;const color=TYPES[m.slot].color;ctx.beginPath();m.trail.forEach((p,i)=>i?ctx.lineTo(X(p.x),Y(p.y)):ctx.moveTo(X(p.x),Y(p.y)));ctx.strokeStyle=color+'55';ctx.lineWidth=1.5;ctx.stroke();if(m.slot===sim.selected&&m.age>=m.boost){ctx.setLineDash([3,5]);line(X(m.x),Y(m.y),X(aim.x),Y(aim.y),color+'60');ctx.setLineDash([]);}ctx.save();ctx.translate(X(m.x),Y(m.y));ctx.rotate(Math.atan2(Math.sin(m.angle)*sy,Math.cos(m.angle)*sx));ctx.fillStyle=color;ctx.shadowColor=color;ctx.shadowBlur=10;ctx.beginPath();ctx.moveTo(7,0);ctx.lineTo(-4,-2.5);ctx.lineTo(-2,0);ctx.lineTo(-4,2.5);ctx.closePath();ctx.fill();ctx.shadowBlur=0;ctx.fillStyle='#ffb76c';ctx.beginPath();ctx.moveTo(-5,-1.5);ctx.lineTo(-11-Math.random()*5,0);ctx.lineTo(-5,1.5);ctx.fill();ctx.restore();}
 for(const e of sim.events){const p=(sim.time-e.time)/.9;ctx.beginPath();ctx.arc(X(e.x),Y(e.y),5+p*27,0,Math.PI*2);ctx.strokeStyle=`rgba(255,${e.hit?175:220},130,${1-p})`;ctx.lineWidth=2;ctx.stroke();}
 const ax=X(aim.x),ay=Y(aim.y),color=TYPES[sim.selected].color;ctx.strokeStyle=color;ctx.lineWidth=1;ctx.beginPath();ctx.arc(ax,ay,9,0,Math.PI*2);ctx.stroke();line(ax-15,ay,ax-5,ay,color);line(ax+5,ay,ax+15,ay,color);line(ax,ay-15,ax,ay-5,color);line(ax,ay+5,ax,ay+15,color);
 if(aim.assisted){line(X(sim.aim.x)-4,Y(sim.aim.y),X(sim.aim.x)+4,Y(sim.aim.y),'#ffffff88');line(X(sim.aim.x),Y(sim.aim.y)-4,X(sim.aim.x),Y(sim.aim.y)+4,'#ffffff88');line(X(sim.aim.x),Y(sim.aim.y),ax,ay,'#ffffff44');}
 [...$('slots').children].forEach((b,i)=>{const remaining=Math.max(0,sim.ready[i]-sim.time),active=sim.missiles.filter(m=>m.slot===i).length;b.classList.toggle('active',i===sim.selected);b.setAttribute('aria-pressed',String(i===sim.selected));b.querySelector('span').textContent=(remaining?remaining.toFixed(1)+'s 재장전':'발사 준비')+(active?' · '+active+'기 비행':'');b.querySelector('.bar').style.width=(1-remaining/CONFIG.reload)*100+'%';});
 const current=sim.missiles.filter(m=>m.slot===sim.selected&&sim.time>=m.launchAt);$('mode').textContent=`0${sim.selected+1} / ${TYPES[sim.selected].name}`;$('telemetry').textContent=paused?'일시정지':current.length?`${current.length}기 ${current[0].age<current[0].boost?'초기 가속':'유도 중'} · ${Math.round(current[0].speed/6)} unit/s ? R ${Math.abs(current[0].omega||0)>.02?Math.round(current[0].speed/Math.abs(current[0].omega)):'?'}`:'발사 대기';$('stats').textContent=`명중 ${sim.hits} · 소멸 ${sim.lost}`;$('fire').disabled=paused||sim.ready[sim.selected]>sim.time;
}
function frame(now){if(last&&!paused&&!document.hidden){accumulator+=Math.min((now-last)/1000,.1);while(accumulator>=1/120){sim.step(1/120);accumulator-=1/120;}}last=now;draw();requestAnimationFrame(frame);}resize();requestAnimationFrame(frame);
