
function deckRow(deck, onClick) {
  const p = deck.progress || { studied: 0, known: 0, weak: 0 };
  return h('button', { class: 'deck', onclick: onClick },
    h('span', { class: 'deck__badge' }, deck.period ? deck.period.replace(' век', '') : ''),
    h('span', { class: 'deck__body' },
      h('span', { class: 'deck__name' }, deck.title),
      h('span', { class: 'muted small' }, deck.topic),
      progressBar(p.known, deck.card_count)),
    h('span', { class: 'muted small' }, p.known + '/' + deck.card_count));
}

defineScreen('home', 'home', view => {
  const body = h('div', { class: 'stack' });
  view.append(header('Привет, ' + userName() + '!', 'Что повторим сегодня?'), body);

  load(body, () => Promise.all([API.decks(), API.myClasses()]), ([deckData, classData]) => {
    const decks = deckData.decks;
    const known = decks.reduce((s, d) => s + d.progress.known, 0);
    const total = decks.reduce((s, d) => s + d.card_count, 0);


    body.append(h('div', { class: 'panel' },
      h('div', { class: 'row' }, h('b', null, 'Твой прогресс'), h('span', { class: 'muted' }, known + ' из ' + countText(total, 'карточки', 'карточек', 'карточек'))),
      progressBar(known, total)));

    
    const started = decks
      .filter(d => d.progress.last_studied_at && d.progress.studied < d.card_count)
      .sort((a, b) => b.progress.last_studied_at.localeCompare(a.progress.last_studied_at))[0];
    if (started) {
      body.append(h('button', { class: 'hero', onclick: () => go('study', { deckId: started.id, mode: 'new' }) },
        h('span', { class: 'label' }, 'Продолжить'),
        h('span', { class: 'hero__title' }, started.title),
        progressBar(started.progress.studied, started.card_count),
        h('span', { class: 'muted' }, 'Изучено ' + started.progress.studied + ' из ' + started.card_count)));
    } else if (decks.length && !decks.some(d => d.progress.studied)) {
      body.append(h('button', { class: 'hero', onclick: () => go('deck', { deckId: decks[0].id }) },
        h('span', { class: 'label' }, 'С чего начать'),
        h('span', { class: 'hero__title' }, decks[0].title),
        h('span', { class: 'muted' }, countText(decks[0].card_count, 'карточка', 'карточки', 'карточек'))));
    }

   
    body.append(h('div', { class: 'quick' },
      quickButton('decks', 'Карточки', () => go('decks')),
      quickButton('test', 'Тест', () => go('decks', { pick: 'test' })),
      quickButton('duel', 'Дуэль', () => go('duels', {}, { root: true }))));

    
    const weak = decks.filter(d => d.progress.weak > 0);
    if (weak.length) {
      body.append(sectionTitle('Пора повторить'),
        h('div', { class: 'deck-list' }, weak.map(d =>
          h('button', { class: 'banner', onclick: () => go('study', { deckId: d.id, mode: 'weak' }) },
            h('span', { class: 'banner__icon banner__icon--warn' }, icon('repeat')),
            h('span', null, h('b', null, d.title),
              h('span', { class: 'muted' }, countText(d.progress.weak, 'слабая карточка', 'слабые карточки', 'слабых карточек')))))));
    }

    
    const cls = classData.classes[0];
    body.append(cls
      ? h('button', { class: 'banner', onclick: () => go('classInfo', { classId: cls.id }) },
          h('span', { class: 'banner__icon' }, icon('class')),
          h('span', null, h('b', null, 'Класс ' + cls.name), h('span', { class: 'muted' }, 'Учитель: ' + cls.teacher_name)))
      : h('button', { class: 'banner', onclick: () => go('classJoin') },
          h('span', { class: 'banner__icon' }, icon('class')),
          h('span', null, h('b', null, 'Вступи в класс'), h('span', { class: 'muted' }, 'Введи код от учителя'))));
  });
});

