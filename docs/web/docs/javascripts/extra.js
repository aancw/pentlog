/**
 * PentLog Documentation - Extra JavaScript
 * Adds animations and interactive enhancements
 */

(function() {
  'use strict';

  /**
   * Initialize scroll animations for elements with fade-in class
   */
  function initScrollAnimations() {
    const fadeElements = document.querySelectorAll('.fade-in');

    if (!fadeElements.length) return;

    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
          observer.unobserve(entry.target);
        }
      });
    }, {
      threshold: 0.1,
      rootMargin: '0px 0px -50px 0px'
    });

    fadeElements.forEach(el => observer.observe(el));
  }

  /**
   * Add copy-to-clipboard functionality enhancement
   */
  function initCodeCopyEnhancement() {
    document.querySelectorAll('.md-typeset pre > code').forEach(codeBlock => {
      const pre = codeBlock.parentElement;

      // Add hover effect to show copy hint
      pre.addEventListener('mouseenter', () => {
        pre.style.cursor = 'pointer';
      });
    });
  }

  /**
   * Initialize smooth scroll for anchor links
   */
  function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
      anchor.addEventListener('click', function(e) {
        const targetId = this.getAttribute('href');
        if (targetId === '#') return;

        const targetElement = document.querySelector(targetId);
        if (targetElement) {
          e.preventDefault();
          targetElement.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
          });
        }
      });
    });
  }

  /**
   * Add active state highlighting for table rows
   */
  function initTableRowHighlight() {
    document.querySelectorAll('.md-typeset table tbody tr').forEach(row => {
      row.addEventListener('mouseenter', () => {
        row.style.transition = 'background-color 0.2s ease';
      });
    });
  }

  /**
   * Initialize hero animation on page load
   */
  function initHeroAnimation() {
    const hero = document.querySelector('.pentlog-hero');
    if (!hero) return;

    // Add visible class after a short delay for entrance animation
    setTimeout(() => {
      hero.classList.add('visible');
    }, 100);
  }

  /**
   * Add ripple effect to buttons
   */
  function initButtonRipple() {
    document.querySelectorAll('.md-button, .pentlog-hero__btn').forEach(button => {
      button.addEventListener('click', function(e) {
        const ripple = document.createElement('span');
        const rect = this.getBoundingClientRect();
        const size = Math.max(rect.width, rect.height);
        const x = e.clientX - rect.left - size / 2;
        const y = e.clientY - rect.top - size / 2;

        ripple.style.cssText = `
          position: absolute;
          border-radius: 50%;
          background: rgba(255, 255, 255, 0.3);
          width: ${size}px;
          height: ${size}px;
          left: ${x}px;
          top: ${y}px;
          transform: scale(0);
          animation: ripple 0.6s ease-out;
          pointer-events: none;
        `;

        this.style.position = 'relative';
        this.style.overflow = 'hidden';
        this.appendChild(ripple);

        setTimeout(() => ripple.remove(), 600);
      });
    });

    // Add ripple keyframe animation
    const style = document.createElement('style');
    style.textContent = `
      @keyframes ripple {
        to {
          transform: scale(2);
          opacity: 0;
        }
      }
    `;
    document.head.appendChild(style);
  }

  /**
   * Initialize typewriter effect for terminal windows
   */
  function initTypewriterEffect() {
    const terminals = document.querySelectorAll('.terminal-window');

    terminals.forEach(terminal => {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            terminal.classList.add('terminal-visible');
            observer.unobserve(terminal);
          }
        });
      }, { threshold: 0.3 });

      observer.observe(terminal);
    });
  }

  /**
   * Keyboard shortcuts helper
   */
  function initKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
      // Press '?' to show keyboard shortcuts
      if (e.key === '?' && !e.target.matches('input, textarea')) {
        e.preventDefault();
        showKeyboardShortcuts();
      }

      // Press '/' to focus search
      if (e.key === '/' && !e.target.matches('input, textarea')) {
        e.preventDefault();
        const searchInput = document.querySelector('.md-search__input');
        if (searchInput) {
          searchInput.focus();
        }
      }
    });
  }

  /**
   * Show keyboard shortcuts modal
   */
  function showKeyboardShortcuts() {
    const modal = document.createElement('div');
    modal.className = 'keyboard-shortcuts-modal';
    modal.innerHTML = `
      <div class="keyboard-shortcuts-overlay" style="
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.6);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
      ">
        <div class="keyboard-shortcuts-content" style="
          background: var(--md-default-bg-color);
          padding: 2rem;
          border-radius: 12px;
          max-width: 400px;
          width: 90%;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        ">
          <h3 style="margin-top: 0;">Keyboard Shortcuts</h3>
          <div style="display: grid; gap: 0.75rem; margin: 1rem 0;">
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>Focus search</span>
              <kbd style="
                background: var(--md-code-bg-color);
                padding: 0.25rem 0.5rem;
                border-radius: 4px;
                font-family: monospace;
                font-size: 0.85rem;
              ">/</kbd>
            </div>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>Show shortcuts</span>
              <kbd style="
                background: var(--md-code-bg-color);
                padding: 0.25rem 0.5rem;
                border-radius: 4px;
                font-family: monospace;
                font-size: 0.85rem;
              ">?</kbd>
            </div>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>Previous page</span>
              <kbd style="
                background: var(--md-code-bg-color);
                padding: 0.25rem 0.5rem;
                border-radius: 4px;
                font-family: monospace;
                font-size: 0.85rem;
              ">←</kbd>
            </div>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>Next page</span>
              <kbd style="
                background: var(--md-code-bg-color);
                padding: 0.25rem 0.5rem;
                border-radius: 4px;
                font-family: monospace;
                font-size: 0.85rem;
              ">→</kbd>
            </div>
          </div>
          <button onclick="this.closest('.keyboard-shortcuts-modal').remove()" style="
            width: 100%;
            padding: 0.75rem;
            background: var(--md-accent-fg-color);
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 500;
          ">Close</button>
        </div>
      </div>
    `;

    document.body.appendChild(modal);

    // Close on overlay click
    modal.querySelector('.keyboard-shortcuts-overlay').addEventListener('click', (e) => {
      if (e.target === e.currentTarget) {
        modal.remove();
      }
    });

    // Close on Escape
    const closeOnEscape = (e) => {
      if (e.key === 'Escape') {
        modal.remove();
        document.removeEventListener('keydown', closeOnEscape);
      }
    };
    document.addEventListener('keydown', closeOnEscape);
  }

  /**
   * Initialize progress indicator for reading
   */
  function initReadingProgress() {
    const article = document.querySelector('.md-content__inner');
    if (!article) return;

    const progressBar = document.createElement('div');
    progressBar.className = 'reading-progress';
    progressBar.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      height: 3px;
      background: linear-gradient(90deg, var(--md-accent-fg-color), var(--md-accent-fg-color--light));
      z-index: 1001;
      transition: width 0.1s ease;
      width: 0%;
    `;

    document.body.appendChild(progressBar);

    window.addEventListener('scroll', () => {
      const scrollTop = window.scrollY;
      const docHeight = article.offsetHeight;
      const winHeight = window.innerHeight;
      const scrollPercent = (scrollTop / (docHeight - winHeight)) * 100;
      progressBar.style.width = Math.min(scrollPercent, 100) + '%';
    });
  }

  /**
   * Initialize all enhancements when DOM is ready
   */
  function init() {
    // Wait for Material for MkDocs to be ready
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', initEnhancements);
    } else {
      initEnhancements();
    }
  }

  function initEnhancements() {
    initScrollAnimations();
    initCodeCopyEnhancement();
    initSmoothScroll();
    initTableRowHighlight();
    initHeroAnimation();
    initButtonRipple();
    initTypewriterEffect();
    initKeyboardShortcuts();
    initReadingProgress();

    // Log initialization for debugging
    console.log('🎯 PentLog Documentation enhancements loaded');
  }

  // Start initialization
  init();

})();
