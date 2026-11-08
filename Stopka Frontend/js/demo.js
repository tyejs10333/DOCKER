
const DEMO_INVITE_CODE = 'DEMO-UCHITEL';
const DEMO_STORAGE_KEY = 'stopka-demo-v1';


const DEMO_DECKS = [
  {
    id: 1, title: 'Крещение Руси', topic: 'Древняя Русь', period: 'X век',
    cards: [
      ['В каком году произошло крещение Руси?', '988'],
      ['Какой князь крестил Русь?', 'Владимир Святославич'],
      ['Из какого государства на Русь пришло христианство?', 'Из Византии'],
      ['Какая княгиня приняла крещение раньше Владимира?', 'Ольга, его бабушка'],
      ['Как Владимира прозвали в былинах?', 'Красно Солнышко']
    ],
    questions: [
      ['В каком году произошло крещение Руси?', ['862', '988', '1054', '1147'], 1],
      ['Какой князь крестил Русь?', ['Олег', 'Ярослав Мудрый', 'Владимир Святославич', 'Игорь'], 2],
      ['Откуда на Русь пришло христианство?', ['Из Византии', 'Из Священной Римской империи', 'Из Хазарского каганата', 'Из Волжской Булгарии'], 0],
      ['Какая княгиня крестилась раньше Владимира?', ['Рогнеда', 'Анна', 'Ольга', 'Ирина'], 2]
    ]
  },
  {
    id: 2, title: 'Куликовская битва', topic: 'Московская Русь', period: 'XIV век',
    cards: [
      ['В каком году была Куликовская битва?', '1380'],
      ['Кто возглавил русское войско?', 'Московский князь Дмитрий Иванович (Донской)'],
      ['Кто командовал ордынским войском?', 'Темник Мамай'],
      ['Где произошла битва?', 'На Куликовом поле, у впадения Непрядвы в Дон'],
      ['Кто, по преданию, благословил Дмитрия на битву?', 'Сергий Радонежский']
    ],
    questions: [
      ['В каком году была Куликовская битва?', ['1240', '1380', '1480', '1242'], 1],
      ['Кто возглавил русское войско?', ['Иван Калита', 'Александр Невский', 'Дмитрий Донской', 'Иван III'], 2],
      ['Кто командовал ордынским войском?', ['Батый', 'Мамай', 'Тохтамыш', 'Ахмат'], 1],
      ['Кто, по преданию, благословил Дмитрия на битву?', ['Сергий Радонежский', 'Андрей Рублёв', 'Нил Сорский', 'Митрополит Филипп'], 0]
    ]
  },
  {
    id: 3, title: 'Эпоха Петра I', topic: 'Российская империя', period: 'XVIII век',
    cards: [
      ['В каком году основан Санкт-Петербург?', '1703'],
      ['В каком году была Полтавская битва?', '1709'],
      ['С какой страной Россия вела Северную войну?', 'Со Швецией'],
      ['Каким миром закончилась Северная война?', 'Ништадтским миром, 1721'],
      ['Какой титул Пётр принял в 1721 году?', 'Император']
    ],
    questions: [
      ['В каком году основан Санкт-Петербург?', ['1700', '1703', '1709', '1721'], 1],
      ['В каком году была Полтавская битва?', ['1700', '1709', '1714', '1721'], 1],
      ['С кем Россия воевала в Северной войне?', ['С Османской империей', 'С Польшей', 'Со Швецией', 'С Пруссией'], 2],
      ['Каким миром закончилась Северная война?', ['Ништадтским', 'Тильзитским', 'Брестским', 'Кючук-Кайнарджийским'], 0]
    ]
  },
  {
    id: 4, title: 'Отмена крепостного права', topic: 'Российская империя', period: 'XIX век',
    cards: [
      ['В каком году отменили крепостное право?', '1861'],
      ['Какой император отменил крепостное право?', 'Александр II'],
      ['Каким документом объявили об отмене?', 'Манифестом 19 февраля 1861 года'],
      ['Как называли крестьян, которые после реформы ещё несли повинности до выкупа земли?', 'Временнообязанные'],
      ['Что платили крестьяне государству за полученную землю?', 'Выкупные платежи']
    ],
    questions: [
      ['В каком году отменили крепостное право?', ['1825', '1861', '1864', '1905'], 1],
      ['Какой император отменил крепостное право?', ['Николай I', 'Александр I', 'Александр II', 'Александр III'], 2],
      ['Каким документом объявили об отмене?', ['Манифестом 17 октября', 'Манифестом 19 февраля', 'Соборным уложением', 'Жалованной грамотой дворянству'], 1],
      ['Как называли крестьян, которые до выкупа земли несли повинности?', ['Временнообязанные', 'Однодворцы', 'Холопы', 'Посессионные'], 0]
    ]
  },
  {
    id: 5, title: 'Первый полёт в космос', topic: 'СССР', period: 'XX век',
    cards: [
      ['Когда человек впервые полетел в космос?', '12 апреля 1961 года'],
      ['Кто был первым космонавтом?', 'Юрий Гагарин'],
      ['Как назывался корабль первого полёта?', '«Восток-1»'],
      ['Сколько длился первый полёт?', 'Около 108 минут'],
      ['Когда запустили первый искусственный спутник Земли?', '4 октября 1957 года']
    ],
    questions: [
      ['Когда человек впервые полетел в космос?', ['4 октября 1957', '12 апреля 1961', '20 июля 1969', '9 мая 1965'], 1],
      ['Как назывался корабль первого полёта?', ['«Союз-1»', '«Мир»', '«Восток-1»', '«Буран»'], 2],
      ['Сколько длился первый полёт?', ['Около 45 минут', 'Около 108 минут', 'Около 3 часов', 'Сутки'], 1],
      ['В каком году запустили первый спутник?', ['1955', '1957', '1961', '1969'], 1]
    ]
  }
];


