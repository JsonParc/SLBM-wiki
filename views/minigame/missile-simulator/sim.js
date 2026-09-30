(function(root){
  'use strict';
  const CONFIG={startSpeed:2097,maxSpeed:2910,acceleration:(2910-2097)/.93,maxSpeedTime:.93,farDistance:900,reload:20,range:12000,interval:.18};
  const TYPES=[{name:'VLS · 4연장',count:4,type:'vls',color:'#72e7ce'},{name:'VLS · 2연장',count:2,type:'vls',color:'#8ab8ff'},{name:'전방 · 2연장',count:2,type:'front',color:'#ffca7e'},{name:'양옆 · 2연장',count:2,type:'side',color:'#de9fff'}];
  const wrap=a=>Math.atan2(Math.sin(a),Math.cos(a));
  function reacquisitionGeometry(origin,direction,aim){
    const ux=Math.cos(direction),uy=Math.sin(direction);
    const along=(aim.x-origin.x)*ux+(aim.y-origin.y)*uy;
    const foot={x:origin.x+along*ux,y:origin.y+along*uy};
    const distance=Math.hypot(aim.x-foot.x,aim.y-foot.y);
    const seconds=(distance*.5)/CONFIG.maxSpeed;
    return {origin,foot,distance,seconds,multiplier:seconds>0&&seconds<1?1/seconds:1};
  }
  function intersects(ax,ay,bx,by,r){let lo=0,hi=1;for(const [p,d,min,max] of [[ax,bx-ax,r.x,r.x+r.w],[ay,by-ay,r.y,r.y+r.h]]){if(Math.abs(d)<1e-9){if(p<min||p>max)return false;}else{let a=(min-p)/d,b=(max-p)/d;if(a>b)[a,b]=[b,a];lo=Math.max(lo,a);hi=Math.min(hi,b);if(lo>hi)return false;}}return true;}
  class Simulation{
    constructor(){this.reset();}
    correctedAim(){
      if(!this.aimAssist)return {...this.aim};
      let best=null,nearest=Infinity;
      for(const t of this.targets){
        const edgeDistance=Math.hypot(Math.max(t.x-this.aim.x,0,this.aim.x-t.x-t.w),Math.max(t.y-this.aim.y,0,this.aim.y-t.y-t.h));
        if(edgeDistance<180&&edgeDistance<nearest){nearest=edgeDistance;best=t;}
      }
      if(!best)return {...this.aim};
      const weight=(1-nearest/180)**2;
      return {x:this.aim.x+(best.x+best.w/2-this.aim.x)*weight,y:this.aim.y+(best.y+best.h/2-this.aim.y)*weight,assisted:true};
    }
    reset(){this.time=0;this.selected=0;this.missiles=[];this.events=[];this.ready=[0,0,0,0];this.aim={x:3200,y:400};this.boost=.7;this.reacquire=720;this.aimAssist=true;this.lateralAcceleration=18000;this.obstaclesEnabled=true;this.obstacles=[{x:4100,y:650,w:180,h:530}];this.hits=0;this.lost=0;this.targets=[{x:5100,y:1100,w:220,h:80},{x:6150,y:1060,w:260,h:120}];}
    select(slot){
      if(slot===this.selected)return;
      for(const m of this.missiles){
        if(m.slot===this.selected){
          if(this.time>=m.launchAt)m.releasedHeading=m.angle;
          m.reacquisition=null;
          m.omega=0;
        }
        if(m.slot===slot){
          m.fast=1.2;
          if(m.releasedHeading!==undefined)m.reacquisition={origin:{x:m.x,y:m.y},direction:m.releasedHeading};
        }
      }
      this.selected=slot;
    }
    fire(){const slot=this.selected;if(this.ready[slot]>this.time)return false;this.ready[slot]=this.time+CONFIG.reload;const def=TYPES[slot];for(let i=0;i<def.count;i++){const angle=def.type==='vls'?-Math.PI/2:def.type==='front'?-.22:i===0?-Math.PI+.3:-.3;this.missiles.push({slot,x:840+(i-(def.count-1)/2)*22,y:1092,angle,speed:CONFIG.startSpeed,distance:0,age:0,launchAt:this.time+i*CONFIG.interval,boost:def.type==='vls'?this.boost:.55,fast:0,trail:[]});}return true;}
    step(dt){
      this.time+=dt;
      this.events=this.events.filter(e=>this.time-e.time<.9);
      for(const m of this.missiles){
        if(m.dead||this.time<m.launchAt)continue;
        const ax=m.x,ay=m.y;
        m.age+=dt;
        const reacquiring=m.slot===this.selected&&!!m.reacquisition&&m.age>=m.boost;
        const aim=this.correctedAim();
        const geometry=reacquiring?reacquisitionGeometry(m.reacquisition.origin,m.reacquisition.direction,aim):null;
        const speedMultiplier=geometry?geometry.multiplier:1;
        m.omega=m.omega||0;
        if(m.age>=m.boost&&m.slot===this.selected){
          const delta=wrap(Math.atan2(aim.y-m.y,aim.x-m.x)-m.angle);
          // Lateral acceleration limits curvature: omega = a/v, radius = v²/a.
          // The frozen released heading defines the reference line, never a destination.
          const authority=reacquiring?3:m.fast>0?1+2*m.fast/1.2:1;
          const maxOmega=Math.min((reacquiring||m.fast>0?this.reacquire*2:480)*Math.PI/180,this.lateralAcceleration*authority/m.speed);
          const desired=Math.max(-maxOmega,Math.min(maxOmega,delta*10));
          const angularAcceleration=18*authority;
          m.omega+=Math.max(-angularAcceleration*dt,Math.min(angularAcceleration*dt,desired-m.omega));
          m.omega=Math.max(-maxOmega,Math.min(maxOmega,m.omega));
          m.angle+=m.omega*dt;
        }else m.omega=0;
        const guided=m.age>=m.boost&&m.slot===this.selected;
        const distance=Math.hypot(aim.x-m.x,aim.y-m.y);
        const cruise=reacquiring?CONFIG.maxSpeed*speedMultiplier:guided?CONFIG.startSpeed+(CONFIG.maxSpeed-CONFIG.startSpeed)*Math.min(1,distance/CONFIG.farDistance):CONFIG.maxSpeed;
        const turnFraction=Math.min(1,Math.abs(m.omega)/(Math.PI*1.5));
        const desiredSpeed=Math.max(CONFIG.startSpeed*.3,cruise*(1-.85*turnFraction));
        const speedRate=desiredSpeed<m.speed?CONFIG.maxSpeed*2:CONFIG.acceleration*speedMultiplier;
        m.speed+=Math.max(-speedRate*dt,Math.min(speedRate*dt,desiredSpeed-m.speed));
        m.fast=Math.max(0,m.fast-dt);
        m.x+=Math.cos(m.angle)*m.speed*dt;
        m.y+=Math.sin(m.angle)*m.speed*dt;
        m.distance+=m.speed*dt;
        // Arriving at the live cursor ends the boost; there is no return waypoint.
        if(reacquiring){
          const dx=m.x-ax,dy=m.y-ay;
          const t=Math.max(0,Math.min(1,((aim.x-ax)*dx+(aim.y-ay)*dy)/(dx*dx+dy*dy||1)));
          if(Math.hypot(ax+dx*t-aim.x,ay+dy*t-aim.y)<65){m.reacquisition=null;m.fast=1.2;}
        }
        m.trail.push({x:m.x,y:m.y});
        if(m.trail.length>720)m.trail.shift();
        const blocked=this.obstaclesEnabled&&this.obstacles.some(t=>intersects(ax,ay,m.x,m.y,t));
        const hit=!blocked&&this.targets.some(t=>intersects(ax,ay,m.x,m.y,t));
        if(hit||blocked||m.y>=1180||m.distance>=CONFIG.range){
          m.dead=true;
          if(hit)this.hits++;else this.lost++;
          this.events.push({x:m.x,y:Math.min(1180,m.y),time:this.time,hit});
        }
      }
      this.missiles=this.missiles.filter(m=>!m.dead);
    }
  }
  root.NavalSim={Simulation,CONFIG,TYPES,intersects,reacquisitionGeometry};
  if(typeof module!=='undefined')module.exports=root.NavalSim;
})(typeof window!=='undefined'?window:globalThis);
