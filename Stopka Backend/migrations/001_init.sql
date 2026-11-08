


CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    max_user_id BIGINT NOT NULL UNIQUE,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    username TEXT,
    photo_url TEXT,
    role TEXT CHECK (role IN ('student', 'teacher')),
    teacher_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ,
    last_active_at TIMESTAMPTZ
);


CREATE TABLE decks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    topic TEXT NOT NULL,
    period TEXT NOT NULL DEFAULT '',
    position INT NOT NULL DEFAULT 0
);


CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    deck_id INT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0
);


CREATE TABLE questions (
    id SERIAL PRIMARY KEY,
    deck_id INT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0
);

CREATE TABLE question_options (
    id SERIAL PRIMARY KEY,
    question_id INT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    position INT NOT NULL DEFAULT 0
);


CREATE UNIQUE INDEX one_correct_option
ON question_options (question_id)
WHERE is_correct;


CREATE TABLE card_progress (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id INT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    level SMALLINT NOT NULL DEFAULT 0 CHECK (level BETWEEN 0 AND 5),
    last_grade TEXT NOT NULL CHECK (last_grade IN ('again', 'hard', 'good')),
    reviews_count INT NOT NULL DEFAULT 0,
    wrong_count INT NOT NULL DEFAULT 0,
    reviewed_at TIMESTAMPTZ NOT NULL,
    due_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, card_id)
);


CREATE TABLE test_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    deck_id INT NOT NULL REFERENCES decks(id),
    status TEXT NOT NULL DEFAULT 'in_progress'
        CHECK (status IN ('in_progress', 'finished')),
    correct INT NOT NULL DEFAULT 0,
    total INT NOT NULL,
    percent INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX ON test_attempts (user_id, finished_at DESC);


CREATE TABLE test_attempt_questions (
    attempt_id BIGINT NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
    question_id INT NOT NULL REFERENCES questions(id),
    position INT NOT NULL,
    PRIMARY KEY (attempt_id, question_id)
);


CREATE TABLE test_answers (
    attempt_id BIGINT NOT NULL REFERENCES test_attempts(id) ON DELETE CASCADE,
    question_id INT NOT NULL REFERENCES questions(id),
    option_id INT NOT NULL REFERENCES question_options(id),
    is_correct BOOLEAN NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (attempt_id, question_id)
);


CREATE TABLE classes (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 40),
    code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE class_members (
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (class_id, user_id)
);

CREATE INDEX ON class_members (user_id);


CREATE TABLE duels (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    deck_id INT NOT NULL REFERENCES decks(id),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'waiting'
        CHECK (status IN ('waiting', 'active', 'finished', 'cancelled')),
    winner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);


CREATE TABLE duel_questions (
    duel_id BIGINT NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    question_id INT NOT NULL REFERENCES questions(id),
    position INT NOT NULL,
    PRIMARY KEY (duel_id, question_id)
);


CREATE TABLE duel_players (
    duel_id BIGINT NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score INT NOT NULL DEFAULT 0,
    answered INT NOT NULL DEFAULT 0,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (duel_id, user_id)
);


CREATE TABLE duel_answers (
    duel_id BIGINT NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id INT NOT NULL REFERENCES questions(id),
    option_id INT NOT NULL REFERENCES question_options(id),
    is_correct BOOLEAN NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (duel_id, user_id, question_id)
);

