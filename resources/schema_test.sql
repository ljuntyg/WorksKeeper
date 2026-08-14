DROP TABLE IF EXISTS collections CASCADE;
DROP TABLE IF EXISTS works CASCADE;
DROP TABLE IF EXISTS series CASCADE;
DROP TABLE IF EXISTS listings CASCADE;
DROP TABLE IF EXISTS canvases CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS contents CASCADE;
DROP TABLE IF EXISTS texts CASCADE;
DROP TABLE IF EXISTS media CASCADE;
DROP TABLE IF EXISTS sources CASCADE;
DROP TABLE IF EXISTS captions CASCADE;
DROP TABLE IF EXISTS fileservers CASCADE;
DROP TABLE IF EXISTS filenodes CASCADE;
DROP TABLE IF EXISTS files CASCADE;
DROP TABLE IF EXISTS filenames CASCADE;

DROP TYPE IF EXISTS l_type CASCADE;
DROP TYPE IF EXISTS c_type CASCADE;

CREATE TYPE l_type AS ENUM ('series', 'work');
CREATE TYPE c_type AS ENUM ('group', 'text', 'media');

CREATE TABLE collections (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    root_series_id BIGINT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE works (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    listing_id BIGINT UNIQUE NOT NULL,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE series (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    listing_id BIGINT UNIQUE,
    title VARCHAR(4096) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE listings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    parent_series_id BIGINT NOT NULL,
    position NUMERIC NOT NULL,
    listing_type L_TYPE NOT NULL,
    UNIQUE (parent_series_id, position)
);

CREATE TABLE canvases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    work_id BIGINT UNIQUE NOT NULL,
    last_edit TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    canvas_id BIGINT UNIQUE,
    content_id BIGINT UNIQUE,
    swap_direction BOOLEAN DEFAULT FALSE NOT NULL,
    CHECK ((canvas_id IS NULL) <> (content_id IS NULL))
);

CREATE TABLE contents (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    parent_group_id BIGINT NOT NULL,
    position NUMERIC NOT NULL,
    content_type C_TYPE NOT NULL,
    UNIQUE (parent_group_id, position)
);

CREATE TABLE texts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    content_id BIGINT UNIQUE NOT NULL,
    content VARCHAR(4194304)
);

CREATE TABLE media (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    content_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE sources (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    media_id BIGINT NOT NULL,
    filename_id BIGINT NOT NULL
);

CREATE TABLE captions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    media_id BIGINT UNIQUE NOT NULL,
    content VARCHAR(4194304)
);

CREATE TABLE fileservers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    scheme VARCHAR(8) NOT NULL,
    host VARCHAR(4096) NOT NULL,
    port INTEGER NOT NULL,
    disk_path VARCHAR(4096) NOT NULL,
    url_path VARCHAR(4096) NOT NULL,
    UNIQUE (host, port)
);

CREATE TABLE files (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    size BIGINT NOT NULL,
    hash VARCHAR(128) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    UNIQUE (hash, size)
);

CREATE TABLE filenodes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    file_id BIGINT NOT NULL,
    fileserver_id BIGINT NOT NULL,
    path VARCHAR(4096) NOT NULL,
    UNIQUE (fileserver_id, path)
);

CREATE TABLE filenames (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
    file_id BIGINT NOT NULL,
    name VARCHAR(4096) NOT NULL,
    UNIQUE (file_id, name)
);

CREATE INDEX sources_media_id_idx ON sources (media_id);
CREATE INDEX sources_filename_id_idx ON sources (filename_id);
CREATE INDEX filenodes_file_id_idx ON filenodes (file_id);
CREATE INDEX filenodes_fileserver_id_idx ON filenodes (fileserver_id);
CREATE INDEX filenames_file_id_idx ON filenames (file_id);

ALTER TABLE collections
ADD CONSTRAINT collections_root_series_fk
FOREIGN KEY (root_series_id) REFERENCES series(id) ON DELETE RESTRICT;

ALTER TABLE listings
ADD CONSTRAINT listings_parent_series_fk
FOREIGN KEY (parent_series_id) REFERENCES series(id) ON DELETE CASCADE;

ALTER TABLE series
ADD CONSTRAINT series_listing_fk
FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE;

ALTER TABLE works
ADD CONSTRAINT works_listing_fk
FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE;

ALTER TABLE canvases
ADD CONSTRAINT canvases_work_fk
FOREIGN KEY (work_id) REFERENCES works(id) ON DELETE CASCADE;

ALTER TABLE groups
ADD CONSTRAINT groups_canvas_fk
FOREIGN KEY (canvas_id) REFERENCES canvases(id) ON DELETE CASCADE;

ALTER TABLE groups
ADD CONSTRAINT groups_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE contents
ADD CONSTRAINT contents_parent_group_fk
FOREIGN KEY (parent_group_id) REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE texts
ADD CONSTRAINT texts_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE media
ADD CONSTRAINT media_content_fk
FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE;

ALTER TABLE sources
ADD CONSTRAINT sources_media_fk
FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE;

ALTER TABLE captions
ADD CONSTRAINT captions_media_fk
FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE;

ALTER TABLE sources
ADD CONSTRAINT sources_filename_fk
FOREIGN KEY (filename_id) REFERENCES filenames(id) ON DELETE RESTRICT;

ALTER TABLE filenodes
ADD CONSTRAINT filenodes_file_fk
FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE;

ALTER TABLE filenodes
ADD CONSTRAINT filenodes_fileserver_fk
FOREIGN KEY (fileserver_id) REFERENCES fileservers(id) ON DELETE CASCADE;

