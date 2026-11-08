
const DUEL_RESULT_TEXT = { win: 'Победа!', lose: 'Поражение', draw: 'Ничья' };

defineScreen('duels', 'duels', view => {
  const body = h('div', { class: 'stack' });
  view.append(header('Дуэли', 'Соревнуйся с одноклассниками'), body);

  load(body, API.duels, ({ open, mine, history }) => {
    body.append(h('div', { class: 'actions actions--two' },
      h('button', { class: 'btn btn--main', onclick: () => go('decks', { pick: 'duel' }) }, 'Создать дуэль'),
      h('button', { class: 'btn btn--light', onclick: () => go('duelJoin') }, 'Войти по коду')));

    if (mine.length) {
      body.append(sectionTitle('Мои дуэли'),
        h('div', { class: 'deck-list' }, mine.map(d => duelRow(d, () => go('duelRoom', { duelId: d.id })))));
    }

    body.append(sectionTitle('Ждут соперника'), open.length
      ? h('div', { class: 'deck-list' }, open.map(d => {
          const author = d.players[0] ? d.players[0].name : 'Одноклассник';
          return h('div', { class: 'deck' },
            h('span', { class: 'deck__badge' }, icon('duel')),
            h('span', { class: 'deck__body' },
              h('span', { class: 'deck__name' }, d.deck.title),
              h('span', { class: 'muted small' }, author + ' · ' + countText(d.question_count, 'вопрос', 'вопроса', 'вопросов'))),
            actionButton('Принять', 'btn btn--main btn--small', async () => {
              const res = await API.joinDuel(d.code);
              go('duelRoom', { duelId: res.duel.id });
            }));
        }))
      : stateBox('Сейчас никто из одноклассников не ждёт соперника. Создай дуэль и отправь код другу.'));

    if (history.length) {
      body.append(sectionTitle('История'),
        h('div', { class: 'deck-list' }, history.map(d => duelRow(d, () => go('duelRoom', { duelId: d.id })))));
    }
  });
});

function duelRow(d, onClick) {
  const meP = d.players.find(p => p.is_me);
  const opp = d.players.find(p => !p.is_me);
  const status = {
    waiting: 'ждём соперника · код ' + d.code,
    active: 'идёт · ' + (meP ? meP.score : 0) + ':' + (opp ? opp.score : 0),
    finished: (DUEL_RESULT_TEXT[d.result] || '') + ' · ' + (meP ? meP.score : 0) + ':' + (opp ? opp.score : 0),
    cancelled: 'отменена'
  }[d.status];
  return h('button', { class: 'deck', onclick: onClick },
    h('span', { class: 'deck__badge' + (d.result === 'win' ? ' deck__badge--good' : '') }, icon('duel')),
    h('span', { class: 'deck__body' },
      h('span', { class: 'deck__name' }, d.deck.title + (opp ? ' — ' + opp.name : '')),
      h('span', { class: 'muted small' }, status)),
    icon('next'));
}

defineScreen('duelCreate', 'duels', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, API.decks, ({ decks }) => {
    const deck = decks.find(d => d.id === params.deckId);
    if (!deck) return body.append(header('Новая дуэль'), stateBox('Колода не найдена'));
    body.append(
      header('Новая дуэль', deck.title),
      h('div', { class: 'panel' },
        h('b', null, 'Как это работает'),
        h('p', { class: 'muted' }, 'Ты получишь код. Отправь его однокласснику: когда он войдёт, вы ответите на одинаковые ' +
          countText(deck.question_count, 'вопрос', 'вопроса', 'вопросов') + '. Побеждает тот, у кого больше верных ответов.')),
      actionButton('Создать и получить код', 'btn btn--main', async () => {
        const { duel } = await API.createDuel(deck.id);
        go('duelRoom', { duelId: duel.id }, { replace: true });
      }));
  });
});

