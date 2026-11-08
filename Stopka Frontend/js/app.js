
function $(elementId) { return document.getElementById(elementId); }


function h(tag, attrs, ...children) {
  const el = document.createElement(tag);
  for (const [key, value] of Object.entries(attrs || {})) {
    if (value === null || value === undefined || value === false) continue;
    if (key.startsWith('on')) el.addEventListener(key.slice(2), value);
    else if (key === 'class') el.className = value;
    else if (key === 'style') el.style.cssText = value;
    else if (value === true) el.setAttribute(key, '');
    else el.setAttribute(key, value);
  }
  for (const child of children.flat()) {
    if (child === null || child === undefined || child === false) continue;
    el.append(child instanceof Node ? child : document.createTextNode(String(child)));
  }
  return el;
}


const ICONS = {
  home: '<path d="M3 11l9-8 9 8"/><path d="M5 10v10h14V10"/>',
  decks: '<rect x="3" y="5" width="14" height="16" rx="2"/><path d="M7 3h12a2 2 0 0 1 2 2v12"/>',
  duel: '<path d="M14.5 17.5L3 6V3h3l11.5 11.5"/><path d="M13 19l6-6"/><path d="M16 16l4 4"/><path d="M19 21l2-2"/><path d="M9.5 6.5L13 3h3v3l-3.5 3.5"/><path d="M5 14l4 4"/><path d="M7 17l-3 3"/><path d="M3 19l2 2"/>',
  class: '<circle cx="9" cy="8" r="3.5"/><path d="M2.5 20a6.5 6.5 0 0 1 13 0"/><path d="M16 4.5a3.5 3.5 0 0 1 0 7"/><path d="M18 14a6.5 6.5 0 0 1 3.5 6"/>',
  profile: '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
  back: '<path d="M15 6l-6 6 6 6"/>',
  next: '<path d="M9 6l6 6-6 6"/>',
  moon: '<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z"/>',
  test: '<path d="M9 11l3 3 8-8"/><path d="M20 12v7a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h9"/>',
  repeat: '<path d="M17 2l4 4-4 4"/><path d="M3 11V9a3 3 0 0 1 3-3h15"/><path d="M7 22l-4-4 4-4"/><path d="M21 13v2a3 3 0 0 1-3 3H3"/>',
  copy: '<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"/>',
  plus: '<path d="M12 5v14"/><path d="M5 12h14"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/>',
  teacher: '<path d="M2 9l10-5 10 5-10 5z"/><path d="M6 11v5c3 2 9 2 12 0v-5"/>'
};
function icon(name) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('viewBox', '0 0 24 24');
  svg.setAttribute('aria-hidden', 'true');
  svg.innerHTML = ICONS[name]; 
  return svg;
}


function plural(n, one, few, many) {
  const n10 = n % 10, n100 = n % 100;
  if (n10 === 1 && n100 !== 11) return one;
  if (n10 >= 2 && n10 <= 4 && (n100 < 12 || n100 > 14)) return few;
  return many;
}
function countText(n, one, few, many) { return n + ' ' + plural(n, one, few, many); }

function formatDate(iso) {
  if (!iso) return '—';
  const date = new Date(iso);
  const today = new Date();
  const days = Math.floor((new Date(today.toDateString()) - new Date(date.toDateString())) / 86400000);
  if (days === 0) return 'сегодня';
  if (days === 1) return 'вчера';
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}
function percentText(value) { return value === null || value === undefined ? '—' : value + '%'; }


function header(title, subtitle, action) {
  return h('header', { class: 'top' },
    stack.length > 1 ? h('button', { class: 'icon-btn', 'aria-label': 'Назад', onclick: back }, icon('back')) : null,
    h('div', { class: 'top__text' },
      h('h1', { class: 'title' }, title),
      subtitle ? h('p', { class: 'muted' }, subtitle) : null),
    action || null);
}