function quickButton(iconName, label, onClick) {
  return h('button', { class: 'quick__item', onclick: onClick }, h('span', { class: 'quick__icon' }, icon(iconName)), label);
}


defineScreen('decks', 'decks', (view, params) => {
  const subtitles = { test: 'Выбери колоду для теста', duel: 'Выбери колоду для дуэли' };
  const body = h('div', { class: 'stack' });
  view.append(header('Колоды', subtitles[params.pick] || 'Темы истории России'), body);

  load(body, API.decks, ({ decks }) => {
    let topic = 'Все';
    const topics = ['Все', ...new Set(decks.map(d => d.topic))];
    const input = h('input', { type: 'search', placeholder: 'Найти колоду', 'aria-label': 'Найти колоду', oninput: draw });
    const chips = h('div', { class: 'chips', role: 'group', 'aria-label': 'Темы' });
    const list = h('div', { class: 'deck-list' });

    function openDeck(deck) {
      if (params.pick === 'test') go('test', { deckId: deck.id });
      else if (params.pick === 'duel') go('duelCreate', { deckId: deck.id });
      else go('deck', { deckId: deck.id });
    }

    function draw() {
      chips.replaceChildren(...topics.map(t => h('button', {
        class: 'chip' + (t === topic ? ' is-active' : ''), 'aria-pressed': String(t === topic),
        onclick: () => { topic = t; draw(); }
      }, t)));
      const text = input.value.trim().toLowerCase();
      const found = decks.filter(d => (topic === 'Все' || d.topic === topic) && d.title.toLowerCase().includes(text));
      list.replaceChildren(...(found.length
        ? found.map(d => deckRow(d, () => openDeck(d)))
        : [stateBox(decks.length ? 'Ничего не нашлось' : 'Колод пока нет')]));
    }

    body.append(h('label', { class: 'search' }, icon('search'), input), chips, list);
    draw();
  });
});


defineScreen('deck', 'decks', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, API.decks, ({ decks }) => {
    const deck = decks.find(d => d.id === params.deckId);
    if (!deck) return body.append(header('Колода'), stateBox('Колода не найдена'));
    const p = deck.progress;
    const left = deck.card_count - p.studied;

    body.append(
      header(deck.title, deck.topic + ' · ' + deck.period),
      statTiles([
        [p.studied + '/' + deck.card_count, 'изучено'],
        [p.known, 'помню', 'good'],
        [p.weak, 'слабые', p.weak ? 'bad' : '']
      ]),
      h('div', { class: 'actions' },
        left > 0
          ? h('button', { class: 'btn btn--main', onclick: () => go('study', { deckId: deck.id, mode: 'new' }) },
              p.studied ? 'Учить дальше (' + left + ')' : 'Начать учить')
          : null,
        p.weak > 0
          ? h('button', { class: 'btn btn--warn', onclick: () => go('study', { deckId: deck.id, mode: 'weak' }) },
              'Повторить слабые (' + p.weak + ')')
          : null,
        h('button', { class: 'btn btn--light', onclick: () => go('study', { deckId: deck.id, mode: 'all' }) }, 'Пролистать все карточки'),
        deck.question_count > 0
          ? h('button', { class: 'btn btn--light', onclick: () => go('test', { deckId: deck.id }) }, 'Пройти тест')
          : null,
        h('button', { class: 'btn btn--light', onclick: () => go('testHistory', { deckId: deck.id, title: deck.title }) }, 'Мои попытки тестов'),
        deck.question_count > 0
          ? h('button', { class: 'btn btn--light', onclick: () => go('duelCreate', { deckId: deck.id }) }, 'Вызвать на дуэль')
          : null));
  });
});


const GRADES = [
  ['again', 'Не помню', 'btn--bad'],
  ['hard', 'Сомневаюсь', 'btn--warn'],
  ['good', 'Помню', 'btn--good']
];