const DEMO_SCHOOL_CLASS = {
  id: 1, name: '7 «Б»', code: 'HIST7B', teacher_name: 'Ирина Петровна (демо)',
  members: ['Маша К.', 'Дима С.', 'Лиза П.', 'Артём В.']
};


const DEMO_TEACHER_STUDENTS = [
  { id: 501, name: 'Маша К.', cards_studied: 18, cards_known: 15, tests_taken: 4, avg_test_percent: 88, duels_played: 3, duels_won: 2, days_ago: 0 },
  { id: 502, name: 'Дима С.', cards_studied: 9, cards_known: 5, tests_taken: 2, avg_test_percent: 63, duels_played: 1, duels_won: 0, days_ago: 1 },
  { id: 503, name: 'Лиза П.', cards_studied: 22, cards_known: 20, tests_taken: 5, avg_test_percent: 95, duels_played: 4, duels_won: 4, days_ago: 0 },
  { id: 504, name: 'Артём В.', cards_studied: 0, cards_known: 0, tests_taken: 0, avg_test_percent: null, duels_played: 0, duels_won: 0, days_ago: null }
];


const demoCards = {};
const demoQuestions = {};
DEMO_DECKS.forEach(deck => {
  deck.cardList = deck.cards.map(([question, answer], i) => {
    const card = { id: deck.id * 100 + i + 1, deck_id: deck.id, question, answer };
    demoCards[card.id] = card;
    return card;
  });
  deck.questionList = deck.questions.map(([text, options, right], i) => {
    const id = 5000 + deck.id * 100 + i + 1;
    const q = {
      id, deck_id: deck.id, text,
      options: options.map((t, k) => ({ id: id * 10 + k + 1, text: t })),
      correct_option_id: id * 10 + right + 1
    };
    demoQuestions[id] = q;
    return q;
  });
});


function demoInitialState() {
  const user = window.WebApp && window.WebApp.initDataUnsafe && window.WebApp.initDataUnsafe.user;
  return {
    me: {
      id: 1,
      first_name: user && user.first_name ? user.first_name : 'Гость',
      last_name: user && user.last_name ? user.last_name : '',
      photo_url: null,
      role: null,
      teacher_verified: false
    },
    reviews: {},        
    attempts: [],       
    joinedClassIds: [],
    teacherClasses: [],
    removedStudents: {}, 
    duels: [],
    nextId: 1000
  };
}

let demo = demoLoad();

