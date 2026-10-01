"use strict";
(() => {
    const path = decodeURIComponent(window.location.pathname).replace(/\/+$/, '');
    if (path !== '/w/JsonParc' && path !== '/JsonParc') return;

    const glyphs = 'アカサタナハマヤラワ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<>+-*/';
    let clicks = 0;
    let reset_timer;
    let canvas;
    let animation_frame;

    function find_trigger() {
        const profile_table = document.querySelector('#main_data table.slbm-member-info');
        const header = profile_table?.querySelector('tbody > tr:first-child > td');
        return header?.textContent.trim() === 'JsonParc' ? header : null;
    }

    function resize() {
        const ratio = window.devicePixelRatio || 1;
        canvas.width = Math.floor(window.innerWidth * ratio);
        canvas.height = Math.floor(window.innerHeight * ratio);
        canvas.style.width = `${window.innerWidth}px`;
        canvas.style.height = `${window.innerHeight}px`;
    }

    function start_matrix() {
        if (canvas) return;
        canvas = document.createElement('canvas');
        canvas.className = 'slbm-matrix-overlay';
        canvas.setAttribute('aria-hidden', 'true');
        document.body.appendChild(canvas);
        resize();
        window.addEventListener('resize', resize);

        const context = canvas.getContext('2d');
        const ratio = window.devicePixelRatio || 1;
        const font_size = 16 * ratio;
        const columns = Math.ceil(canvas.width / font_size);
        const positions = Array.from({length: columns}, () => -Math.random() * 45);
        const finished = Array.from({length: columns}, () => false);
        let cleanup_timer;

        function finish() {
            window.cancelAnimationFrame(animation_frame);
            window.clearTimeout(cleanup_timer);
            window.removeEventListener('resize', resize);
            canvas.remove();
            canvas = null;
        }

        function draw() {
            context.fillStyle = 'rgba(0, 0, 0, 0.035)';
            context.fillRect(0, 0, canvas.width, canvas.height);
            context.font = `${font_size}px monospace`;
            positions.forEach((position, index) => {
                const character = glyphs[Math.floor(Math.random() * glyphs.length)];
                if (finished[index]) return;
                context.fillStyle = index % 7 === 0 ? 'rgba(215, 255, 207, 0.28)' : 'rgba(32, 232, 59, 0.18)';
                context.fillText(character, index * font_size, position * font_size);
                positions[index] = position + 0.86;
                if (position * font_size > canvas.height + font_size * 2) {
                    finished[index] = true;
                }
            });
            if (finished.every(Boolean)) {
                finish();
                return;
            }
            animation_frame = window.requestAnimationFrame(draw);
        }
        cleanup_timer = window.setTimeout(finish, 4000);
        draw();
    }

    function setup() {
        const trigger = find_trigger();
        if (!trigger) return;
        trigger.classList.add('slbm-matrix-trigger');
        trigger.setAttribute('title', '...');
        trigger.addEventListener('click', () => {
            clicks += 1;
            window.clearTimeout(reset_timer);
            reset_timer = window.setTimeout(() => { clicks = 0; }, 2000);
            if (clicks >= 10) start_matrix();
        });
    }

    window.addEventListener('beforeunload', () => {
        window.cancelAnimationFrame(animation_frame);
        window.removeEventListener('resize', resize);
    });
    window.addEventListener('DOMContentLoaded', setup);
})();