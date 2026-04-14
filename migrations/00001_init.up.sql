-- milk_tea_go 初始 schema：用户、登录 refresh、品类、奶茶单品、评论、统计表 + 触发器
-- PostgreSQL 14+

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ========== 用户 ==========
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64)  NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    nickname        VARCHAR(64)  NOT NULL DEFAULT '',
    avatar_url      VARCHAR(512) NOT NULL DEFAULT '',
    phone           VARCHAR(32)  NOT NULL DEFAULT '',
    bio             VARCHAR(512) NOT NULL DEFAULT '',
    gender          SMALLINT     NOT NULL DEFAULT 0,
    birthday        DATE,
    status          SMALLINT     NOT NULL DEFAULT 1,
    email_verified  BOOLEAN      NOT NULL DEFAULT FALSE,
    last_login_at   TIMESTAMPTZ,
    last_login_ip   INET,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_status ON users (status);
CREATE INDEX idx_users_created_at ON users (created_at DESC);

COMMENT ON TABLE users IS '用户主表';
COMMENT ON COLUMN users.password_hash IS 'bcrypt 密码摘要';
COMMENT ON COLUMN users.status IS '1 正常 0 禁用 2 待验证';

-- ========== Refresh Token（JWT + DB 轮换/吊销） ==========
CREATE TABLE refresh_tokens (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jti             UUID         NOT NULL UNIQUE,
    token_hash      CHAR(64)     NOT NULL,
    family_id       UUID         NOT NULL,
    parent_jti      UUID,
    expires_at      TIMESTAMPTZ  NOT NULL,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  VARCHAR(64)  NOT NULL DEFAULT '',
    ip              INET,
    user_agent      TEXT,
    device_id       VARCHAR(128) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens (expires_at);
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family_id);

COMMENT ON TABLE refresh_tokens IS 'Refresh Token 仅存 SHA-256 摘要';

-- ========== 奶茶品类 ==========
CREATE TABLE categories (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    slug            VARCHAR(128) NOT NULL UNIQUE,
    description     TEXT         NOT NULL DEFAULT '',
    icon_url        VARCHAR(512) NOT NULL DEFAULT '',
    cover_url       VARCHAR(512) NOT NULL DEFAULT '',
    sort_order      INT          NOT NULL DEFAULT 0,
    is_hot          BOOLEAN      NOT NULL DEFAULT FALSE,
    status          SMALLINT     NOT NULL DEFAULT 1,
    extra           JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_sort ON categories (sort_order DESC, id);

COMMENT ON TABLE categories IS '奶茶品类';

-- ========== 奶茶单品 ==========
CREATE TABLE teas (
    id              BIGSERIAL PRIMARY KEY,
    category_id     BIGINT       NOT NULL REFERENCES categories(id),
    name            VARCHAR(128) NOT NULL,
    shop_name       VARCHAR(128) NOT NULL DEFAULT '',
    subtitle        VARCHAR(256) NOT NULL DEFAULT '',
    description     TEXT         NOT NULL DEFAULT '',
    image_url       VARCHAR(512) NOT NULL DEFAULT '',
    gallery         JSONB        NOT NULL DEFAULT '[]',
    tags            TEXT[]       NOT NULL DEFAULT '{}',
    price_min_cent  INT          NOT NULL DEFAULT 0,
    price_max_cent  INT          NOT NULL DEFAULT 0,
    currency        CHAR(3)      NOT NULL DEFAULT 'CNY',
    sku_code        VARCHAR(64)  NOT NULL DEFAULT '',
    calories_kcal   INT,
    sugar_level     VARCHAR(32)  NOT NULL DEFAULT '',
    caffeine_mg     INT,
    is_seasonal     BOOLEAN      NOT NULL DEFAULT FALSE,
    is_recommended  BOOLEAN      NOT NULL DEFAULT FALSE,
    status          SMALLINT     NOT NULL DEFAULT 1,
    view_count      BIGINT       NOT NULL DEFAULT 0,
    extra           JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teas_category ON teas (category_id);
CREATE INDEX idx_teas_status ON teas (status);
CREATE INDEX idx_teas_name ON teas (name);

COMMENT ON TABLE teas IS '奶茶单品';

-- ========== 评论 ==========
CREATE TABLE reviews (
    id              BIGSERIAL PRIMARY KEY,
    tea_id          BIGINT       NOT NULL REFERENCES teas(id) ON DELETE CASCADE,
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stars           SMALLINT     NOT NULL CHECK (stars >= 1 AND stars <= 5),
    title           VARCHAR(128) NOT NULL DEFAULT '',
    comment         TEXT         NOT NULL DEFAULT '',
    images          JSONB        NOT NULL DEFAULT '[]',
    helpful_count   INT          NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 1,
    ip              INET,
    extra           JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (tea_id, user_id)
);

CREATE INDEX idx_reviews_tea ON reviews (tea_id, created_at DESC);
CREATE INDEX idx_reviews_user ON reviews (user_id, created_at DESC);

COMMENT ON TABLE reviews IS '点评；每用户每单品一条，可更新';
COMMENT ON COLUMN reviews.status IS '1 可见 0 隐藏 2 待审';

-- ========== 单品统计 ==========
CREATE TABLE tea_stats (
    tea_id          BIGINT PRIMARY KEY REFERENCES teas(id) ON DELETE CASCADE,
    review_count    BIGINT       NOT NULL DEFAULT 0,
    rating_sum      BIGINT       NOT NULL DEFAULT 0,
    avg_rating      NUMERIC(4,2) NOT NULL DEFAULT 0,
    last_review_at  TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tea_stats_rank ON tea_stats (avg_rating DESC, review_count DESC);

-- 重算某 tea_id 的统计（仅统计 status=1 的可见评论）
CREATE OR REPLACE FUNCTION refresh_tea_stats(p_tea_id BIGINT)
RETURNS VOID AS $$
DECLARE
    cnt BIGINT;
    s BIGINT;
    avg NUMERIC(4,2);
    last_at TIMESTAMPTZ;
BEGIN
    SELECT
        COUNT(*)::BIGINT,
        COALESCE(SUM(stars), 0)::BIGINT,
        ROUND(COALESCE(AVG(stars), 0)::NUMERIC, 2)::NUMERIC(4,2),
        MAX(created_at)
    INTO cnt, s, avg, last_at
    FROM reviews
    WHERE tea_id = p_tea_id AND status = 1;

    IF cnt IS NULL OR cnt = 0 THEN
        DELETE FROM tea_stats WHERE tea_id = p_tea_id;
        RETURN;
    END IF;

    INSERT INTO tea_stats (tea_id, review_count, rating_sum, avg_rating, last_review_at, updated_at)
    VALUES (p_tea_id, cnt, s, avg, last_at, NOW())
    ON CONFLICT (tea_id) DO UPDATE SET
        review_count = EXCLUDED.review_count,
        rating_sum = EXCLUDED.rating_sum,
        avg_rating = EXCLUDED.avg_rating,
        last_review_at = EXCLUDED.last_review_at,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_reviews_refresh_stats()
RETURNS TRIGGER AS $$
DECLARE
    tid BIGINT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        tid := OLD.tea_id;
    ELSE
        tid := NEW.tea_id;
    END IF;
    PERFORM refresh_tea_stats(tid);
    IF TG_OP = 'UPDATE' AND OLD.tea_id IS DISTINCT FROM NEW.tea_id THEN
        PERFORM refresh_tea_stats(OLD.tea_id);
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reviews_refresh_stats
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW EXECUTE FUNCTION trg_reviews_refresh_stats();