defineScreen('study', 'decks', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, () => API.deckCards(params.deckId, params.mode), ({ deck, cards }) => {
    if (!cards.length) {
      const text = params.mode === 'weak' ? 'Слабых карточек нет. Отлично!'
        : params.mode === 'new' ? 'Все карточки этой колоды уже изучены'
        : 'В этой колоде пока нет карточек';
      body.append(header(deck.title), stateBox(text,
        params.mode === 'all' ? null : 'Пролистать все',
        () => go('study', { deckId: deck.id, mode: 'all' }, { replace: true })));
      return;
    }

    let index = 0;
    let answerShown = false;
    const counts = { again: 0, hard: 0, good: 0 };

    const counter = h('span', { class: 'muted' });
    const bar = h('span', { class: 'progress' }, h('span', { class: 'progress__bar' }));
    const side = h('span', { class: 'label' });
    const text = h('span', { class: 'flashcard__text' });
    const hint = h('span', { class: 'muted' });
    const card = h('button', { class: 'flashcard', 'aria-live': 'polite', onclick: showAnswer }, side, text, hint);
    const showBtn = h('button', { class: 'btn btn--main', onclick: showAnswer }, 'Показать ответ');
    const error = h('p', { class: 'error', hidden: true });
    const gradeButtons = GRADES.map(([grade, label, cls]) =>
      h('button', { class: 'btn ' + cls, onclick: () => rate(grade) }, label));
    const gradeRow = h('div', { class: 'actions actions--three', hidden: true }, gradeButtons);

    const modeTitle = { new: 'Новые карточки', weak: 'Повторение слабых', all: 'Все карточки' }[params.mode];
    body.append(header(deck.title, modeTitle), h('div', { class: 'row' }, counter), bar, card, showBtn, gradeRow, error);

    function showCard() {
      const c = cards[index];
      answerShown = false;
      card.classList.remove('is-answer');
      side.textContent = 'Вопрос';
      text.textContent = c.question;
      hint.textContent = 'Нажми, чтобы увидеть ответ';
      counter.textContent = (index + 1) + ' из ' + cards.length;
      bar.firstChild.style.width = (index / cards.length * 100) + '%';
      showBtn.hidden = false;
      gradeRow.hidden = true;
      error.hidden = true;
    }

    function showAnswer() {
      if (answerShown) return;
      answerShown = true;
      card.classList.add('is-answer');
      side.textContent = 'Ответ';
      text.textContent = cards[index].answer;
      hint.textContent = 'Насколько хорошо ты это помнишь?';
      showBtn.hidden = true;
      gradeRow.hidden = false;
    }

    async function rate(grade) {
      if (!answerShown || gradeButtons[0].disabled) return;
      gradeButtons.forEach(b => { b.disabled = true; });
      try {
        
        await API.reviewCard(cards[index].id, grade);
      } catch (err) {
        error.textContent = errorText(err) + '. Оценка не сохранилась, нажми ещё раз';
        error.hidden = false;
        gradeButtons.forEach(b => { b.disabled = false; });
        return;
      }
      gradeButtons.forEach(b => { b.disabled = false; });
      counts[grade]++;
      index++;
      if (index < cards.length) showCard();
      else go('studyResult', { deckId: deck.id, title: deck.title, counts }, { replace: true });
    }

    showCard();
  });
});

defineScreen('studyResult', 'decks', (view, params) => {
  const { again, hard, good } = params.counts;
  const total = again + hard + good;
  const weak = again + hard;
  view.append(
    h('div', { class: 'result' },
      ring(good, total, 'помню'),
      h('h1', { class: 'title' }, 'Готово!'),
      h('p', { class: 'muted' }, weak
        ? 'Слабые карточки сохранены. Повтори их, пока они свежи в памяти'
        : 'Ты помнишь все карточки из этого захода')),
    statTiles([[good, 'помню', 'good'], [hard, 'сомневаюсь', 'warn'], [again, 'не помню', 'bad']]),
    h('div', { class: 'actions' },
      weak ? h('button', { class: 'btn btn--warn', onclick: () => go('study', { deckId: params.deckId, mode: 'weak' }, { replace: true }) }, 'Повторить слабые сейчас') : null,
      h('button', { class: 'btn btn--main', onclick: () => go('deck', { deckId: params.deckId }, { replace: true }) }, 'К колоде')));
});


