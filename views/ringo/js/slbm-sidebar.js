"use strict";
(() => {
    const root = document.getElementById('slbm-sidebar-content');
    if (!root) return;
    let timer, last_html = '', pending = false;
    const reduced_motion = window.matchMedia('(prefers-reduced-motion: reduce)');
    const relative_time = new Intl.RelativeTimeFormat(document.documentElement.lang || 'ko', {numeric:'auto'});
    function update_times() {
        root.querySelectorAll('time').forEach(node => {
            const delta = Math.min(0, Math.round((new Date(node.dateTime) - Date.now()) / 1000));
            if (!Number.isFinite(delta)) return;
            const unit = Math.abs(delta) < 60 ? 'second' : Math.abs(delta) < 3600 ? 'minute' : Math.abs(delta) < 86400 ? 'hour' : 'day';
            const divisor = {second:1,minute:60,hour:3600,day:86400}[unit];
            node.textContent = relative_time.format(Math.trunc(delta/divisor),unit);
        });
    }
    function setup_hall() {
        clearInterval(timer);
        const hall = root.querySelector('.slbm-hall');
        if (!hall) return;
        const button = hall.querySelector('button');
        const rows = [...hall.querySelectorAll('li')];
        const stage = hall.querySelector('.slbm-hall-stage');
        let index = 0, pinned = false, hovering = false, focused = false;
        hall.classList.add('is-ready');
        if (!rows.length) hall.classList.add('is-empty');
        function render() {
            const row = document.createElement('div');
            if (rows.length) row.innerHTML = rows[index].innerHTML;
            stage.replaceChildren(row);
            return row;
        }
        function expand() {
            const open = pinned || hovering || focused;
            hall.classList.toggle('is-open',open);
            button.setAttribute('aria-expanded',String(open));
            if (open) stage.getAnimations({subtree:true}).forEach(animation => animation.cancel());
        }
        hall.addEventListener('pointerenter', event => { if (event.pointerType !== 'touch') { hovering=true;expand(); } });
        hall.addEventListener('pointerleave', () => { hovering=false;expand(); });
        hall.addEventListener('focusin', () => { focused=true;expand(); });
        hall.addEventListener('focusout', event => { if (!hall.contains(event.relatedTarget)) { focused=false;expand(); } });
        button.addEventListener('click', () => { pinned=!pinned; focused=false; hovering=false; expand(); });
        hall.addEventListener('keydown', event => { if(event.key==='Escape'){pinned=hovering=focused=false;expand();} });
        render();
        timer=setInterval(async () => {
            if(rows.length<2 || hall.classList.contains('is-open') || document.hidden || reduced_motion.matches) return;
            const old = stage.firstElementChild;
            try {
                await old.animate([{transform:'rotateX(0deg)',opacity:1},{transform:'rotateX(-90deg)',opacity:0}], {duration:200, easing:'ease-in',transformOrigin:'center bottom'}).finished;
                if (!stage.isConnected || hall.classList.contains('is-open') || reduced_motion.matches) return;
                index=(index+1)%rows.length;
                render().animate([{transform:'rotateX(90deg)',opacity:0},{transform:'rotateX(0deg)',opacity:1}],{duration:240,easing:'ease-out'});
            } catch (_) { /* Detached during refresh. */ }
        },4000);
    }
    async function refresh() {
        if (pending || document.hidden) return;
        pending=true;
        try {
            const response=await fetch('/clan/sidebar',{credentials:'same-origin',cache:'no-store'});
            if(!response.ok || response.redirected) throw new Error('sidebar unavailable');
            const html=await response.text();
            if(html!==last_html && !root.querySelector('.is-open')) {
                last_html=html;root.innerHTML=html;setup_hall();
            }
            update_times();
        } catch (_) { if(!last_html) root.textContent=root.dataset.error; }
        finally {pending=false;}
    }
    refresh();setInterval(refresh,30000);
    document.addEventListener('visibilitychange',()=>{if(!document.hidden)refresh();});
})();