function progressBar(part, total) {
  const width = total ? Math.min(100, Math.round(part / total * 100)) : 0;
  return h('span', { class: 'progress', role: 'progressbar', 'aria-valuemin': 0, 'aria-valuemax': total, 'aria-valuenow': part },
    h('span', { class: 'progress__bar', style: `width:${width}%` }));
}

function statTiles(items) {
  return h('div', { class: 'stats' },
    items.map(([value, label, tone]) => h('div', { class: 'stat' },
      h('b', { class: tone || '' }, value), h('span', { class: 'muted' }, label))));
}

function sectionTitle(text) { return h('h2', { class: 'subtitle' }, text); }

function stateBox(text, buttonText, onClick) {
  return h('div', { class: 'state' },
    h('p', { class: 'muted' }, text),
    buttonText ? h('button', { class: 'btn btn--main btn--small', onclick: onClick }, buttonText) : null);
}
function loadingState() { return h('div', { class: 'state state--loading' }, h('span', { class: 'spinner', 'aria-hidden': 'true' }), h('p', { class: 'muted' }, 'Загрузка…')); }

function errorText(err) {
  return err instanceof ApiError ? err.message : 'Что-то пошло не так';
}


async function load(container, loader, draw) {
  const token = navToken;
  container.replaceChildren(loadingState());
  try {
    const data = await loader();
    if (token !== navToken) return;
    container.replaceChildren();
    draw(data);
  } catch (err) {
    if (token !== navToken) return;
    if (!(err instanceof ApiError)) console.error(err);
    if (err.status === 401) return showFatal('Сессия устарела', 'Закрой приложение и снова открой его из чата с ботом.');
    container.replaceChildren(stateBox(errorText(err), 'Попробовать снова', () => load(container, loader, draw)));
  }
}


function actionButton(label, className, action) {
  const error = h('p', { class: 'error', hidden: true });
  const btn = h('button', { class: className, onclick: run }, label);
  async function run() {
    if (btn.disabled) return;
    btn.disabled = true;
    error.hidden = true;
    try {
      await action();
    } catch (err) {
      if (!(err instanceof ApiError)) console.error(err);
      error.textContent = errorText(err);
      error.hidden = false;
    } finally {
      btn.disabled = false;
    }
  }
  return h('div', { class: 'action' }, btn, error);
}


function confirmButton(label, confirmLabel, action) {
  let armed = false, timer = null;
  const wrap = actionButton(label, 'btn btn--danger', async () => {
    if (!armed) {
      armed = true;
      btn.textContent = confirmLabel;
      timer = setTimeout(() => { armed = false; btn.textContent = label; }, 4000);
      return;
    }
    clearTimeout(timer);
    await action();
  });
  const btn = wrap.querySelector('button');
  return wrap;
}

function toast(text) {
  const el = $('toast');
  el.textContent = text;
  el.hidden = false;
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => { el.hidden = true; }, 2500);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    toast('Скопировано');
  } catch (e) {
    toast('Не получилось скопировать. Выдели текст и скопируй вручную');
  }
}


function codeBox(code, hint) {
  return h('div', { class: 'code-box' },
    h('span', { class: 'code-box__code', 'aria-label': 'Код ' + code.split('').join(' ') }, code),
    hint ? h('p', { class: 'muted' }, hint) : null,
    h('button', { class: 'btn btn--light btn--small', onclick: () => copyText(code) }, icon('copy'), 'Скопировать код'));
}


const screens = {};
let stack = [];
let navToken = 0;
let screenCleanup = null;


function defineScreen(name, tab, render) {
  screens[name] = { tab, render };
}

function go(name, params, options) {
  const entry = { name, params: params || {} };
  if (options && options.root) stack = [entry];
  else if (options && options.replace) stack[stack.length - 1] = entry;
  else stack.push(entry);
  renderScreen();
}

function back() {
  if (stack.length < 2) return;
  stack.pop();
  renderScreen();
}

