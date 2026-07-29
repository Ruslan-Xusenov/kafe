/* ==============================
   KafePlat — DaisyUI Landing JS
   ============================== */

// ---- NAVBAR SCROLL SHADOW ----
const navbar = document.querySelector('.navbar');
if (navbar) {
  window.addEventListener('scroll', () => {
    if (window.scrollY > 40) {
      navbar.classList.add('shadow-xl');
    } else {
      navbar.classList.remove('shadow-xl');
    }
  });
}

// ---- SCROLL REVEAL ----
const revealObserver = new IntersectionObserver((entries) => {
  entries.forEach((entry, i) => {
    if (entry.isIntersecting) {
      const delay = parseInt(entry.target.dataset.revealDelay || '0') * 100;
      setTimeout(() => {
        entry.target.classList.add('revealed');
      }, delay);
      revealObserver.unobserve(entry.target);
    }
  });
}, { threshold: 0.1, rootMargin: '0px 0px -50px 0px' });

document.querySelectorAll('[data-reveal]').forEach(el => revealObserver.observe(el));

// ---- BURGER MENU (mobile) ----
// DaisyUI handles this natively via <details> or dropdown
// Extra: close dropdown on link click
document.querySelectorAll('.dropdown a').forEach(link => {
  link.addEventListener('click', () => {
    const dropdown = link.closest('.dropdown');
    if (dropdown) {
      const btn = dropdown.querySelector('[tabindex="0"]');
      if (btn) btn.blur();
    }
  });
});

// ---- SMOOTH SCROLL ----
document.querySelectorAll('a[href^="#"]').forEach(a => {
  a.addEventListener('click', (e) => {
    const id = a.getAttribute('href').slice(1);
    if (!id) return;
    const el = document.getElementById(id);
    if (el) {
      e.preventDefault();
      el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  });
});

// ---- ACTIVE NAV HIGHLIGHT ----
const sections = document.querySelectorAll('section[id]');
const navLinks = document.querySelectorAll('.navbar a[href^="#"]');
window.addEventListener('scroll', () => {
  let current = '';
  sections.forEach(s => {
    if (window.scrollY >= s.offsetTop - 120) current = s.id;
  });
  navLinks.forEach(a => {
    a.classList.toggle('text-primary', a.getAttribute('href') === `#${current}`);
  });
});

// ---- LIVE ORDER FEED ----
const mockOrders = [
  { num: '#1045', name: "Manti × 4, Choy", status: 'badge-warning', label: 'Oshxonada' },
  { num: '#1046', name: "Non kabob × 2", status: 'badge-success', label: '✓ Tayyor' },
  { num: '#1047', name: "Shurpa + Non", status: 'badge-info', label: '🆕 Yangi', blink: true },
  { num: '#1048', name: "Lag'mon, Cola", status: 'badge-info', label: '🆕 Yangi', blink: true },
  { num: '#1049', name: "Somsa × 6", status: 'badge-warning', label: 'Oshxonada' },
];
let mockIdx = 0;

function addLiveOrder() {
  const container = document.getElementById('liveOrders');
  if (!container) return;
  const items = container.querySelectorAll('.order-row');
  if (items.length >= 3) {
    const oldest = items[items.length - 1];
    oldest.style.transition = 'all 0.3s ease';
    oldest.style.opacity = '0';
    oldest.style.transform = 'translateX(-10px)';
    setTimeout(() => oldest.remove(), 320);
  }
  const order = mockOrders[mockIdx++ % mockOrders.length];
  const div = document.createElement('div');
  div.className = `order-row flex items-center gap-2 p-2 rounded-lg bg-base-100/30 border border-base-content/5 mb-1.5 text-[11px]${order.blink ? ' new-order-blink' : ''}`;
  div.style.cssText = 'opacity:0; transform:translateX(10px); transition:all 0.4s ease;';
  div.innerHTML = `
    <span class="text-base-content/40 w-8">${order.num}</span>
    <span class="flex-1">${order.name}</span>
    <span class="badge ${order.status} badge-sm">${order.label}</span>
  `;
  const title = container.querySelector('div');
  if (title) {
    title.insertAdjacentElement('afterend', div);
  } else {
    container.appendChild(div);
  }
  requestAnimationFrame(() => requestAnimationFrame(() => {
    div.style.opacity = '1';
    div.style.transform = 'translateX(0)';
  }));
}
setInterval(addLiveOrder, 3000);

// ---- SCROLL TO TOP ----
const scrollBtn = document.createElement('button');
scrollBtn.innerHTML = '↑';
scrollBtn.setAttribute('aria-label', 'Yuqoriga');
scrollBtn.className = 'btn btn-primary btn-circle fixed bottom-6 right-6 z-50 shadow-xl shadow-primary/30';
scrollBtn.style.cssText = 'opacity:0; transform:translateY(12px); transition:all 0.3s ease; pointer-events:none;';
document.body.appendChild(scrollBtn);

window.addEventListener('scroll', () => {
  const show = window.scrollY > 500;
  scrollBtn.style.opacity = show ? '1' : '0';
  scrollBtn.style.transform = show ? 'translateY(0)' : 'translateY(12px)';
  scrollBtn.style.pointerEvents = show ? 'auto' : 'none';
});
scrollBtn.addEventListener('click', () => window.scrollTo({ top: 0, behavior: 'smooth' }));

// ---- STAGGER CARD ANIMATIONS ----
function staggerCards(selector, baseDelay = 80) {
  const cards = document.querySelectorAll(selector);
  const obs = new IntersectionObserver((entries) => {
    entries.forEach((entry, idx) => {
      if (entry.isIntersecting) {
        setTimeout(() => {
          entry.target.style.opacity = '1';
          entry.target.style.transform = 'translateY(0)';
        }, idx * baseDelay);
        obs.unobserve(entry.target);
      }
    });
  }, { threshold: 0.08 });

  cards.forEach(card => {
    card.style.opacity = '0';
    card.style.transform = 'translateY(24px)';
    card.style.transition = 'opacity 0.55s ease, transform 0.55s ease';
    obs.observe(card);
  });
}

// Apply stagger to grid cards (not data-reveal ones to avoid conflict)
staggerCards('.stats .stat', 80);

// ---- PRINTER POPUP PULSE ----
const popup = document.getElementById('printerPopup');
if (popup) {
  setInterval(() => {
    popup.style.transition = 'transform 0.3s ease, opacity 0.3s ease';
    popup.style.transform = 'scale(1.05)';
    popup.style.opacity = '0.7';
    setTimeout(() => {
      popup.style.transform = 'scale(1)';
      popup.style.opacity = '1';
    }, 400);
  }, 4000);
}

console.log('%c☕ KafePlat', 'font-size:22px;font-weight:900;color:#f97316;');
console.log('%cDaisyUI Light/Night Theme — Kafe Boshqaruv Tizimi', 'color:#64748b;');

// ---- THEME TOGGLE ----
const themeToggle = document.getElementById('themeToggle');
if (themeToggle) {
  const lightIcon = themeToggle.querySelector('.light-icon');
  const darkIcon = themeToggle.querySelector('.dark-icon');

  function updateIcons(theme) {
    if (theme === 'night') {
      lightIcon.classList.remove('hidden');
      darkIcon.classList.add('hidden');
    } else {
      lightIcon.classList.add('hidden');
      darkIcon.classList.remove('hidden');
    }
  }

  const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
  updateIcons(currentTheme);

  themeToggle.addEventListener('click', () => {
    const theme = document.documentElement.getAttribute('data-theme');
    const newTheme = theme === 'light' ? 'night' : 'light';
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateIcons(newTheme);
  });
}
