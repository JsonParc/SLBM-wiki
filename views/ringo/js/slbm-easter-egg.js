"use strict";
(() => {
    const path = decodeURIComponent(window.location.pathname).replace(/\/+$/, '');
    const effect_settings = {
        '/w/JsonParc': {name: 'JsonParc', hue: 28, error: '[0x11 0000 0000] ERROR: Connection failed: Connection refused.'},
        '/JsonParc': {name: 'JsonParc', hue: 28, error: '[0x11 0000 0000] ERROR: Connection failed: Connection refused.'},
        '/w/뉴비에요': {name: '뉴비에요', hue: 2, error: '[0x11 000 000] ERROR: SuSnubaNohopeDja: Connection Refused'},
        '/뉴비에요': {name: '뉴비에요', hue: 2, error: '[0x11 000 000] ERROR: SuSnubaNohopeDja: Connection Refused'},
    };
    const effect = effect_settings[path];
    if (!effect) return;

    const glyphs = 'アカサタナハマヤラワ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ<>+-*/';
    let clicks = 0;
    let reset_timer;
    let canvas;
    let animation_frame;
    let terminal_overlay;

    function find_trigger() {
        const profile_table = document.querySelector('#main_data table.slbm-member-info');
        const header = profile_table?.querySelector('tbody > tr:first-child > td');
        return header?.textContent.trim() === effect.name ? header : null;
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
        const font_size = 14 * ratio;
        const columns = Math.ceil(canvas.width / font_size);
        const positions = Array.from({length: columns}, () => -Math.random() * 45);
        const finished = Array.from({length: columns}, () => false);
        const stream_lengths = Array.from({length: columns}, (_, index) => 26 + index % 10);
        const opacities = Array.from({length: columns}, (_, index) => 0.48 + (index * 13 % 15) / 100);
        let passes = 0;
        let cleanup_timer;

        function finish() {
            window.cancelAnimationFrame(animation_frame);
            window.clearTimeout(cleanup_timer);
            window.removeEventListener('resize', resize);
            canvas.remove();
            canvas = null;
            start_terminal_sequence();
        }

        function draw() {
            context.clearRect(0, 0, canvas.width, canvas.height);
            context.font = `${font_size}px monospace`;
            positions.forEach((position, index) => {
                if (finished[index]) return;
                const hue = effect.hue + index % 13;
                const brightness = 48 + (index * 17 % 10);
                for (let trail = 0; trail < stream_lengths[index]; trail++) {
                    const character = glyphs[Math.floor(Math.random() * glyphs.length)];
                    const y = (position - trail) * font_size;
                    if (y < -font_size || y > canvas.height + font_size) continue;
                    context.fillStyle = `hsla(${hue}, 100%, ${brightness + (index % 7 === 0 ? 12 : 0)}%, ${opacities[index]})`;
                    context.fillText(character, index * font_size, y);
                }
                positions[index] = position + 0.86;
                if (position * font_size > canvas.height + font_size * 2) {
                    finished[index] = true;
                }
            });
            if (finished.every(Boolean)) {
                passes += 1;
                if (passes < 2) {
                    positions.forEach((_, index) => {
                        positions[index] = -Math.random() * 45;
                        finished[index] = false;
                    });
                    animation_frame = window.requestAnimationFrame(draw);
                    return;
                }
                finish();
                return;
            }
            animation_frame = window.requestAnimationFrame(draw);
        }
        cleanup_timer = window.setTimeout(finish, 4000);
        draw();
    }

    function start_terminal_sequence() {
        if (terminal_overlay) return;
        terminal_overlay = document.createElement('div');
        terminal_overlay.className = 'slbm-terminal-overlay';
        terminal_overlay.innerHTML = '<div class="slbm-terminal-window"><div class="slbm-terminal-title">C:\\Windows\\System32\\cmd.exe</div><div class="slbm-terminal-output"></div></div>';
        document.body.appendChild(terminal_overlay);

        const output = terminal_overlay.querySelector('.slbm-terminal-output');
        const error_line = effect.error;
        const connecting_line = 'connecting...';
        const slow_line_count = 5;
        const total_line_count = 55;
        let line_count = 0;
        let output_timer;
        let finish_timer;

        function append_line() {
            const line = document.createElement('div');
            line.className = 'slbm-terminal-line';
            line.textContent = line_count % 2 === 0 ? error_line : connecting_line;
            line.style.setProperty('--glitch-delay', `${(line_count % 7) * 0.035}s`);
            output.appendChild(line);
            output.scrollTop = output.scrollHeight;
            line_count += 1;
            if (line_count >= total_line_count) {
                window.clearTimeout(output_timer);
                terminal_overlay.classList.add('is-noisy');
                finish_timer = window.setTimeout(() => {
                    terminal_overlay.classList.add('is-flash');
                    window.setTimeout(() => terminal_overlay.classList.add('is-black'), 180);
                    window.setTimeout(() => {
                        terminal_overlay.remove();
                        terminal_overlay = null;
                    }, 1200);
                }, 900);
                return;
            }
            output_timer = window.setTimeout(append_line, line_count < slow_line_count ? 225 : 28);
        }

        append_line();
        window.addEventListener('beforeunload', () => {
            window.clearTimeout(output_timer);
            window.clearTimeout(finish_timer);
        }, {once: true});
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