function demoLoad() {
  try {
    const saved = JSON.parse(localStorage.getItem(DEMO_STORAGE_KEY));
    if (saved && saved.me) return saved;
  } catch (e) {}
  return demoInitialState();
}
function demoSave() {
  try { localStorage.setItem(DEMO_STORAGE_KEY, JSON.stringify(demo)); } catch (e) {}
}


function demoError(status, code, message) {
  throw new ApiError(status, code, message);
}
function now() { return new Date().toISOString(); }
function newId() { demo.nextId += 1; return demo.nextId; }
function randomCode(length) {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  let s = '';
  for (let i = 0; i < length; i++) s += chars[Math.floor(Math.random() * chars.length)];
  return s;
}
function findDeck(id) {
  const deck = DEMO_DECKS.find(d => d.id === Number(id));
  if (!deck) demoError(404, 'deck_not_found', 'Колода не найдена');
  return deck;
}
function needRole(role) {
  if (demo.me.role !== role) demoError(403, 'forbidden', 'Нет доступа');
}
function percent(part, total) {
  return total ? Math.round(part / total * 100) : 0;
}

function deckProgress(deck) {
  let studied = 0, known = 0, weak = 0, last = null;
  deck.cardList.forEach(card => {
    const r = demo.reviews[card.id];
    if (!r) return;
    studied++;
    if (r.last_grade === 'good') known++; else weak++;
    if (!last || r.reviewed_at > last) last = r.reviewed_at;
  });
  return { studied, known, weak, last_studied_at: last };
}

function myStats() {
  const finished = demo.attempts.filter(a => a.status === 'finished');
  const decks = DEMO_DECKS.map(deck => {
    const p = deckProgress(deck);
    return { deck_id: deck.id, title: deck.title, studied: p.studied, known: p.known, total: deck.cardList.length };
  });
  const myDuels = demo.duels.filter(d => d.status === 'finished');
  return {
    cards_studied: decks.reduce((s, d) => s + d.studied, 0),
    cards_known: decks.reduce((s, d) => s + d.known, 0),
    cards_total: decks.reduce((s, d) => s + d.total, 0),
    tests_taken: finished.length,
    avg_test_percent: finished.length
      ? Math.round(finished.reduce((s, a) => s + a.percent, 0) / finished.length)
      : null,
    duels_played: myDuels.length,
    duels_won: myDuels.filter(d => d.winner_user_id === demo.me.id).length,
    decks,
    recent_tests: finished.slice(-5).reverse().map(attemptSummary)
  };
}

function attemptSummary(a) {
  const deck = findDeck(a.deck_id);
  return {
    id: a.id, deck: { id: deck.id, title: deck.title },
    correct: a.correct, total: a.question_ids.length, percent: a.percent, finished_at: a.finished_at
  };
}

function attemptResult(a) {
  const deck = findDeck(a.deck_id);
  return {
    attempt: {
      id: a.id, deck: { id: deck.id, title: deck.title }, status: a.status,
      correct: a.correct, total: a.question_ids.length, percent: a.percent,
      started_at: a.started_at, finished_at: a.finished_at
    },
    review: a.question_ids.map(qid => {
      const q = demoQuestions[qid];
      const chosen = a.answers[qid] || null;
      return {
        question_id: qid, text: q.text, options: q.options,
        chosen_option_id: chosen, correct_option_id: q.correct_option_id,
        correct: chosen === q.correct_option_id
      };
    })
  };
}

function schoolClassView() {
  const c = DEMO_SCHOOL_CLASS;
  return { id: c.id, name: c.name, teacher_name: c.teacher_name, member_count: c.members.length + 1 };
}

