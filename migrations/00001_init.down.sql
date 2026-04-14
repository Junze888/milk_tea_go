DROP TRIGGER IF EXISTS reviews_refresh_stats ON reviews;
DROP FUNCTION IF EXISTS trg_reviews_refresh_stats();
DROP FUNCTION IF EXISTS refresh_tea_stats(BIGINT);

DROP TABLE IF EXISTS tea_stats;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS teas;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