function ring(part, total, label) {
  const pct = total ? Math.round(part / total * 100) : 0;
  return h('div', { class: 'ring', style: `--pct:${pct}` },
    h('span', { class: 'ring__inner' }, h('b', null, part + '/' + total), h('span', { class: 'muted' }, label)));
}


defineScreen('test', 'decks', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, API.decks, ({ decks }) => {
    const deck = decks.find(d => d.id === params.deckId);
    if (!deck) return body.append(header('Тест'), stateBox('Колода не найдена'));
    body.append(
      header('Тест', deck.title),
      h('div', { class: 'panel' },
        h('b', null, countText(deck.question_count, 'вопрос', 'вопроса', 'вопросов')),
        h('p', { class: 'muted' }, 'В каждом вопросе один правильный ответ из четырёх. После ответа сразу увидишь, верно или нет. Результат сохранится в истории.')),
      actionButton('Начать тест', 'btn btn--main', async () => {
        const data = await API.startTest(deck.id);
        runTest(body, data);
      }));
  });
});

function runTest(body, { attempt, questions }) {
  let index = 0;
  let right = 0, wrong = 0;
  const token = navToken;

  const score = h('div', { class: 'score' });
  const counter = h('span', { class: 'muted' });
  const bar = h('span', { class: 'progress' }, h('span', { class: 'progress__bar' }));
  const question = h('p', { class: 'question' });
  const options = h('div', { class: 'options' });
  const next = h('div', { class: 'actions' });
  const error = h('p', { class: 'error', hidden: true });

  body.replaceChildren(header('Тест', attempt.deck.title), h('div', { class: 'row' }, counter, score), bar, question, options, error, next);

  function drawScore() {
    score.replaceChildren(h('span', { class: 'good' }, '✓ ' + right), h('span', { class: 'bad' }, '✗ ' + wrong));
  }

  function showQuestion() {
    const q = questions[index];
    counter.textContent = 'Вопрос ' + (index + 1) + ' из ' + questions.length;
    bar.firstChild.style.width = (index / questions.length * 100) + '%';
    question.textContent = q.text;
    next.replaceChildren();
    error.hidden = true;
    options.replaceChildren(...q.options.map(o =>
      h('button', { class: 'option', 'data-id': o.id, onclick: () => choose(q, o.id) }, o.text)));
  }

  async function choose(q, optionId) {
    const buttons = [...options.children];
    if (buttons[0].disabled) return;
    buttons.forEach(b => { b.disabled = true; });
    let res;
    try {
      res = await API.answerTest(attempt.id, q.id, optionId);
    } catch (err) {
      if (token !== navToken) return;
      error.textContent = errorText(err);
      error.hidden = false;
      buttons.forEach(b => { b.disabled = false; });
      return;
    }
    if (token !== navToken) return;
    right = res.correct_count;
    wrong = res.wrong_count;
    drawScore();
    buttons.forEach(b => {
      const bid = Number(b.dataset.id);
      if (bid === res.correct_option_id) b.classList.add('is-right');
      else if (bid === optionId) b.classList.add('is-wrong');
    });
    const last = index === questions.length - 1;
    next.append(last
      ? actionButton('Завершить тест', 'btn btn--main', async () => {
          const result = await API.finishTest(attempt.id);
          go('testResult', { result }, { replace: true });
        })
      : h('button', { class: 'btn btn--main', onclick: () => { index++; showQuestion(); } }, 'Следующий вопрос'));
  }

  drawScore();
  showQuestion();
}

