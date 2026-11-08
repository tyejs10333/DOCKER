
BEGIN;

DO $seed$
DECLARE
    v_deck_id INTEGER;
    v_question_id INTEGER;
BEGIN
    SELECT id INTO v_deck_id
    FROM decks
    WHERE title = 'DEV: История России'
    ORDER BY id
    LIMIT 1;

    IF v_deck_id IS NULL THEN
        INSERT INTO decks (title, topic, period, position)
        VALUES ('DEV: История России', 'История России', '988–1380', 1)
        RETURNING id INTO v_deck_id;

        INSERT INTO cards (deck_id, question, answer, position)
        VALUES
            (v_deck_id, 'В каком году произошло Крещение Руси?', 'В 988 году.', 1),
            (v_deck_id, 'Какой князь связан с Крещением Руси?', 'Владимир Святославич.', 2),
            (v_deck_id, 'Когда состоялась Куликовская битва?', 'В 1380 году.', 3),
            (v_deck_id, 'Кто возглавлял русское войско на Куликовом поле?', 'Князь Дмитрий Донской.', 4),
            (v_deck_id, 'Как назывался свод законов Древней Руси?', 'Русская Правда.', 5);

        INSERT INTO questions (deck_id, text, position)
        VALUES (v_deck_id, 'В каком году произошло Крещение Руси?', 1)
        RETURNING id INTO v_question_id;
        INSERT INTO question_options (question_id, text, is_correct, position)
        VALUES
            (v_question_id, '988', TRUE, 1),
            (v_question_id, '862', FALSE, 2),
            (v_question_id, '1147', FALSE, 3),
            (v_question_id, '1380', FALSE, 4);

        INSERT INTO questions (deck_id, text, position)
        VALUES (v_deck_id, 'Какой князь связан с Крещением Руси?', 2)
        RETURNING id INTO v_question_id;
        INSERT INTO question_options (question_id, text, is_correct, position)
        VALUES
            (v_question_id, 'Владимир Святославич', TRUE, 1),
            (v_question_id, 'Иван IV', FALSE, 2),
            (v_question_id, 'Пётр I', FALSE, 3),
            (v_question_id, 'Александр II', FALSE, 4);

        INSERT INTO questions (deck_id, text, position)
        VALUES (v_deck_id, 'Когда состоялась Куликовская битва?', 3)
        RETURNING id INTO v_question_id;
        INSERT INTO question_options (question_id, text, is_correct, position)
        VALUES
            (v_question_id, '1380', TRUE, 1),
            (v_question_id, '1240', FALSE, 2),
            (v_question_id, '1480', FALSE, 3),
            (v_question_id, '1612', FALSE, 4);

        INSERT INTO questions (deck_id, text, position)
        VALUES (v_deck_id, 'Кто возглавлял русское войско на Куликовом поле?', 4)
        RETURNING id INTO v_question_id;
        INSERT INTO question_options (question_id, text, is_correct, position)
        VALUES
            (v_question_id, 'Дмитрий Донской', TRUE, 1),
            (v_question_id, 'Александр Невский', FALSE, 2),
            (v_question_id, 'Иван Грозный', FALSE, 3),
            (v_question_id, 'Ярослав Мудрый', FALSE, 4);

        INSERT INTO questions (deck_id, text, position)
        VALUES (v_deck_id, 'Как назывался свод законов Древней Руси?', 5)
        RETURNING id INTO v_question_id;
        INSERT INTO question_options (question_id, text, is_correct, position)
        VALUES
            (v_question_id, 'Русская Правда', TRUE, 1),
            (v_question_id, 'Судебник 1497 года', FALSE, 2),
            (v_question_id, 'Соборное уложение 1649 года', FALSE, 3),
            (v_question_id, 'Табель о рангах', FALSE, 4);
    END IF;
END
$seed$;

COMMIT;
