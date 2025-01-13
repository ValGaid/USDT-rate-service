-- +goose Up
-- +goose StatementBegin
CREATE TABLE usdt (
                      id SERIAL PRIMARY KEY,
                      timestamp TIMESTAMP NOT NULL,
                      ask_price TEXT NOT NULL,
                      bid_price TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE usdt;
-- +goose StatementEnd