defineScreen('testResult', 'decks', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);
  const loader = params.result ? () => Promise.resolve(params.result) : () => API.testResult(params.attemptId);

  load(body, loader, ({ attempt, review }) => {
    const mistakes = review.filter(r => !r.correct);
    body.append(
      header('Результат теста', attempt.deck.title + ' · ' + formatDate(attempt.finished_at)),
      h('div', { class: 'result' },
        ring(attempt.correct, attempt.total, 'верно'),
        h('h1', { class: 'title' }, attempt.percent >= 80 ? 'Отлично!' : attempt.percent >= 50 ? 'Неплохо' : 'Стоит повторить'),
        h('p', { class: 'muted' }, mistakes.length ? 'Разбор ошибок ниже' : 'Ни одной ошибки')),
      statTiles([[attempt.correct, 'верно', 'good'], [attempt.total - attempt.correct, 'ошибки', 'bad'], [attempt.percent + '%', 'результат']]),
      sectionTitle('Разбор ответов'),
      h('div', { class: 'deck-list' }, review.map((r, i) => {
        const chosen = r.options.find(o => o.id === r.chosen_option_id);
        const right = r.options.find(o => o.id === r.correct_option_id);
        return h('div', { class: 'review ' + (r.correct ? 'review--right' : 'review--wrong') },
          h('b', null, (i + 1) + '. ' + r.text),
          h('span', { class: r.correct ? 'good' : 'bad' }, 'Твой ответ: ' + (chosen ? chosen.text : 'нет ответа')),
          r.correct ? null : h('span', null, 'Правильно: ' + right.text));
      })),
      h('div', { class: 'actions' },
        h('button', { class: 'btn btn--main', onclick: () => go('test', { deckId: attempt.deck.id }, { replace: true }) }, 'Пройти ещё раз'),
        h('button', { class: 'btn btn--light', onclick: () => go('testHistory', { deckId: attempt.deck.id, title: attempt.deck.title }) }, 'Все попытки')));
  });
});

defineScreen('testHistory', 'decks', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(header('Попытки тестов', params.title || 'Все колоды'), body);

  load(body, () => API.testHistory(params.deckId), ({ attempts }) => {
    if (!attempts.length) {
      body.append(stateBox('Ты ещё не проходил тесты' + (params.title ? ' по этой колоде' : ''),
        params.deckId ? 'Пройти тест' : null, () => go('test', { deckId: params.deckId }, { replace: true })));
      return;
    }
    body.append(h('div', { class: 'deck-list' }, attempts.map(a =>
      h('button', { class: 'deck', onclick: () => go('testResult', { attemptId: a.id }) },
        h('span', { class: 'deck__badge' }, a.percent + '%'),
        h('span', { class: 'deck__body' },
          h('span', { class: 'deck__name' }, a.deck.title),
          h('span', { class: 'muted small' }, formatDate(a.finished_at) + ' · ' + a.correct + ' из ' + a.total)),
        icon('next')))));
  });
});


defineScreen('myClasses', 'myClasses', view => {
  const body = h('div', { class: 'stack' });
  view.append(header('Класс'), body);

  load(body, API.myClasses, ({ classes }) => {
    if (!classes.length) {
      body.append(joinForm());
      return;
    }
    body.append(
      h('div', { class: 'deck-list' }, classes.map(c =>
        h('button', { class: 'banner', onclick: () => go('classInfo', { classId: c.id }) },
          h('span', { class: 'banner__icon' }, icon('class')),
          h('span', null, h('b', null, c.name),
            h('span', { class: 'muted' }, c.teacher_name + ' · ' + countText(c.member_count, 'участник', 'участника', 'участников')))))),
      h('button', { class: 'btn btn--light', onclick: () => go('classJoin') }, 'Вступить в другой класс'));
  });
});

