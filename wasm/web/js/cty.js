/*
 * Replaces the two bits of Bootstrap JS this app used: collapse toggles and the theme
 * switch. Everything is delegated off document, so it keeps working after go-app
 * replaces DOM nodes on re-render.
 */
(function () {
    'use strict';

    var THEME_KEY = 'cty-theme';

    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        try { localStorage.setItem(THEME_KEY, theme); } catch (e) {}
    }

    function toggleTheme() {
        var current = document.documentElement.getAttribute('data-theme');
        setTheme(current === 'dark' ? 'light' : 'dark');
    }

    function close(panel, button) {
        panel.classList.remove('open');
        if (button) button.setAttribute('aria-expanded', 'false');
    }

    // A panel's button is whichever control points at its id.
    function buttonFor(panel) {
        return document.querySelector('[data-target="#' + panel.id + '"]');
    }

    function toggleCollapse(button) {
        var selector = button.getAttribute('data-target');
        if (!selector) return;

        var panel = document.querySelector(selector);
        if (!panel) return;

        var willOpen = !panel.classList.contains('open');

        // data-parent means the panels in that container are mutually exclusive.
        var parentSelector = panel.getAttribute('data-parent');
        if (willOpen && parentSelector) {
            var parent = document.querySelector(parentSelector);
            if (parent) {
                parent.querySelectorAll('.collapse.open').forEach(function (other) {
                    if (other !== panel && other.getAttribute('data-parent') === parentSelector) {
                        close(other, buttonFor(other));
                    }
                });
            }
        }

        panel.classList.toggle('open', willOpen);
        button.setAttribute('aria-expanded', willOpen ? 'true' : 'false');
    }

    document.addEventListener('click', function (event) {
        var themeBtn = event.target.closest('[data-theme-toggle]');
        if (themeBtn) {
            event.preventDefault();
            toggleTheme();
            return;
        }

        var toggle = event.target.closest('[data-toggle="collapse"]');
        if (toggle) {
            event.preventDefault();
            toggleCollapse(toggle);
        }
    });

    // React to OS changes only while the user has not pinned a preference.
    var media = window.matchMedia('(prefers-color-scheme: dark)');
    var onChange = function (e) {
        var stored = null;
        try { stored = localStorage.getItem(THEME_KEY); } catch (err) {}
        if (!stored) document.documentElement.setAttribute('data-theme', e.matches ? 'dark' : 'light');
    };

    if (media.addEventListener) {
        media.addEventListener('change', onChange);
    } else if (media.addListener) {
        media.addListener(onChange);
    }
})();