function renderScreen() {
  if (screenCleanup) screenCleanup();
  screenCleanup = null;
  navToken++;

  const { name, params } = stack[stack.length - 1];
  const screen = screens[name];
  const view = $('view');
  view.replaceChildren();
  window.scrollTo(0, 0);
  highlightMenu(screen.tab);
  updateMaxBackButton();

  
  const cleanup = screen.render(view, params);
  if (typeof cleanup === 'function') screenCleanup = cleanup;
}


function updateMaxBackButton() {
  const bb = window.WebApp && window.WebApp.BackButton;
  if (!bb) return;
  if (stack.length > 1) bb.show(); else bb.hide();
}


const MENUS = {
  student: [
    ['home', 'home', 'Главная'],
    ['decks', 'decks', 'Колоды'],
    ['duels', 'duel', 'Дуэли'],
    ['myClasses', 'class', 'Класс'],
    ['profile', 'profile', 'Профиль']
  ],
  teacher: [
    ['teacherHome', 'class', 'Классы'],
    ['profile', 'profile', 'Профиль']
  ]
};

function buildMenu() {
  const items = MENUS[me.role] || [];
  $('menuItems').replaceChildren(...items.map(([screen, iconName, label]) =>
    h('button', { class: 'menu__item', 'data-tab': screen, onclick: () => go(screen, {}, { root: true }) },
      icon(iconName), label)));
  $('menuItems').style.setProperty('--items', items.length);
  $('menu').hidden = items.length === 0;
}

function highlightMenu(tab) {
  document.querySelectorAll('.menu__item').forEach(btn => {
    const active = btn.dataset.tab === tab;
    btn.classList.toggle('is-active', active);
    if (active) btn.setAttribute('aria-current', 'page'); else btn.removeAttribute('aria-current');
  });
}


let me = null;

function showFatal(title, text, retry) {
  navToken++;
  $('menu').hidden = true;
  $('view').replaceChildren(
    h('div', { class: 'fatal' },
      h('span', { class: 'logo-icon logo-icon--big', 'aria-hidden': 'true' }, icon('decks')),
      h('h1', { class: 'title' }, title),
      h('p', { class: 'muted' }, text),
      retry ? h('button', { class: 'btn btn--main', onclick: retry }, 'Попробовать снова') : null));
}

async function start() {
  applySavedTheme();
  $('demoNote').hidden = !DEMO_MODE;

  const bb = window.WebApp && window.WebApp.BackButton;
  if (bb) bb.onClick(back);

 
  if (!DEMO_MODE && !DEV_WITHOUT_MAX && !maxInitData()) {
    return showFatal('Открой «Стопку» в MAX', 'Приложение работает только внутри MAX. Найди нашего бота и нажми кнопку «Открыть» в чате.');
  }

  $('view').replaceChildren(loadingState());
  try {
    me = await API.me();
  } catch (err) {
    if (err.status === 401) {
      return showFatal('Не удалось подтвердить вход', 'Закрой приложение и снова открой его из чата с ботом в MAX.');
    }
    return showFatal('Не удалось загрузить', errorText(err), start);
  }
  enterApp();
}

function enterApp() {
  if (!me.role) {
    $('menu').hidden = true;
    go('role', { first: true }, { root: true });
    return;
  }
  buildMenu();
  if (me.role === 'teacher') {
    go('teacherHome', {}, { root: true });
    return;
  }
  go('home', {}, { root: true });

 
  const startParam = window.WebApp && window.WebApp.initDataUnsafe && window.WebApp.initDataUnsafe.start_param;
  if (startParam && startParam.startsWith('duel_') && !enterApp.startParamUsed) {
    enterApp.startParamUsed = true;
    go('duelJoin', { code: startParam.slice(5) });
  }
}

function userName() {
  return (me && me.user && me.user.first_name) || 'друг';
}