function teacherClassStudents(cls) {
  const removed = demo.removedStudents[cls.id] || [];
  return DEMO_TEACHER_STUDENTS.filter(s => !removed.includes(s.id)).map(s => ({
    id: s.id, name: s.name,
    cards_studied: s.cards_studied, cards_known: s.cards_known,
    tests_taken: s.tests_taken, avg_test_percent: s.avg_test_percent,
    duels_played: s.duels_played,
    last_active_at: s.days_ago === null ? null : new Date(Date.now() - s.days_ago * 86400000).toISOString()
  }));
}
function teacherClassView(cls) {
  const students = teacherClassStudents(cls);
  const tested = students.filter(s => s.avg_test_percent !== null);
  return {
    id: cls.id, name: cls.name, code: cls.code, created_at: cls.created_at,
    member_count: students.length,
    avg_test_percent: tested.length ? Math.round(tested.reduce((s, x) => s + x.avg_test_percent, 0) / tested.length) : null
  };
}
function findTeacherClass(id) {
  const cls = demo.teacherClasses.find(c => c.id === Number(id));
  if (!cls) demoError(404, 'class_not_found', 'Класс не найден');
  return cls;
}


const DEMO_OPPONENT_JOIN_MS = 3000;
const DEMO_OPPONENT_STEP_MS = 3500;

function updateDemoDuel(d) {
  const t = Date.now();
  if (d.status === 'waiting' && d.created_by_me && t - Date.parse(d.created_at) >= DEMO_OPPONENT_JOIN_MS) {
    d.opponent = { user_id: 900, name: 'Соперник (демо)' };
    d.status = 'active';
    d.started_at = now();
  }
  if (d.status === 'active') {
    const total = d.question_ids.length;
    d.opponent_answered = Math.min(total, Math.floor((t - Date.parse(d.started_at)) / DEMO_OPPONENT_STEP_MS));
    const myDone = Object.keys(d.my_answers).length === total;
    if (myDone && d.opponent_answered === total) finishDemoDuel(d);
  }
}

function finishDemoDuel(d, forfeit) {
  const total = d.question_ids.length;
  d.opponent_answered = forfeit ? d.opponent_answered : total;
  const my = duelMyScore(d);
  const opp = duelOpponentScore(d);
  d.status = 'finished';
  d.finished_at = now();
  if (forfeit) d.winner_user_id = d.opponent.user_id;
  else d.winner_user_id = my > opp ? demo.me.id : opp > my ? d.opponent.user_id : null;
}

function duelMyScore(d) {
  return Object.entries(d.my_answers).filter(([qid, opt]) => demoQuestions[qid].correct_option_id === opt).length;
}
function duelOpponentScore(d) {
  
  let score = 0;
  for (let i = 0; i < (d.opponent_answered || 0); i++) if (i % 3 !== 2) score++;
  return score;
}

function duelView(d) {
  updateDemoDuel(d);
  const deck = findDeck(d.deck_id);
  const total = d.question_ids.length;
  const myAnswered = Object.keys(d.my_answers).length;
  const players = [{
    user_id: demo.me.id, name: demo.me.first_name, is_me: true,
    score: duelMyScore(d), answered: myAnswered, finished: myAnswered === total
  }];
  if (d.opponent) {
    players.push({
      user_id: d.opponent.user_id, name: d.opponent.name, is_me: false,
      score: duelOpponentScore(d), answered: d.opponent_answered || 0,
      finished: (d.opponent_answered || 0) === total
    });
  }
  let result = null;
  if (d.status === 'finished') {
    result = d.winner_user_id === demo.me.id ? 'win' : d.winner_user_id === null ? 'draw' : 'lose';
  }
  const view = {
    id: d.id, code: d.code, status: d.status,
    deck: { id: deck.id, title: deck.title },
    question_count: total, created_at: d.created_at,
    players, winner_user_id: d.winner_user_id || null, result
  };
  if (d.status === 'active' || d.status === 'finished') {
    view.questions = d.question_ids.map(qid => {
      const q = demoQuestions[qid];
      return { id: q.id, text: q.text, options: q.options };
    });
    view.my_answers = Object.entries(d.my_answers).map(([qid, opt]) => ({
      question_id: Number(qid), option_id: opt, correct: demoQuestions[qid].correct_option_id === opt
    }));
  }
  return view;
}

function duelSummary(d) {
  const v = duelView(d);
  delete v.questions;
  delete v.my_answers;
  return v;
}

function findDuel(id) {
  const d = demo.duels.find(x => x.id === Number(id));
  if (!d) demoError(404, 'duel_not_found', 'Дуэль не найдена');
  return d;
}


