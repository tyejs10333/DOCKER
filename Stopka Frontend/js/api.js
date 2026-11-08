

class ApiError extends Error {
  constructor(status, code, message) {
    super(message);
    this.status = status; 
    this.code = code;     
  }
}


function maxInitData() {
  return (window.WebApp && window.WebApp.initData) || '';
}

const DEFAULT_MESSAGES = {
  400: 'Что-то не так с запросом',
  401: 'Не удалось подтвердить вход через MAX',
  403: 'Нет доступа',
  404: 'Не найдено',
  409: 'Действие сейчас недоступно',
  500: 'Ошибка на сервере. Попробуй позже'
};

async function request(method, path, body) {
  if (DEMO_MODE) return demoRequest(method, path, body);

  const headers = { 'X-Max-Init-Data': maxInitData() };
  if (DEV_WITHOUT_MAX) headers['X-Dev-User-Id'] = DEV_USER_ID;
  if (body !== undefined) headers['Content-Type'] = 'application/json';

  let res;
  try {
    res = await fetch(CONFIG.apiUrl + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined
    });
  } catch (e) {
    throw new ApiError(0, 'network', 'Нет связи с сервером. Проверь интернет');
  }

  let data = null;
  if (res.status !== 204) {
    try { data = await res.json(); } catch (e) { data = null; }
  }

  if (!res.ok) {
    
    const err = (data && data.error) || {};
    throw new ApiError(
      res.status,
      err.code || 'http_' + res.status,
      err.message || DEFAULT_MESSAGES[res.status] || DEFAULT_MESSAGES[500]
    );
  }
  return data;
}

const id = value => encodeURIComponent(value);


const API = {
  // Профиль и роль
  me: () => request('GET', '/api/me'),
  setRole: (role, inviteCode) => request('PUT', '/api/me/role', inviteCode ? { role, invite_code: inviteCode } : { role }),
  myStats: () => request('GET', '/api/me/stats'),
  deleteMe: () => request('DELETE', '/api/me'),

  // Колоды и карточки
  decks: () => request('GET', '/api/decks'),
  deckCards: (deckId, mode) => request('GET', `/api/decks/${id(deckId)}/cards?mode=${id(mode)}`),
  reviewCard: (cardId, grade) => request('POST', '/api/study/reviews', { card_id: cardId, grade }),

  // Тесты
  startTest: deckId => request('POST', '/api/tests', { deck_id: deckId }),
  answerTest: (attemptId, questionId, optionId) =>
    request('POST', `/api/tests/${id(attemptId)}/answers`, { question_id: questionId, option_id: optionId }),
  finishTest: attemptId => request('POST', `/api/tests/${id(attemptId)}/finish`),
  testResult: attemptId => request('GET', `/api/tests/${id(attemptId)}`),
  testHistory: deckId => request('GET', '/api/tests' + (deckId ? `?deck_id=${id(deckId)}` : '')),

  // Классы ученика
  myClasses: () => request('GET', '/api/classes'),
  joinClass: code => request('POST', '/api/classes/join', { code }),
  classInfo: classId => request('GET', `/api/classes/${id(classId)}`),

  // Учитель
  teacherClasses: () => request('GET', '/api/teacher/classes'),
  createClass: name => request('POST', '/api/teacher/classes', { name }),
  teacherClass: classId => request('GET', `/api/teacher/classes/${id(classId)}`),
  classStats: classId => request('GET', `/api/teacher/classes/${id(classId)}/stats`),
  studentProgress: (classId, studentId) =>
    request('GET', `/api/teacher/classes/${id(classId)}/students/${id(studentId)}`),
  removeStudent: (classId, studentId) =>
    request('DELETE', `/api/teacher/classes/${id(classId)}/students/${id(studentId)}`),

  // Дуэли
  duels: () => request('GET', '/api/duels'),
  createDuel: deckId => request('POST', '/api/duels', { deck_id: deckId }),
  joinDuel: code => request('POST', '/api/duels/join', { code }),
  duel: duelId => request('GET', `/api/duels/${id(duelId)}`),
  answerDuel: (duelId, questionId, optionId) =>
    request('POST', `/api/duels/${id(duelId)}/answers`, { question_id: questionId, option_id: optionId }),
  leaveDuel: duelId => request('POST', `/api/duels/${id(duelId)}/leave`)
};