defineScreen('duelJoin', 'duels', (view, params) => {
  const input = h('input', { class: 'input input--code', id: 'duelCode', maxlength: 8, autocomplete: 'off', placeholder: 'Код', value: params.code || '' });
  view.append(
    header('Войти в дуэль'),
    h('form', { class: 'panel', onsubmit: e => e.preventDefault() },
      h('label', { for: 'duelCode' }, h('b', null, 'Код дуэли')),
      h('p', { class: 'muted' }, 'Попроси код у того, кто создал дуэль'),
      input,
      actionButton('Войти', 'btn btn--main', async () => {
        const code = input.value.trim().toUpperCase();
        if (!/^[A-Z0-9]{4,8}$/.test(code)) throw new ApiError(400, 'invalid_code', 'Код — это 4–8 латинских букв или цифр');
        const { duel } = await API.joinDuel(code);
        go('duelRoom', { duelId: duel.id }, { replace: true });
      })));
});


defineScreen('duelRoom', 'duels', (view, params) => {
  let duel = null;
  let shownStatus = null;
  let timer = null;
  let stopped = false;
  let busy = false; 

  const content = h('div', { class: 'stack' });
  const connection = h('p', { class: 'hint', hidden: true }, 'Нет связи с сервером, пробуем снова…');
  view.append(content, connection);
  content.append(loadingState());

  async function refresh() {
    try {
      const res = await API.duel(params.duelId);
      if (stopped) return;
      connection.hidden = true;
      duel = res.duel;
      draw();
    } catch (err) {
      if (stopped) return;
      if (!duel) {
        content.replaceChildren(header('Дуэль'), stateBox(errorText(err), 'Попробовать снова', refresh));
        return;
      }
      connection.hidden = false;
    }
    schedule();
  }

  function schedule() {
    clearTimeout(timer);
    if (!stopped && duel && (duel.status === 'waiting' || duel.status === 'active')) {
      timer = setTimeout(refresh, CONFIG.duelPollMs);
    }
  }

  function draw() {

    if (duel.status !== shownStatus) {
      shownStatus = duel.status;
      if (duel.status === 'waiting') drawWaiting();
      else if (duel.status === 'active') drawGame();
      else if (duel.status === 'finished') drawFinished();
      else drawCancelled();
    } else if (duel.status === 'active') {
      updateScoreboard();
      if (!busy && myNextQuestion() === null) drawWaitingForOpponent();
    }
  }


  function drawWaiting() {
    const link = CONFIG.botName ? 'https://max.ru/' + CONFIG.botName + '?startapp=duel_' + duel.code : null;
    content.replaceChildren(
      header('Дуэль', duel.deck.title),
      h('div', { class: 'panel waiting' },
        h('span', { class: 'spinner', 'aria-hidden': 'true' }),
        h('b', null, 'Ждём соперника'),
        codeBox(duel.code, 'Отправь этот код однокласснику. Он откроет «Дуэли» → «Войти по коду»'),
        link ? h('button', { class: 'btn btn--light btn--small', onclick: () => copyText(link) }, 'Скопировать ссылку-приглашение') : null,
        DEMO_MODE ? h('p', { class: 'hint' }, 'В демо-режиме соперник подключится сам через пару секунд') : null),
      actionButton('Отменить дуэль', 'btn btn--light', async () => {
        const res = await API.leaveDuel(duel.id);
        duel = res.duel;
        draw();
      }));
  }


  const scoreboard = h('div', { class: 'scoreboard' });
  const questionArea = h('div', { class: 'stack' });

  function drawGame() {
    content.replaceChildren(
      header('Дуэль', duel.deck.title),
      scoreboard,
      questionArea,
      confirmButton('Сдаться', 'Точно сдаться? Нажми ещё раз', async () => {
        const res = await API.leaveDuel(duel.id);
        duel = res.duel;
        draw();
      }));
    updateScoreboard();
    drawQuestion();
  }

  function updateScoreboard() {
    const total = duel.question_count;
    scoreboard.replaceChildren(...duel.players.map(p =>
      h('div', { class: 'player' + (p.is_me ? ' is-me' : '') },
        h('span', { class: 'player__name' }, p.is_me ? 'Ты' : p.name),
        h('b', { class: 'player__score' }, p.score),
        progressBar(p.answered, total),
        h('span', { class: 'muted small' }, p.answered + ' из ' + total))));
  }

  function myNextQuestion() {
    const answered = new Set(duel.my_answers.map(a => a.question_id));
    const index = duel.questions.findIndex(q => !answered.has(q.id));
    return index === -1 ? null : index;
  }

  function drawQuestion() {
    const index = myNextQuestion();
    if (index === null) return drawWaitingForOpponent();
    const q = duel.questions[index];
    const error = h('p', { class: 'error', hidden: true });
    const options = h('div', { class: 'options' }, q.options.map(o =>
      h('button', { class: 'option', 'data-id': o.id, onclick: () => answer(q, o.id, options, error) }, o.text)));
    questionArea.replaceChildren(
      h('span', { class: 'muted' }, 'Вопрос ' + (index + 1) + ' из ' + duel.question_count),
      h('p', { class: 'question' }, q.text),
      options, error);
  }

  async function answer(q, optionId, options, error) {
    if (busy) return;
    busy = true;
    const buttons = [...options.children];
    buttons.forEach(b => { b.disabled = true; });
    try {
      const res = await API.answerDuel(duel.id, q.id, optionId);
      if (stopped) return;
      buttons.forEach(b => {
        const bid = Number(b.dataset.id);
        if (bid === res.correct_option_id) b.classList.add('is-right');
        else if (bid === optionId) b.classList.add('is-wrong');
      });
      duel = res.duel;
      updateScoreboard();

      setTimeout(() => {
        busy = false;
        if (stopped) return;
        if (duel.status !== 'active') draw();
        else drawQuestion();
      }, 900);
    } catch (err) {
      busy = false;
      error.textContent = errorText(err);
      error.hidden = false;
      buttons.forEach(b => { b.disabled = false; });
    }
  }

  function drawWaitingForOpponent() {
    if (questionArea.dataset.done) return;
    questionArea.dataset.done = '1';
    questionArea.replaceChildren(h('div', { class: 'panel waiting' },
      h('span', { class: 'spinner', 'aria-hidden': 'true' }),
      h('b', null, 'Ты ответил на все вопросы'),
      h('p', { class: 'muted' }, 'Ждём, пока соперник закончит')));
  }

  function drawFinished() {
    const meP = duel.players.find(p => p.is_me);
    const opp = duel.players.find(p => !p.is_me);
    content.replaceChildren(
      header('Дуэль', duel.deck.title),
      h('div', { class: 'result result--' + duel.result },
        h('h1', { class: 'title' }, DUEL_RESULT_TEXT[duel.result]),
        h('p', { class: 'duel-score' }, (meP ? meP.score : 0) + ' : ' + (opp ? opp.score : 0)),
        h('p', { class: 'muted' }, 'Ты' + ' против ' + (opp ? opp.name : 'соперника'))),
      duel.my_answers && duel.my_answers.length
        ? h('div', { class: 'deck-list' }, duel.questions.map((q, i) => {
            const a = duel.my_answers.find(x => x.question_id === q.id);
            const chosen = a && q.options.find(o => o.id === a.option_id);
            return h('div', { class: 'review ' + (a && a.correct ? 'review--right' : 'review--wrong') },
              h('b', null, (i + 1) + '. ' + q.text),
              h('span', { class: a && a.correct ? 'good' : 'bad' }, 'Твой ответ: ' + (chosen ? chosen.text : 'нет ответа')));
          }))
        : null,
      h('div', { class: 'actions' },
        h('button', { class: 'btn btn--main', onclick: () => go('duelCreate', { deckId: duel.deck.id }, { replace: true }) }, 'Реванш'),
        h('button', { class: 'btn btn--light', onclick: () => go('duels', {}, { root: true }) }, 'К дуэлям')));
  }

  function drawCancelled() {
    content.replaceChildren(
      header('Дуэль', duel.deck.title),
      stateBox('Дуэль отменена', 'К дуэлям', () => go('duels', {}, { root: true })));
  }

  refresh();


  return () => {
    stopped = true;
    clearTimeout(timer);
  };
});