function classmateDuel() {
  const deck = DEMO_DECKS[1];
  return {
    id: 777, code: 'MASHA7', status: 'waiting',
    deck: { id: deck.id, title: deck.title }, question_count: 4,
    created_at: now(),
    players: [{ user_id: 501, name: 'Маша К. (демо)', is_me: false, score: 0, answered: 0, finished: false }],
    winner_user_id: null, result: null
  };
}

function createDemoDuel(deckId, opponent) {
  const deck = findDeck(deckId);
  const d = {
    id: newId(), code: randomCode(6), status: opponent ? 'active' : 'waiting',
    deck_id: deck.id, question_ids: deck.questionList.map(q => q.id),
    created_by_me: !opponent, created_at: now(),
    started_at: opponent ? now() : null,
    opponent: opponent || null, opponent_answered: 0,
    my_answers: {}, winner_user_id: null
  };
  demo.duels.push(d);
  return d;
}


const DEMO_ROUTES = [
  ['GET', /^\/api\/me$/, () => meView()],

  ['PUT', /^\/api\/me\/role$/, (m, body) => {
    const role = body && body.role;
    if (role !== 'student' && role !== 'teacher') demoError(400, 'invalid_role', 'Неизвестная роль');
    if (role === 'teacher' && !demo.me.teacher_verified) {
      const code = String(body.invite_code || '').trim().toUpperCase();
      if (code !== DEMO_INVITE_CODE) demoError(403, 'invalid_invite_code', 'Неверный код учителя');
      demo.me.teacher_verified = true;
    }
    demo.me.role = role;
    if (role === 'teacher' && demo.teacherClasses.length === 0) {
      demo.teacherClasses.push({ id: newId(), name: '8 «А» (демо)', code: randomCode(6), created_at: now() });
    }
    return meView();
  }],

  ['GET', /^\/api\/me\/stats$/, () => myStats()],

  ['DELETE', /^\/api\/me$/, () => { demo = demoInitialState(); return null; }],

  ['GET', /^\/api\/decks$/, () => ({
    decks: DEMO_DECKS.map(d => ({
      id: d.id, title: d.title, topic: d.topic, period: d.period,
      card_count: d.cardList.length, question_count: d.questionList.length,
      progress: deckProgress(d)
    }))
  })],

  ['GET', /^\/api\/decks\/(\d+)\/cards$/, (m, body, query) => {
    const deck = findDeck(m[1]);
    const mode = query.get('mode') || 'all';
    let cards = deck.cardList;
    if (mode === 'new') cards = cards.filter(c => !demo.reviews[c.id]);
    if (mode === 'weak') cards = cards.filter(c => demo.reviews[c.id] && demo.reviews[c.id].last_grade !== 'good');
    return {
      deck: { id: deck.id, title: deck.title },
      cards: cards.map(c => ({ id: c.id, question: c.question, answer: c.answer, level: demo.reviews[c.id] ? demo.reviews[c.id].level : 0 }))
    };
  }],

  ['POST', /^\/api\/study\/reviews$/, (m, body) => {
    const card = demoCards[body && body.card_id];
    if (!card) demoError(404, 'card_not_found', 'Карточка не найдена');
    if (!['again', 'hard', 'good'].includes(body.grade)) demoError(400, 'invalid_grade', 'Неверная оценка');
    const old = demo.reviews[card.id] || { level: 0 };
    const level = body.grade === 'again' ? 0 : body.grade === 'hard' ? Math.max(1, old.level) : Math.min(5, old.level + 1);
    const days = [0, 1, 2, 4, 7, 14][level];
    demo.reviews[card.id] = { level, last_grade: body.grade, reviewed_at: now() };
    return { card_id: card.id, level, next_review_at: new Date(Date.now() + days * 86400000).toISOString() };
  }],

  ['POST', /^\/api\/tests$/, (m, body) => {
    const deck = findDeck(body && body.deck_id);
    const a = {
      id: newId(), deck_id: deck.id, question_ids: deck.questionList.map(q => q.id),
      answers: {}, status: 'in_progress', correct: 0, percent: 0,
      started_at: now(), finished_at: null
    };
    demo.attempts.push(a);
    return {
      attempt: { id: a.id, deck: { id: deck.id, title: deck.title }, status: a.status, started_at: a.started_at },
      questions: a.question_ids.map(qid => ({ id: qid, text: demoQuestions[qid].text, options: demoQuestions[qid].options }))
    };
  }],

  ['POST', /^\/api\/tests\/(\d+)\/answers$/, (m, body) => {
    const a = demo.attempts.find(x => x.id === Number(m[1]));
    if (!a) demoError(404, 'attempt_not_found', 'Попытка не найдена');
    if (a.status !== 'in_progress') demoError(409, 'attempt_finished', 'Тест уже завершён');
    const q = demoQuestions[body && body.question_id];
    if (!q || !a.question_ids.includes(q.id)) demoError(400, 'invalid_question', 'Такого вопроса нет в тесте');
    if (a.answers[q.id]) demoError(409, 'already_answered', 'На этот вопрос уже ответили');
    if (!q.options.some(o => o.id === body.option_id)) demoError(400, 'invalid_option', 'Такого варианта нет');
    a.answers[q.id] = body.option_id;
    const answered = Object.entries(a.answers);
    const correctCount = answered.filter(([qid, opt]) => demoQuestions[qid].correct_option_id === opt).length;
    return {
      correct: body.option_id === q.correct_option_id,
      correct_option_id: q.correct_option_id,
      correct_count: correctCount,
      wrong_count: answered.length - correctCount
    };
  }],

  ['POST', /^\/api\/tests\/(\d+)\/finish$/, m => {
    const a = demo.attempts.find(x => x.id === Number(m[1]));
    if (!a) demoError(404, 'attempt_not_found', 'Попытка не найдена');
    if (a.status === 'in_progress') {
      a.correct = Object.entries(a.answers).filter(([qid, opt]) => demoQuestions[qid].correct_option_id === opt).length;
      a.percent = percent(a.correct, a.question_ids.length);
      a.status = 'finished';
      a.finished_at = now();
    }
    return attemptResult(a);
  }],

  ['GET', /^\/api\/tests\/(\d+)$/, m => {
    const a = demo.attempts.find(x => x.id === Number(m[1]));
    if (!a || a.status !== 'finished') demoError(404, 'attempt_not_found', 'Попытка не найдена');
    return attemptResult(a);
  }],

  ['GET', /^\/api\/tests$/, (m, body, query) => {
    const deckId = query.get('deck_id');
    const list = demo.attempts.filter(a => a.status === 'finished' && (!deckId || a.deck_id === Number(deckId)));
    return { attempts: list.slice().reverse().map(attemptSummary) };
  }],

  ['GET', /^\/api\/classes$/, () => {
    needRole('student');
    return { classes: demo.joinedClassIds.includes(DEMO_SCHOOL_CLASS.id) ? [schoolClassView()] : [] };
  }],

  ['POST', /^\/api\/classes\/join$/, (m, body) => {
    needRole('student');
    const code = String(body && body.code || '').trim().toUpperCase();
    if (!/^[A-Z0-9]{4,8}$/.test(code)) demoError(400, 'invalid_code', 'Код должен быть из 4–8 латинских букв или цифр');
    if (code !== DEMO_SCHOOL_CLASS.code) demoError(404, 'class_not_found', 'Класс с таким кодом не найден');
    if (demo.joinedClassIds.includes(DEMO_SCHOOL_CLASS.id)) demoError(409, 'already_member', 'Ты уже в этом классе');
    demo.joinedClassIds.push(DEMO_SCHOOL_CLASS.id);
    return { class: schoolClassView() };
  }],

  ['GET', /^\/api\/classes\/(\d+)$/, m => {
    needRole('student');
    if (Number(m[1]) !== DEMO_SCHOOL_CLASS.id || !demo.joinedClassIds.includes(DEMO_SCHOOL_CLASS.id)) {
      demoError(404, 'class_not_found', 'Класс не найден');
    }
    const s = myStats();
    return {
      class: schoolClassView(),
      members: [
        { id: demo.me.id, name: demo.me.first_name, is_me: true },
        ...DEMO_SCHOOL_CLASS.members.map((name, i) => ({ id: 501 + i, name: name + ' (демо)', is_me: false }))
      ],
      my_progress: {
        cards_studied: s.cards_studied, cards_known: s.cards_known,
        tests_taken: s.tests_taken, avg_test_percent: s.avg_test_percent
      }
    };
  }],

  ['GET', /^\/api\/teacher\/classes$/, () => {
    needRole('teacher');
    return { classes: demo.teacherClasses.map(teacherClassView) };
  }],

  ['POST', /^\/api\/teacher\/classes$/, (m, body) => {
    needRole('teacher');
    const name = String(body && body.name || '').trim();
    if (name.length < 1 || name.length > 40) demoError(400, 'invalid_name', 'Название должно быть от 1 до 40 символов');
    const cls = { id: newId(), name, code: randomCode(6), created_at: now() };
    demo.teacherClasses.push(cls);
    demo.removedStudents[cls.id] = DEMO_TEACHER_STUDENTS.map(s => s.id); 
    return { class: teacherClassView(cls) };
  }],

  ['GET', /^\/api\/teacher\/classes\/(\d+)$/, m => {
    needRole('teacher');
    const cls = findTeacherClass(m[1]);
    return { class: teacherClassView(cls), students: teacherClassStudents(cls) };
  }],

  ['GET', /^\/api\/teacher\/classes\/(\d+)\/stats$/, m => {
    needRole('teacher');
    const cls = findTeacherClass(m[1]);
    const students = teacherClassStudents(cls);
    const tested = students.filter(s => s.avg_test_percent !== null);
    const empty = students.length === 0;
    return {
      students_count: students.length,
      active_students_7d: students.filter(s => s.last_active_at).length,
      avg_test_percent: tested.length ? Math.round(tested.reduce((s, x) => s + x.avg_test_percent, 0) / tested.length) : null,
      cards_studied_total: students.reduce((s, x) => s + x.cards_studied, 0),
      decks: empty ? [] : DEMO_DECKS.slice(0, 3).map((d, i) => ({
        deck_id: d.id, title: d.title, students_started: Math.max(0, students.length - 1 - i), avg_known_percent: [82, 64, 45][i]
      })),
      hardest_cards: empty ? [] : [
        { card_id: 304, question: DEMO_DECKS[2].cardList[3].question, deck_title: DEMO_DECKS[2].title, wrong_rate: 58, reviews: 12 },
        { card_id: 205, question: DEMO_DECKS[1].cardList[4].question, deck_title: DEMO_DECKS[1].title, wrong_rate: 41, reviews: 17 },
        { card_id: 104, question: DEMO_DECKS[0].cardList[3].question, deck_title: DEMO_DECKS[0].title, wrong_rate: 33, reviews: 9 }
      ]
    };
  }],

  ['GET', /^\/api\/teacher\/classes\/(\d+)\/students\/(\d+)$/, m => {
    needRole('teacher');
    const cls = findTeacherClass(m[1]);
    const s = teacherClassStudents(cls).find(x => x.id === Number(m[2]));
    if (!s) demoError(404, 'student_not_found', 'Ученик не найден в этом классе');
    const full = DEMO_TEACHER_STUDENTS.find(x => x.id === s.id);
    const share = s.cards_studied ? s.cards_known / s.cards_studied : 0;
    return {
      student: { id: s.id, name: s.name, last_active_at: s.last_active_at },
      stats: {
        cards_studied: s.cards_studied, cards_known: s.cards_known, tests_taken: s.tests_taken,
        avg_test_percent: s.avg_test_percent, duels_played: s.duels_played, duels_won: full.duels_won
      },
      decks: s.cards_studied ? DEMO_DECKS.slice(0, 3).map(d => ({
        deck_id: d.id, title: d.title, total: d.cardList.length,
        studied: Math.min(d.cardList.length, Math.ceil(s.cards_studied / 4)),
        known: Math.round(Math.min(d.cardList.length, Math.ceil(s.cards_studied / 4)) * share)
      })) : [],
      recent_tests: s.tests_taken ? [
        { id: 9001, deck: { id: 1, title: DEMO_DECKS[0].title }, correct: 3, total: 4, percent: 75, finished_at: now() }
      ] : [],
      weak_cards: s.cards_studied && share < 0.9 ? [
        { card_id: 304, question: DEMO_DECKS[2].cardList[3].question, deck_title: DEMO_DECKS[2].title }
      ] : []
    };
  }],

  ['DELETE', /^\/api\/teacher\/classes\/(\d+)\/students\/(\d+)$/, m => {
    needRole('teacher');
    const cls = findTeacherClass(m[1]);
    const list = demo.removedStudents[cls.id] || (demo.removedStudents[cls.id] = []);
    if (!list.includes(Number(m[2]))) list.push(Number(m[2]));
    return null;
  }],

  ['GET', /^\/api\/duels$/, () => {
    needRole('student');
    const mine = demo.duels.filter(d => { updateDemoDuel(d); return d.status === 'waiting' || d.status === 'active'; });
    const history = demo.duels.filter(d => d.status === 'finished' || d.status === 'cancelled');
    const classmateTaken = demo.duels.some(d => d.code === 'MASHA7');
    return {
      open: demo.joinedClassIds.length && !classmateTaken ? [classmateDuel()] : [],
      mine: mine.map(duelSummary),
      history: history.slice(-10).reverse().map(duelSummary)
    };
  }],

  ['POST', /^\/api\/duels$/, (m, body) => {
    needRole('student');
    return { duel: duelView(createDemoDuel(body && body.deck_id)) };
  }],

  ['POST', /^\/api\/duels\/join$/, (m, body) => {
    needRole('student');
    const code = String(body && body.code || '').trim().toUpperCase();
    if (code === 'MASHA7' && demo.joinedClassIds.length && !demo.duels.some(d => d.code === code)) {
      const d = createDemoDuel(2, { user_id: 501, name: 'Маша К. (демо)' });
      d.code = code;
      return { duel: duelView(d) };
    }
    const own = demo.duels.find(d => d.code === code);
    if (own && own.created_by_me) demoError(409, 'own_duel', 'Это твоя дуэль. Отправь код сопернику');
    demoError(404, 'duel_not_found', 'Дуэль с таким кодом не найдена');
  }],

  ['GET', /^\/api\/duels\/(\d+)$/, m => ({ duel: duelView(findDuel(m[1])) })],

  ['POST', /^\/api\/duels\/(\d+)\/answers$/, (m, body) => {
    const d = findDuel(m[1]);
    updateDemoDuel(d);
    if (d.status !== 'active') demoError(409, 'duel_not_active', 'Дуэль сейчас не идёт');
    const q = demoQuestions[body && body.question_id];
    if (!q || !d.question_ids.includes(q.id)) demoError(400, 'invalid_question', 'Такого вопроса нет в дуэли');
    if (d.my_answers[q.id]) demoError(409, 'already_answered', 'На этот вопрос уже ответили');
    d.my_answers[q.id] = body.option_id;
    return { correct: body.option_id === q.correct_option_id, correct_option_id: q.correct_option_id, duel: duelView(d) };
  }],

  ['POST', /^\/api\/duels\/(\d+)\/leave$/, m => {
    const d = findDuel(m[1]);
    updateDemoDuel(d);
    if (d.status === 'waiting') d.status = 'cancelled';
    else if (d.status === 'active') finishDemoDuel(d, true);
    return { duel: duelView(d) };
  }]
];

function meView() {
  return {
    user: { id: demo.me.id, first_name: demo.me.first_name, last_name: demo.me.last_name, photo_url: demo.me.photo_url },
    role: demo.me.role,
    teacher_verified: demo.me.teacher_verified
  };
}


async function demoRequest(method, path, body) {
  await new Promise(r => setTimeout(r, 250)); 
  const url = new URL(path, 'http://demo');
  for (const [m, pattern, handler] of DEMO_ROUTES) {
    if (m !== method) continue;
    const match = url.pathname.match(pattern);
    if (!match) continue;
    const result = handler(match, body, url.searchParams);
    demoSave();

    return result === null ? null : JSON.parse(JSON.stringify(result));
  }
  demoError(404, 'not_found', 'Такого запроса нет: ' + method + ' ' + url.pathname);
}
