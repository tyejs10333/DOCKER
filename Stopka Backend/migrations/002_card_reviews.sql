CREATE TABLE IF NOT EXISTS card_reviews (
 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 card_id INT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
 grade TEXT NOT NULL CHECK (grade IN ('again','hard','good')),
 reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,card_id)
);
CREATE INDEX IF NOT EXISTS card_reviews_card_idx ON card_reviews(card_id);
