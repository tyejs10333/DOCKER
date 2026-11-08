

defineScreen('teacherHome', 'teacherHome', view => {
  const body = h('div', { class: 'stack' });
  view.append(header('Здравствуйте, ' + userName() + '!', 'Ваши классы'), body);

  load(body, API.teacherClasses, ({ classes }) => {
    const students = classes.reduce((s, c) => s + c.member_count, 0);
    body.append(statTiles([
      [classes.length, plural(classes.length, 'класс', 'класса', 'классов')],
      [students, plural(students, 'ученик', 'ученика', 'учеников')]
    ]));

    if (!classes.length) {
      body.append(stateBox('У вас пока нет классов. Создайте класс и отправьте ученикам код.'));
    } else {
      body.append(h('div', { class: 'deck-list' }, classes.map(c =>
        h('button', { class: 'deck', onclick: () => go('teacherClass', { classId: c.id }) },
          h('span', { class: 'deck__badge' }, icon('class')),
          h('span', { class: 'deck__body' },
            h('span', { class: 'deck__name' }, c.name),
            h('span', { class: 'muted small' },
              countText(c.member_count, 'ученик', 'ученика', 'учеников') + ' · средний балл ' + percentText(c.avg_test_percent))),
          h('span', { class: 'code-chip' }, c.code)))));
    }
    body.append(h('button', { class: 'btn btn--main', onclick: () => go('createClass') }, icon('plus'), 'Создать класс'));
  });
});

defineScreen('createClass', 'teacherHome', view => {
  const input = h('input', { class: 'input', id: 'className', maxlength: 40, autocomplete: 'off', placeholder: 'Например: 7 «Б»' });
  view.append(
    header('Новый класс'),
    h('form', { class: 'panel', onsubmit: e => e.preventDefault() },
      h('label', { for: 'className' }, h('b', null, 'Название')),
      input,
      actionButton('Создать', 'btn btn--main', async () => {
        const name = input.value.trim();
        if (!name) throw new ApiError(400, 'invalid_name', 'Введите название класса');
        const { class: cls } = await API.createClass(name);
        go('teacherClass', { classId: cls.id, justCreated: true }, { replace: true });
      })));
  input.focus();
});

defineScreen('teacherClass', 'teacherHome', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, () => Promise.all([API.teacherClass(params.classId), API.classStats(params.classId)]), ([{ class: cls, students }, stats]) => {
    body.append(
      header('Класс ' + cls.name, countText(cls.member_count, 'ученик', 'ученика', 'учеников')),
      h('div', { class: 'panel' },
        h('b', null, params.justCreated ? 'Класс создан! Отправьте ученикам этот код' : 'Код для вступления'),
        codeBox(cls.code, 'Ученик открывает «Стопку» → «Класс» → вводит код')));

    if (!students.length) {
      body.append(stateBox('В классе пока нет учеников. Как только кто-то вступит по коду, он появится здесь.'));
      return;
    }

    body.append(
      sectionTitle('Статистика класса'),
      statTiles([
        [percentText(stats.avg_test_percent), 'средний балл'],
        [stats.cards_studied_total, 'карточек изучено']
      ]));

    if (stats.decks.length) {
      body.append(sectionTitle('Темы'),
        h('div', { class: 'panel' }, stats.decks.map(d =>
          h('div', { class: 'progress-row' },
            h('div', { class: 'row' }, h('span', null, d.title),
              h('span', { class: 'muted small' }, d.students_started + ' начали · помнят ' + d.avg_known_percent + '%')),
            progressBar(d.avg_known_percent, 100)))));
    }

    if (stats.hardest_cards.length) {
      body.append(sectionTitle('Трудные карточки'),
        h('div', { class: 'panel' }, stats.hardest_cards.map(c =>
          h('div', { class: 'hard-card' },
            h('span', null, c.question),
            h('span', { class: 'muted small' }, c.deck_title + ' · ошибаются в ' + c.wrong_rate + '% ответов')))));
    }

    body.append(sectionTitle('Ученики'),
      h('div', { class: 'deck-list' }, students.map(s =>
        h('button', { class: 'deck', onclick: () => go('teacherStudent', { classId: cls.id, studentId: s.id }) },
          h('span', { class: 'avatar avatar--small' }, s.name.charAt(0)),
          h('span', { class: 'deck__body' },
            h('span', { class: 'deck__name' }, s.name),
            h('span', { class: 'muted small' },
              s.last_active_at
                ? 'карточек ' + s.cards_known + '/' + s.cards_studied + ' · тесты ' + percentText(s.avg_test_percent) + ' · ' + formatDate(s.last_active_at)
                : 'ещё не занимался')),
          icon('next')))));
  });
});

defineScreen('teacherStudent', 'teacherHome', (view, params) => {
  const body = h('div', { class: 'stack' });
  view.append(body);

  load(body, () => API.studentProgress(params.classId, params.studentId), data => {
    const { student, stats: s } = data;
    body.append(
      header(student.name, ''),
      statTiles([
        [s.cards_known + '/' + s.cards_studied, 'помнит / изучил'],
        [s.tests_taken, 'тестов'],
        [percentText(s.avg_test_percent), 'средний балл'],
        [s.duels_won + '/' + s.duels_played, 'побед в дуэлях']
      ]));

    body.append(sectionTitle('Темы'), data.decks.length
      ? h('div', { class: 'panel' }, data.decks.map(d =>
          h('div', { class: 'progress-row' },
            h('div', { class: 'row' }, h('span', null, d.title), h('span', { class: 'muted small' }, d.known + '/' + d.total)),
            progressBar(d.known, d.total))))
      : stateBox('Пока ничего не изучал'));

    body.append(sectionTitle('Тесты'), data.recent_tests.length
      ? h('div', { class: 'panel' }, data.recent_tests.map(t =>
          h('div', { class: 'row' }, h('span', null, t.deck.title), h('span', { class: 'muted small' }, t.percent + '% · ' + formatDate(t.finished_at)))))
      : stateBox('Тестов пока нет'));

    if (data.weak_cards.length) {
      body.append(sectionTitle('Даётся труднее всего'),
        h('div', { class: 'panel' }, data.weak_cards.map(c =>
          h('div', { class: 'hard-card' }, h('span', null, c.question), h('span', { class: 'muted small' }, c.deck_title)))));
    }

    body.append(confirmButton('Удалить из класса', 'Точно удалить? Нажмите ещё раз', async () => {
      await API.removeStudent(params.classId, params.studentId);
      toast(student.name + ' удалён из класса');
      back();
    }));
  });
});

function renderTeacherProfile(view) {
  const name = [me.user.first_name, me.user.last_name].filter(Boolean).join(' ');
  view.append(
    header('Профиль'),
    h('div', { class: 'panel profile' },
      h('span', { class: 'avatar' }, name.charAt(0) || '?'),
      h('span', null, h('b', { class: 'hero__title' }, name), h('br'), h('span', { class: 'muted' }, 'Учитель'))),
    settingsPanel());
}
