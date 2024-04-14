CREATE TABLE IF NOT EXISTS banners (
    id BIGSERIAL PRIMARY KEY,
    content JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS banners_history (
    banner_id BIGINT NOT NULL,
    content JSONB NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (banner_id, updated_at),
    FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS banner_mappings (
    banner_id BIGINT NOT NULL,
    feature_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (banner_id, feature_id, tag_id),
    UNIQUE (feature_id, tag_id),
    UNIQUE (banner_id, tag_id)
);

-- Триггер для удаления всех записей кроме последних трех
CREATE OR REPLACE FUNCTION keep_last_3_versions()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM banners_history
    WHERE banner_id = NEW.banner_id
    AND updated_at NOT IN (
        SELECT updated_at
        FROM (
            SELECT updated_at
            FROM banners_history
            WHERE banner_id = NEW.banner_id
            ORDER BY updated_at DESC
            LIMIT 3
        ) AS subquery
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER keep_last_3_versions_trigger
AFTER INSERT ON banners_history
FOR EACH ROW EXECUTE FUNCTION keep_last_3_versions();