defineScreen('classJoin', 'myClasses', view => {
  view.append(header('Вступить в класс'), joinForm());
});

function joinForm() {
  const input = h('input', { class: 'input input--code', id: 'classCode', maxlength: 8, autocomplete: 'off', placeholder: 'Код' });
  return h('form', { class: 'panel', onsubmit: e => e.preventDefault() },
    h('label', { for: 'classCode' }, h('b', null, 'Код класса')),
    h('p', { class: 'muted' }, 'Его даёт учитель. Например: HIST7B'),
    DEMO_MODE ? h('p', { class: 'hint' }, 'Демо-код: HIST7B') : null,
    input,
    actionButton('Вступить', 'btn btn--main', async () => {
      const code = input.value.trim().toUpperCase();
      if (!/^[A-Z0-9]{4,8}$/.test(code)) {
        throw new ApiError(400, 'invalid_code', 'Код — это 4–8 латинских букв или цифр');
      }
      const { class: cls } = await API.joinClass(code);
      toast('Ты вступил в класс ' + cls.name);
      go('classInfo', { classId: cls.id }, { replace: true });
    }));
}

defineScreen('classInfo', 'myClasses', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, () => API.classInfo(params.classId), ({ class: cls, members, my_progress: p }) => {
    body.append(
      header('Класс ' + cls.name, 'Учитель: ' + cls.teacher_name),
      sectionTitle('Мой прогресс'),
      statTiles([
        [p.cards_known + '/' + p.cards_studied, 'помню / изучено'],
        [p.tests_taken, 'тестов'],
        [percentText(p.avg_test_percent), 'средний балл']
      ]),
      sectionTitle('Участники · ' + members.length),
      h('div', { class: 'panel members' }, members.map(m =>
        h('div', { class: 'member' + (m.is_me ? ' is-me' : '') },
          h('span', { class: 'avatar avatar--small' }, m.name.charAt(0)),
          h('span', null, m.name + (m.is_me ? ' (ты)' : ''))))));
  });
});


defineScreen('profile', 'profile', view => {
  if (me.role === 'teacher') return renderTeacherProfile(view);

  const name = [me.user.first_name, me.user.last_name].filter(Boolean).join(' ');
  const body = h('div', { class: 'stack' });
  view.append(
    header('Профиль'),
    h('div', { class: 'panel profile' },
      h('span', { class: 'avatar' }, name.charAt(0) || '?'),
      h('span', null, h('b', { class: 'hero__title' }, name), h('br'), h('span', { class: 'muted' }, 'Ученик'))),
    body,
    settingsPanel());

  load(body, API.myStats, s => {
    body.append(
      statTiles([
        [s.cards_known + '/' + s.cards_total, 'карточек помню'],
        [s.tests_taken, 'тестов пройдено'],
        [percentText(s.avg_test_percent), 'средний балл'],
        [s.duels_won + '/' + s.duels_played, 'побед в дуэлях']
      ]),
      sectionTitle('Прогресс по колодам'),
      h('div', { class: 'panel' }, s.decks.map(d =>
        h('div', { class: 'progress-row' },
          h('div', { class: 'row' }, h('span', null, d.title), h('span', { class: 'muted small' }, d.known + '/' + d.total)),
          progressBar(d.known, d.total)))),
      sectionTitle('Последние тесты'),
      s.recent_tests.length
        ? h('div', { class: 'deck-list' }, s.recent_tests.map(t =>
            h('button', { class: 'deck', onclick: () => go('testResult', { attemptId: t.id }) },
              h('span', { class: 'deck__badge' }, t.percent + '%'),
              h('span', { class: 'deck__body' }, h('span', { class: 'deck__name' }, t.deck.title),
                h('span', { class: 'muted small' }, formatDate(t.finished_at))),
              icon('next'))))
        : stateBox('Пока нет пройденных тестов'));
  });
});
