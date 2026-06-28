DROP TABLE IF EXISTS works CASCADE;
DROP TABLE IF EXISTS series CASCADE;
DROP TABLE IF EXISTS canvases CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS texts CASCADE;
DROP TABLE IF EXISTS media CASCADE;
DROP TABLE IF EXISTS sources CASCADE;
DROP TABLE IF EXISTS captions CASCADE;

DROP TYPE IF EXISTS m_type CASCADE;
CREATE TYPE m_type AS ENUM ('sound', 'video', 'image');

CREATE TABLE works (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    series_id BIGINT,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE series (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    parent_id BIGINT,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE canvases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    work_id BIGINT NOT NULL,
    last_edit TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    parent_id BIGINT,
    canvas_id BIGINT NOT NULL,
    idx INTEGER NOT NULL,
    swap_direction BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE texts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    group_id BIGINT NOT NULL,
    idx INTEGER NOT NULL,
    content VARCHAR(4194304)
);

CREATE TABLE media (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    group_id BIGINT NOT NULL,
    idx INTEGER NOT NULL,
    media_type M_TYPE NOT NULL
);

CREATE TABLE sources (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    media_id BIGINT NOT NULL,
    link VARCHAR(4096) NOT NULL
);

CREATE TABLE captions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    media_id BIGINT NOT NULL,
    content VARCHAR(4194304)
);

ALTER TABLE works
ADD CONSTRAINT works_series_fk
FOREIGN KEY (series_id) REFERENCES series(id);

ALTER TABLE series
ADD CONSTRAINT series_parent_fk
FOREIGN KEY (parent_id) REFERENCES series(id);

ALTER TABLE canvases
ADD CONSTRAINT canvases_work_fk
FOREIGN KEY (work_id) REFERENCES works(id);

ALTER TABLE groups
ADD CONSTRAINT groups_parent_fk
FOREIGN KEY (parent_id) REFERENCES groups(id);

ALTER TABLE groups
ADD CONSTRAINT groups_canvas_fk
FOREIGN KEY (canvas_id) REFERENCES canvases(id);

ALTER TABLE texts
ADD CONSTRAINT texts_group_fk
FOREIGN KEY (group_id) REFERENCES groups(id);

ALTER TABLE media
ADD CONSTRAINT media_group_fk
FOREIGN KEY (group_id) REFERENCES groups(id);

ALTER TABLE sources
ADD CONSTRAINT sources_media_fk
FOREIGN KEY (media_id) REFERENCES media(id);

ALTER TABLE captions
ADD CONSTRAINT captions_media_fk
FOREIGN KEY (media_id) REFERENCES media(id);

INSERT INTO series (title, created_at) VALUES('Series #1', NOW());
INSERT INTO series (parent_id, title, created_at) VALUES(1, 'Series #2 (nested)', NOW());
INSERT INTO series (title, created_at) VALUES('Series #3', NOW());

INSERT INTO works (series_id, title, created_at) VALUES(2, 'Work #1 (nested)', NOW());
INSERT INTO works (title, created_at) VALUES('Work #2', NOW());

INSERT INTO canvases (work_id, last_edit) VALUES(1, NOW());

INSERT INTO groups (canvas_id, idx) VALUES(1, 0);
INSERT INTO groups (parent_id, canvas_id, idx) VALUES(1, 1, 0);

INSERT INTO texts (group_id, idx, content) VALUES(2, 0, 'Text #1');
INSERT INTO texts (group_id, idx, content) VALUES(1, 1, 'Text #2');

INSERT INTO media (group_id, idx, media_type) VALUES(2, 1, 'video');

INSERT INTO sources (media_id, link) VALUES(1, 'http://localhost:8080/media/video/introsong.mp4');

INSERT INTO captions (media_id, content) VALUES(1, 'Caption #1');