defineScreen('role', null, (view, params) => {
  let picked = null;
  const area = h('div', { class: 'stack' });

  function option(role, title, text, iconName) {
    return h('button', {
      class: 'role-card' + (me.role === role ? ' is-current' : ''),
      'aria-pressed': 'false',
      onclick: e => pick(role, e.currentTarget)
    },
      h('span', { class: 'role-card__icon' }, icon(iconName)),
      h('span', null, h('b', null, title), h('span', { class: 'muted' }, text)));
  }

  const studentBtn = option('student', 'Я ученик', 'Учу карточки, прохожу тесты и дуэли', 'decks');
  const teacherBtn = option('teacher', 'Я учитель', 'Создаю классы и смотрю успехи учеников', 'teacher');

  function pick(role, btn) {
    picked = role;
    [studentBtn, teacherBtn].forEach(b => b.setAttribute('aria-pressed', String(b === btn)));
    drawConfirm();
  }

  async function save(inviteCode) {
    me = await API.setRole(picked, inviteCode);
    toast(me.role === 'teacher' ? 'Ты вошёл как учитель' : 'Ты вошёл как ученик');
    enterApp();
  }

  function drawConfirm() {
    area.replaceChildren();
    if (picked === 'teacher' && !me.teacher_verified) {
      
      const input = h('input', { class: 'input', id: 'inviteCode', autocomplete: 'off', placeholder: 'Код учителя', maxlength: 32 });
      area.append(h('form', { class: 'panel', onsubmit: e => e.preventDefault() },
        h('label', { for: 'inviteCode' }, h('b', null, 'Код учителя')),
        h('p', { class: 'muted' }, 'Чтобы ученики не видели чужую статистику, учительский доступ открывается только по коду. Его выдаёт команда «Стопки».'),
        DEMO_MODE ? h('p', { class: 'hint' }, 'Демо-код: ' + DEMO_INVITE_CODE) : null,
        input,
        actionButton('Подтвердить', 'btn btn--main', async () => {
          const code = input.value.trim();
          if (!code) throw new ApiError(400, 'empty', 'Введи код учителя');
          await save(code);
        })));
      input.focus();
    } else {
      area.append(actionButton('Продолжить', 'btn btn--main', () => save()));
    }
  }

  const first = params.first || !me.role;
  view.append(
    first
      ? h('div', { class: 'welcome' },
          h('span', { class: 'logo-icon logo-icon--big', 'aria-hidden': 'true' }, icon('decks')),
          h('h1', { class: 'title' }, 'Привет, ' + userName() + '!'),
          h('p', { class: 'muted' }, '«Стопка» помогает учить историю России: карточки, тесты и дуэли с одноклассниками. Кто ты?'))
      : header('Сменить роль', 'Твои данные сохранятся'),
    studentBtn, teacherBtn, area);
});


const darkQuery = window.matchMedia('(prefers-color-scheme: dark)');
function isDark() {
  const t = document.documentElement.dataset.theme;
  return t === 'dark' || (!t && darkQuery.matches);
}
function setTheme(dark) {
  document.documentElement.dataset.theme = dark ? 'dark' : 'light';
  try { localStorage.setItem('theme', dark ? 'dark' : 'light'); } catch (e) {}
}
function applySavedTheme() {
  try {
    const saved = localStorage.getItem('theme');
    if (saved) document.documentElement.dataset.theme = saved;
  } catch (e) {}
}


function settingsPanel() {
  const label = h('span', { id: 'themeLabel' }, 'Тёмная тема');
  const sw = h('button', {
    class: 'switch', role: 'switch', 'aria-checked': String(isDark()), 'aria-labelledby': 'themeLabel',
    onclick: () => { setTheme(!isDark()); sw.setAttribute('aria-checked', String(isDark())); }
  }, h('span'));

  return h('div', { class: 'panel' },
    h('div', { class: 'row' }, label, sw),
    h('button', { class: 'row row--link', onclick: () => go('role') },
      h('span', null, 'Сменить роль'), h('span', { class: 'muted' }, me.role === 'teacher' ? 'сейчас учитель' : 'сейчас ученик')),
    
    confirmButton('Удалить мои данные', 'Удалить всё безвозвратно? Нажми ещё раз', async () => {
      await API.deleteMe();
      enterApp.startParamUsed = false;
      toast('Данные удалены');
      await start();
    }));
}