ALTER TABLE filenames
ADD CONSTRAINT filenames_file_fk
FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE;

-- ===== mock data =====

-- The local file server, matching the static file handlers registered in main.
INSERT INTO fileservers (scheme, host, port, disk_path, url_path)
VALUES ('http', 'localhost', 8080, './resources/media', '/media/'); -- fileserver.id 1

-- Placeholder digests; real ones are computed from the bytes on upload.
INSERT INTO files (size, hash, mime_type)
VALUES (1048576, '1111111111111111111111111111111111111111111111111111111111111111', 'video/mp4'); -- file.id 1

INSERT INTO files (size, hash, mime_type)
VALUES (65536, '2222222222222222222222222222222222222222222222222222222222222222', 'image/png'); -- file.id 2

INSERT INTO files (size, hash, mime_type)
VALUES (262144, '3333333333333333333333333333333333333333333333333333333333333333', 'audio/ogg'); -- file.id 3

INSERT INTO filenames (file_id, name) VALUES (1, 'introsong.mp4'); -- filename.id 1
INSERT INTO filenames (file_id, name) VALUES (2, 'plots.png');     -- filename.id 2
INSERT INTO filenames (file_id, name) VALUES (3, 'test.ogg');      -- filename.id 3

INSERT INTO filenodes (file_id, fileserver_id, path) VALUES (1, 1, 'videos/introsong.mp4'); -- filenode.id 1
INSERT INTO filenodes (file_id, fileserver_id, path) VALUES (2, 1, 'images/plots.png');     -- filenode.id 2
INSERT INTO filenodes (file_id, fileserver_id, path) VALUES (3, 1, 'sounds/test.ogg');      -- filenode.id 3

-- Root series for the only collection
INSERT INTO series (title, created_at) VALUES ('Series #1 (root)', NOW());          -- series.id 1, listing_id NULL

-- Nested series
INSERT INTO series (title, created_at) VALUES ('Series #2 (nested)', NOW());        -- series.id 2

INSERT INTO listings (parent_series_id, position, listing_type)
VALUES (1, 1.0, 'series'); -- listing.id 1: Series #2 under Series #1

UPDATE series SET listing_id = 1 WHERE id = 2;

-- Another series in the same collection
INSERT INTO series (title, created_at) VALUES ('Series #3', NOW());                 -- series.id 3

INSERT INTO listings (parent_series_id, position, listing_type)
VALUES (1, 2.0, 'series'); -- listing.id 2: Series #3 under Series #1

UPDATE series SET listing_id = 2 WHERE id = 3;

-- Works
INSERT INTO listings (parent_series_id, position, listing_type)
VALUES (2, 1.0, 'work');   -- listing.id 3: Work #1 under Series #2

INSERT INTO listings (parent_series_id, position, listing_type)
VALUES (1, 3.0, 'work');   -- listing.id 4: Work #2 under Series #1

INSERT INTO works (listing_id, title)
VALUES (3, 'Work #1 (nested)'); -- work.id 1

INSERT INTO works (listing_id, title)
VALUES (4, 'Work #2'); -- work.id 2

INSERT INTO canvases (work_id) VALUES (1); -- canvas.id 1

INSERT INTO canvases (work_id) VALUES (2); -- canvas.id 2

INSERT INTO groups (canvas_id) VALUES (1); -- group.id 1: root group for Work #1

INSERT INTO groups (canvas_id) VALUES (2); -- group.id 2: root group for Work #2

-- Single collection
INSERT INTO collections (root_series_id) VALUES (1);

-- Canvas #1 contents: a nested group + a top-level text

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (1, 1.0, 'group'); -- content.id 1

INSERT INTO groups (content_id) VALUES (1); -- group.id 3

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (1, 2.0, 'text'); -- content.id 2

INSERT INTO texts (content_id, content)
VALUES (2, 'Text #2');

-- Inside nested group 3: a text, then a video with source + caption

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (3, 1.0, 'text'); -- content.id 3

INSERT INTO texts (content_id, content)
VALUES (3, 'Text #1');

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (3, 2.0, 'media'); -- content.id 4

INSERT INTO media (content_id)
VALUES (4); -- media.id 1

INSERT INTO captions (media_id, content)
VALUES (1, 'Caption #1');

INSERT INTO sources (media_id, filename_id)
VALUES (1, 1);

-- Canvas #2 contents: a single top-level text

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (2, 1.0, 'text'); -- content.id 5

INSERT INTO texts (content_id, content)
VALUES (5, 'Text #3');

-- Series #3 contents: a Work exercising the image and sound Media branches

INSERT INTO listings (parent_series_id, position, listing_type)
VALUES (3, 1.0, 'work');   -- listing.id 5: Work #3 under Series #3

INSERT INTO works (listing_id, title)
VALUES (5, 'Work #3 (in Series #3)'); -- work.id 3

INSERT INTO canvases (work_id) VALUES (3); -- canvas.id 3

INSERT INTO groups (canvas_id) VALUES (3); -- group.id 4: root group for Work #3

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (4, 1.0, 'media'); -- content.id 6

INSERT INTO media (content_id)
VALUES (6); -- media.id 2

INSERT INTO sources (media_id, filename_id)
VALUES (2, 2);

INSERT INTO captions (media_id, content)
VALUES (2, 'Caption #2 (on an image)');

INSERT INTO contents (parent_group_id, position, content_type)
VALUES (4, 2.0, 'media'); -- content.id 7

INSERT INTO media (content_id)
VALUES (7); -- media.id 3

INSERT INTO sources (media_id, filename_id)
VALUES (3, 3);